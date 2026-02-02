'use client';

import { useCallback } from 'react';
import { useProjectStore } from '@/stores/useProjectStore';
import { ImageDropzone } from '@/components/upload/ImageDropzone';
import { ImageList } from '@/components/upload/ImageList';
import { ChatHistory } from '@/components/chat/ChatHistory';
import { UnifiedInput } from '@/components/upload/UnifiedInput';
import { useGeneration } from '@/lib/hooks/useGeneration';
import { useChat } from '@/lib/hooks/useChat';
import type { ImageFile } from '@/types';

function generateId() {
  return Math.random().toString(36).substring(2, 11);
}

export function InteractionPanel() {
  const images = useProjectStore((state) => state.images);
  const addImages = useProjectStore((state) => state.addImages);
  const chatMessages = useProjectStore((state) => state.chatMessages);
  const status = useProjectStore((state) => state.status);
  const generatedCode = useProjectStore((state) => state.generatedCode);
  const sessionId = useProjectStore((state) => state.sessionId);

  const { generate } = useGeneration();
  const { sendMessage } = useChat();

  const addAssistantMessage = useProjectStore((state) => state.addAssistantMessage);
  const startNewProject = useProjectStore((state) => state.startNewProject);

  const isProcessing = status === 'uploading' || status === 'generating';
  const hasCode = !!generatedCode?.code;

  // 开始新原型
  const handleNewProject = useCallback(() => {
    if (window.confirm('确定要开始新原型吗？当前的原型和对话记录将被清空。')) {
      startNewProject();
    }
  }, [startNewProject]);

  // 处理发送
  const handleSend = useCallback(
    async (text: string, pastedImages: File[]) => {
      // 如果有粘贴的图片，添加到图片列表
      if (pastedImages.length > 0) {
        const newImages: ImageFile[] = pastedImages.map((file, i) => ({
          id: generateId(),
          file,
          preview: URL.createObjectURL(file),
          order: images.length + i,
        }));
        addImages(newImages);
      }

      // 判断是生成还是对话修改
      if (!hasCode) {
        // 首次生成：需要有图片
        if (images.length > 0 || pastedImages.length > 0) {
          await generate(text);
        } else {
          // 没有图片时在对话区域提示用户
          addAssistantMessage('⚠️ 请先上传 UI 设计稿图片，系统将根据设计稿自动推断交互逻辑并生成可交互原型。\n\n您可以：\n1. 点击上方"拖拽或点击上传设计稿"区域上传图片\n2. 直接粘贴图片到输入框中');
        }
      } else {
        // 对话修改：需要有文字
        if (text && sessionId) {
          await sendMessage(sessionId, text);
        }
      }
    },
    [images.length, hasCode, sessionId, addImages, generate, sendMessage, addAssistantMessage]
  );

  // 确定按钮文字
  const getButtonText = () => {
    if (isProcessing) {
      return status === 'uploading' ? '上传中...' : '生成中...';
    }
    return hasCode ? '发送' : '生成原型';
  };

  // 确定 placeholder
  const getPlaceholder = () => {
    if (hasCode) {
      return '描述修改需求，如：把按钮改成蓝色...';
    }
    return '描述需求（可选），支持粘贴图片...';
  };

  return (
    <div className="flex h-full flex-col bg-slate-50">
      {/* 标题 */}
      <div className="border-b border-slate-200 bg-white p-4">
        <div className="flex items-center justify-between">
          <h2 className="font-semibold text-slate-900">设计稿上传</h2>
          {hasCode && !isProcessing && (
            <button
              onClick={handleNewProject}
              className="flex items-center gap-1 rounded-md bg-blue-50 px-2 py-1 text-xs font-medium text-blue-600 hover:bg-blue-100 transition-colors"
            >
              <svg className="h-3.5 w-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M12 4v16m8-8H4" />
              </svg>
              新建原型
            </button>
          )}
        </div>
        <p className="mt-1 text-xs text-slate-500">
          {hasCode
            ? '输入描述修改当前原型，或点击"新建原型"重新开始'
            : '上传 UI 设计稿序列，系统将自动推断交互逻辑'}
        </p>
      </div>

      {/* 上传区域（仅在未生成原型时显示） */}
      {!hasCode && (
        <div className="border-b border-slate-200 bg-white p-3">
          <ImageDropzone disabled={isProcessing} />
        </div>
      )}

      {/* 已上传图片（仅在未生成原型时显示） */}
      {!hasCode && images.length > 0 && (
        <div className="border-b border-slate-200 bg-white p-3">
          <div className="mb-2">
            <span className="text-xs font-medium text-slate-600">
              已上传 {images.length} 张图片
              {isProcessing && <span className="ml-2 text-slate-400">（生成中，无法修改）</span>}
            </span>
          </div>
          <ImageList layout="horizontal" disabled={isProcessing} />
        </div>
      )}

      {/* 对话历史 */}
      <ChatHistory messages={chatMessages} isGenerating={status === 'generating'} />

      {/* 统一输入框 */}
      <UnifiedInput
        onSend={handleSend}
        disabled={isProcessing}
        placeholder={getPlaceholder()}
        buttonText={getButtonText()}
        hasUploadedImages={images.length > 0 && !hasCode}
      />
    </div>
  );
}
