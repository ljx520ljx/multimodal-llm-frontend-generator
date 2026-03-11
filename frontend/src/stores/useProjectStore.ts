'use client';

import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { GenerationStatus, ImageFile, GeneratedCode, ChatMessage } from '@/types';

interface ProjectState {
  // 视图状态
  codeExpanded: boolean;

  // 会话状态
  sessionId: string | null;

  // 图片状态
  images: ImageFile[];
  imageIds: string[];

  // 生成状态
  generatedCode: GeneratedCode | null;
  generatedCss: string;
  thinkingContent: string;
  status: GenerationStatus;
  errorMessage: string | null;

  // 编辑器状态
  activeFile: 'App.tsx' | 'styles.css';

  // 生成模式
  generationMode: 'fast' | 'quality';

  // 分享状态
  shareUrl: string | null;

  // 对话历史
  chatMessages: ChatMessage[];

  // 视图操作
  toggleCodeExpanded: () => void;

  // 会话操作
  setSessionId: (id: string) => void;

  // 图片操作
  addImage: (image: ImageFile) => void;
  addImages: (images: ImageFile[]) => void;
  removeImage: (id: string) => void;
  reorderImages: (images: ImageFile[]) => void;
  clearImages: () => void;

  // 生成操作
  setStatus: (status: GenerationStatus) => void;
  setThinking: (content: string) => void;
  appendThinking: (content: string) => void;
  setGeneratedCode: (code: GeneratedCode) => void;
  updateCode: (code: string) => void;
  updateCss: (css: string) => void;
  setError: (message: string) => void;
  reset: () => void;
  startNewProject: () => void;  // 完全重置，开始新原型

  // 图片ID操作
  setImageIds: (ids: string[]) => void;

  // 编辑器操作
  setActiveFile: (file: 'App.tsx' | 'styles.css') => void;

  // 生成模式操作
  setGenerationMode: (mode: 'fast' | 'quality') => void;

  // 分享操作
  setShareUrl: (url: string | null) => void;

  // 对话历史操作
  addUserMessage: (text: string, images?: string[]) => void;
  addAssistantMessage: (content: string) => void;
  updateLastAssistantMessage: (content: string) => void;
  appendToLastAssistantMessage: (content: string) => void;
  clearChatMessages: () => void;
}

const initialState = {
  codeExpanded: false,
  sessionId: null as string | null,
  images: [] as ImageFile[],
  imageIds: [] as string[],
  generatedCode: null,
  generatedCss: '',
  thinkingContent: '',
  status: 'idle' as GenerationStatus,
  errorMessage: null,
  activeFile: 'App.tsx' as const,
  generationMode: 'fast' as const,
  shareUrl: null as string | null,
  chatMessages: [] as ChatMessage[],
};

function generateMessageId() {
  return `msg_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`;
}

export const useProjectStore = create<ProjectState>()(
  persist(
    (set) => ({
      ...initialState,

      // 视图操作
      toggleCodeExpanded: () =>
        set((state) => ({ codeExpanded: !state.codeExpanded })),

      // 会话操作
      setSessionId: (id) => set({ sessionId: id }),

      // 图片操作
      addImage: (image) =>
        set((state) => ({
          images: [...state.images, image],
        })),
      addImages: (images) =>
        set((state) => ({
          images: [...state.images, ...images],
        })),
      removeImage: (id) =>
        set((state) => {
          const removed = state.images.find((img) => img.id === id);
          if (removed?.preview) {
            URL.revokeObjectURL(removed.preview);
          }
          return { images: state.images.filter((img) => img.id !== id) };
        }),
      reorderImages: (images) => set({ images }),
      clearImages: () =>
        set((state) => {
          state.images.forEach((img) => {
            if (img.preview) URL.revokeObjectURL(img.preview);
          });
          return { images: [] };
        }),

      // 生成操作
      setStatus: (status) => set({ status }),
      setThinking: (content) => set({ thinkingContent: content }),
      appendThinking: (content) =>
        set((state) => ({
          thinkingContent: state.thinkingContent + content,
        })),
      setGeneratedCode: (code) =>
        set({ generatedCode: code }),
      updateCode: (code) =>
        set(() => ({
          generatedCode: { code, timestamp: Date.now() },
        })),
      updateCss: (css) => set({ generatedCss: css }),
      setError: (message) =>
        set({ status: 'error', errorMessage: message }),
      reset: () =>
        set({
          generatedCode: null,
          generatedCss: '',
          thinkingContent: '',
          status: 'idle',
          errorMessage: null,
          activeFile: 'App.tsx',
          chatMessages: [],
        }),
      startNewProject: () =>
        set((state) => {
          state.images.forEach((img) => {
            if (img.preview) URL.revokeObjectURL(img.preview);
          });
          return { ...initialState };
        }),

      // 图片ID操作
      setImageIds: (ids) => set({ imageIds: ids }),

      // 编辑器操作
      setActiveFile: (file) => set({ activeFile: file }),

      // 生成模式操作
      setGenerationMode: (mode) => set({ generationMode: mode }),

      // 分享操作
      setShareUrl: (url) => set({ shareUrl: url }),

      // 对话历史操作
      addUserMessage: (text, images) =>
        set((state) => ({
          chatMessages: [
            ...state.chatMessages,
            {
              id: generateMessageId(),
              role: 'user',
              content: text,
              images,
              timestamp: Date.now(),
            },
          ],
        })),
      addAssistantMessage: (content) =>
        set((state) => ({
          chatMessages: [
            ...state.chatMessages,
            {
              id: generateMessageId(),
              role: 'assistant',
              content,
              timestamp: Date.now(),
            },
          ],
        })),
      updateLastAssistantMessage: (content) =>
        set((state) => {
          const messages = [...state.chatMessages];
          const lastIndex = messages.length - 1;
          if (lastIndex >= 0 && messages[lastIndex].role === 'assistant') {
            messages[lastIndex] = { ...messages[lastIndex], content };
          }
          return { chatMessages: messages };
        }),
      appendToLastAssistantMessage: (content) =>
        set((state) => {
          const messages = [...state.chatMessages];
          const lastIndex = messages.length - 1;
          if (lastIndex >= 0 && messages[lastIndex].role === 'assistant') {
            messages[lastIndex] = {
              ...messages[lastIndex],
              content: messages[lastIndex].content + content,
            };
          }
          return { chatMessages: messages };
        }),
      clearChatMessages: () => set({ chatMessages: [] }),
    }),
    {
      name: 'project-store',
      partialize: (state) => ({
        sessionId: state.sessionId,
        generationMode: state.generationMode,
        codeExpanded: state.codeExpanded,
      }),
    }
  )
);
