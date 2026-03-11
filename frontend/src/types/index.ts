// 生成状态
export type GenerationStatus =
  | 'idle'
  | 'uploading'
  | 'generating'
  | 'completed'
  | 'error';

// 图片文件
export interface ImageFile {
  id: string;
  file: File;
  preview: string;
  order: number;
}

// 生成结果
export interface GeneratedCode {
  code: string;
  thinking?: string;
  timestamp: number;
}

// API 响应
export interface UploadResponse {
  success: boolean;
  imageIds: string[];
  message?: string;
}

export interface GenerateResponse {
  code: string;
  thinking?: string;
}

// SSE 事件
export interface SSEEvent {
  type: 'thinking' | 'code' | 'done' | 'error' | 'agent_start' | 'agent_result';
  content: string;
  agent?: string;
}

// 对话消息
export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  images?: string[]; // 图片预览 URL
  timestamp: number;
}

// 用户
export interface User {
  id: string;
  email: string;
  display_name: string;
  avatar_url?: string;
  github_login?: string;
  created_at: string;
}

// 认证 Token
export interface TokenPair {
  access_token: string;
  expires_at: number;
}

// 项目
export interface Project {
  id: string;
  user_id: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}
