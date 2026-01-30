const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

interface RequestOptions extends RequestInit {
  timeout?: number;
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
  options: RequestOptions = {}
): Promise<T> {
  const { timeout = 30000, ...fetchOptions } = options;

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

    return response.json();
  } finally {
    clearTimeout(timeoutId);
  }
}

export const api = {
  // 健康检查
  health: () => request<{ status: string }>('/health'),

  // 上传图片
  upload: async (files: File[]) => {
    const formData = new FormData();
    files.forEach((file) => {
      formData.append('images[]', file);
    });

    return request<{ session_id: string; images: Array<{ id: string; filename: string; order: number }> }>('/api/upload', {
      method: 'POST',
      body: formData,
      timeout: 60000,
    });
  },

  // 生成代码 (返回 Response 用于 SSE)
  generate: async (sessionId: string, imageIds: string[], framework: string = 'react'): Promise<Response> => {
    const response = await fetch(`${API_BASE}/api/generate`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        session_id: sessionId,
        image_ids: imageIds,
        framework: framework,
      }),
    });

    if (!response.ok) {
      throw new APIError('Generation failed', response.status);
    }

    return response;
  },

  // 对话修正
  chat: async (sessionId: string, message: string): Promise<Response> => {
    const response = await fetch(`${API_BASE}/api/chat`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ session_id: sessionId, message }),
    });

    if (!response.ok) {
      throw new APIError('Chat failed', response.status);
    }

    return response;
  },
};

export { APIError };
