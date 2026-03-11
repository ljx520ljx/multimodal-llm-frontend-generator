import { z } from 'zod';

// 健康检查响应
export const HealthResponseSchema = z.object({
  status: z.string(),
});

// 上传响应
export const UploadResponseSchema = z.object({
  session_id: z.string(),
  images: z.array(
    z.object({
      id: z.string(),
      filename: z.string(),
      order: z.number(),
    })
  ),
});

export type UploadResponse = z.infer<typeof UploadResponseSchema>;

// SSE 事件
export const SSEEventSchema = z.object({
  type: z.enum(['thinking', 'code', 'done', 'error', 'agent_start', 'agent_result']),
  content: z.string(),
  agent: z.string().optional(),
});

export type SSEEvent = z.infer<typeof SSEEventSchema>;

// 分享响应
export const ShareResponseSchema = z.object({
  short_code: z.string(),
  url: z.string(),
});

export type ShareResponse = z.infer<typeof ShareResponseSchema>;

// 用户响应
export const UserSchema = z.object({
  id: z.string(),
  email: z.string(),
  display_name: z.string(),
  avatar_url: z.string().optional(),
  github_login: z.string().optional(),
  created_at: z.string(),
});

export const TokenPairSchema = z.object({
  access_token: z.string(),
  expires_at: z.number(),
});

export const AuthResponseSchema = z.object({
  user: UserSchema,
  token: TokenPairSchema,
});

export type AuthResponse = z.infer<typeof AuthResponseSchema>;

export const MeResponseSchema = z.object({
  user: UserSchema,
});

// 项目响应
export const ProjectSchema = z.object({
  id: z.string(),
  user_id: z.string(),
  name: z.string(),
  description: z.string(),
  created_at: z.string(),
  updated_at: z.string(),
});

export const ProjectListResponseSchema = z.object({
  projects: z.array(ProjectSchema),
});

export const ProjectResponseSchema = z.object({
  project: ProjectSchema,
});

export type ProjectResponse = z.infer<typeof ProjectResponseSchema>;
