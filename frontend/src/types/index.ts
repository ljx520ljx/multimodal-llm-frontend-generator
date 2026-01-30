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
  type: 'thinking' | 'code' | 'done' | 'error';
  content: string;
}

// 对话消息
export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  images?: string[]; // 图片预览 URL
  timestamp: number;
}
