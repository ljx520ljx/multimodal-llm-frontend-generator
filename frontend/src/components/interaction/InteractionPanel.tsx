'use client';

import { useCallback, useEffect } from 'react';
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

  const generationMode = useProjectStore((state) => state.generationMode);
  const setGenerationMode = useProjectStore((state) => state.setGenerationMode);

  const { generate, regenerate } = useGeneration();
  const { sendMessage } = useChat();

  const addAssistantMessage = useProjectStore((state) => state.addAssistantMessage);
  const startNewProject = useProjectStore((state) => state.startNewProject);

  const isProcessing = status === 'uploading' || status === 'generating';
  const hasCode = !!generatedCode?.code;

  // 全局剪贴板粘贴监听（允许不聚焦 textarea 时粘贴图片）
  useEffect(() => {
    const handleGlobalPaste = (e: globalThis.ClipboardEvent) => {
      // 如果焦点在 textarea/input 上，交给组件自己处理
      const target = e.target as HTMLElement;
      if (target.tagName === 'TEXTAREA' || target.tagName === 'INPUT') return;

      const items = e.clipboardData?.items;
      if (!items) return;

      const imageFiles: File[] = [];
      for (const item of Array.from(items)) {
        if (item.type.startsWith('image/')) {
          const file = item.getAsFile();
          if (file) imageFiles.push(file);
        }
      }

      if (imageFiles.length > 0) {
        e.preventDefault();
        const newImages: ImageFile[] = imageFiles.map((file, i) => ({
          id: generateId(),
          file,
          preview: URL.createObjectURL(file),
          order: images.length + i,
        }));
        addImages(newImages);
      }
    };

    document.addEventListener('paste', handleGlobalPaste);
    return () => document.removeEventListener('paste', handleGlobalPaste);
  }, [images.length, addImages]);

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
        // 首次生成
        if (images.length > 0 || pastedImages.length > 0) {
          // 有图片：正常生成
          await generate(text);
        } else if (text && generationMode === 'quality') {
          // 无图片但有文字描述且为精细模式：文字描述生成 UI
          await generate(text);
        } else if (text) {
          // 无图片有文字但为快速模式：提示切换模式
          addAssistantMessage('💡 纯文字描述生成需要使用"精细"模式。请切换到精细模式后再试，或上传 UI 设计稿图片。');
        } else {
          // 没有图片也没有文字
          addAssistantMessage('⚠️ 请上传 UI 设计稿图片，或输入文字描述（精细模式）来生成可交互原型。\n\n您可以：\n1. 点击上方"拖拽或点击上传设计稿"区域上传图片\n2. 直接粘贴图片到输入框中\n3. 切换到"精细"模式，输入文字描述生成 UI');
        }
      } else {
        // 对话修改：需要有文字
        if (text && sessionId) {
          await sendMessage(sessionId, text);
        }
      }
    },
    [images.length, hasCode, sessionId, generationMode, addImages, generate, sendMessage, addAssistantMessage]
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

      {/* 生成模式切换（仅在未生成原型时显示） */}
      {!hasCode && (
        <div className="flex items-center gap-2 border-b border-slate-200 bg-white px-4 py-2">
          <span className="text-xs text-slate-500">模式</span>
          <div className="flex rounded-md bg-slate-100 p-0.5">
            <button
              onClick={() => setGenerationMode('fast')}
              disabled={isProcessing}
              className={`rounded px-2.5 py-1 text-xs font-medium transition-colors ${
                generationMode === 'fast'
                  ? 'bg-white text-slate-900 shadow-sm'
                  : 'text-slate-500 hover:text-slate-700'
              } disabled:opacity-50`}
            >
              快速
            </button>
            <button
              onClick={() => setGenerationMode('quality')}
              disabled={isProcessing}
              className={`rounded px-2.5 py-1 text-xs font-medium transition-colors ${
                generationMode === 'quality'
                  ? 'bg-white text-slate-900 shadow-sm'
                  : 'text-slate-500 hover:text-slate-700'
              } disabled:opacity-50`}
            >
              精细
            </button>
          </div>
          <span className="text-[10px] text-slate-400">
            {generationMode === 'fast' ? '单步生成，速度快' : '多步分析，质量高'}
          </span>
        </div>
      )}

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


      {/* 错误恢复按钮 */}
      {status === 'error' && (
        <div className="flex items-center justify-center border-t border-slate-200 bg-red-50 px-4 py-3">
          <button
            onClick={regenerate}
            className="flex items-center gap-1.5 rounded-md bg-red-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-red-700 transition-colors"
          >
            <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M1 4v6h6M23 20v-6h-6" />
              <path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 0 1 3.51 15" />
            </svg>
            重新生成
          </button>
        </div>
      )}

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
