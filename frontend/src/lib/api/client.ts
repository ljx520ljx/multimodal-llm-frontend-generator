import type { ZodSchema } from 'zod';
import {
  HealthResponseSchema,
  UploadResponseSchema,
  ShareResponseSchema,
  AuthResponseSchema,
  MeResponseSchema,
  ProjectListResponseSchema,
  ProjectResponseSchema,
} from './schemas';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

function getAuthToken(): string | null {
  if (typeof window === 'undefined') return null;
  try {
    const raw = localStorage.getItem('auth-store');
    if (!raw) return null;
    const data = JSON.parse(raw);
    return data?.state?.token?.access_token || null;
  } catch {
    return null;
  }
}

function authHeaders(): Record<string, string> {
  const token = getAuthToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

interface RequestOptions<T = unknown> extends RequestInit {
  timeout?: number;
  schema?: ZodSchema<T>;
}

class APIError extends Error {
  constructor(
    message: string,
    public status: number,
    public data?: unknown
  ) {
    super(message);
    this.name = 'APIError';
  }
}

async function request<T>(
  endpoint: string,
  options: RequestOptions<T> = {}
): Promise<T> {
  const { timeout = 30000, schema, ...fetchOptions } = options;

  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeout);

  try {
    const response = await fetch(`${API_BASE}${endpoint}`, {
      ...fetchOptions,
      signal: controller.signal,
    });

    if (!response.ok) {
      const data = await response.json().catch(() => null);
      throw new APIError(
        data?.message || `Request failed with status ${response.status}`,
        response.status,
        data
      );
    }

    const data = await response.json();

    if (schema) {
      return schema.parse(data);
    }

    return data as T;
  } finally {
    clearTimeout(timeoutId);
  }
}

export const api = {
  // 健康检查
  health: () => request('/health', { schema: HealthResponseSchema }),

  // 上传图片（支持真实上传进度回调）
  upload: (files: File[], onProgress?: (percent: number) => void) => {
    return new Promise<{ session_id: string; images: Array<{ id: string; filename: string; order: number }> }>((resolve, reject) => {
      const formData = new FormData();
      files.forEach((file) => {
        formData.append('images[]', file);
      });

      const xhr = new XMLHttpRequest();
      xhr.open('POST', `${API_BASE}/api/upload`);
      xhr.timeout = 60000;
      const token = getAuthToken();
      if (token) {
        xhr.setRequestHeader('Authorization', `Bearer ${token}`);
      }

      xhr.upload.onprogress = (event) => {
        if (event.lengthComputable && onProgress) {
          const percent = Math.round((event.loaded / event.total) * 100);
          onProgress(percent);
        }
      };

      xhr.onload = () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            const raw = JSON.parse(xhr.responseText);
            resolve(UploadResponseSchema.parse(raw));
          } catch (e) {
            reject(new APIError(
              e instanceof Error ? e.message : 'Invalid response',
              xhr.status
            ));
          }
        } else {
          let data;
          try { data = JSON.parse(xhr.responseText); } catch { /* ignore */ }
          reject(new APIError(
            data?.message || `Upload failed with status ${xhr.status}`,
            xhr.status,
            data
          ));
        }
      };

      xhr.onerror = () => reject(new APIError('Network error', 0));
      xhr.ontimeout = () => reject(new APIError('Upload timeout', 0));

      xhr.send(formData);
    });
  },

  // 生成代码 (返回 Response 用于 SSE)
  // mode: 'fast' 使用 Go 单步生成, 'quality' 使用 Python 多 Agent Pipeline
  generate: async (sessionId: string, imageIds: string[], framework: string = 'react', mode: 'fast' | 'quality' = 'fast', description?: string, signal?: AbortSignal): Promise<Response> => {
    const endpoint = mode === 'quality' ? '/api/agent/generate' : '/api/generate';
    const body: Record<string, unknown> = {
      session_id: sessionId,
      image_ids: imageIds,
      framework: framework,
    };
    if (description) {
      body.description = description;
    }
    const response = await fetch(`${API_BASE}${endpoint}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...authHeaders(),
      },
      body: JSON.stringify(body),
      signal,
    });

    if (!response.ok) {
      throw new APIError('Generation failed', response.status);
    }

    return response;
  },

  // 分享原型
  share: (sessionId: string) =>
    request('/api/share', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ session_id: sessionId }),
      schema: ShareResponseSchema,
    }),

  // 对话修正
  chat: async (sessionId: string, message: string, signal?: AbortSignal): Promise<Response> => {
    const response = await fetch(`${API_BASE}/api/chat`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...authHeaders(),
      },
      body: JSON.stringify({ session_id: sessionId, message }),
      signal,
    });

    if (!response.ok) {
      throw new APIError('Chat failed', response.status);
    }

    return response;
  },

  // 认证
  auth: {
    register: (email: string, password: string, displayName: string) =>
      request('/api/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password, display_name: displayName }),
        schema: AuthResponseSchema,
      }),

    login: (email: string, password: string) =>
      request('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
        schema: AuthResponseSchema,
      }),

    refresh: (token: string) =>
      request('/api/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token }),
        schema: AuthResponseSchema,
      }),

    me: (token?: string) =>
      request('/api/auth/me', {
        headers: token
          ? { Authorization: `Bearer ${token}` }
          : authHeaders(),
        schema: MeResponseSchema,
      }),

    getGitHubAuthURL: () => `${API_BASE}/api/auth/github`,
  },

  // 项目管理
  projects: {
    list: () =>
      request('/api/projects', {
        headers: authHeaders(),
        schema: ProjectListResponseSchema,
      }),

    create: (name: string, description: string = '') =>
      request('/api/projects', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...authHeaders() },
        body: JSON.stringify({ name, description }),
        schema: ProjectResponseSchema,
      }),

    get: (id: string) =>
      request(`/api/projects/${id}`, {
        headers: authHeaders(),
        schema: ProjectResponseSchema,
      }),

    update: (id: string, name: string, description: string = '') =>
      request(`/api/projects/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', ...authHeaders() },
        body: JSON.stringify({ name, description }),
        schema: ProjectResponseSchema,
      }),

    delete: (id: string) =>
      request(`/api/projects/${id}`, {
        method: 'DELETE',
        headers: authHeaders(),
      }),
  },
};

export { APIError };
