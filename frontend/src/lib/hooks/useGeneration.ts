'use client';

import { useCallback, useRef, useState } from 'react';
import { useProjectStore } from '@/stores/useProjectStore';
import { api, readSSEStream } from '@/lib/api';

// 清理代码中的 markdown 代码块标记
function cleanCodeContent(code: string): string {
  let cleaned = code;

  // 0. 移除 thinking 标签
  cleaned = cleaned.replace(/<thinking>[\s\S]*?<\/thinking>/gi, '');
  cleaned = cleaned.replace(/<\/?thinking>/gi, '');

  // 1. 提取代码块内容（如果有），支持多种语言标记
  const codeBlockPattern = /```(?:html|jsx|tsx|javascript|typescript)?\s*\n?([\s\S]*?)```/g;
  const matches: string[] = [];
  let match;
  while ((match = codeBlockPattern.exec(cleaned)) !== null) {
    matches.push(match[1]);
  }

  if (matches.length > 0) {
    // 优先选择包含完整 HTML 结构的代码块
    let bestMatch: string | null = null;
    for (const blockContent of matches) {
      if (blockContent.includes('<!DOCTYPE') || blockContent.includes('<html')) {
        bestMatch = blockContent;
        break;
      }
      if (!bestMatch && (blockContent.includes('<div') || blockContent.includes('<body'))) {
        bestMatch = blockContent;
      }
    }
    cleaned = bestMatch || matches[matches.length - 1];
  } else {
    // 2. 如果没有代码块，移除可能的代码块标记
    cleaned = cleaned.replace(/^```[\w]*\s*\n?/gm, '');
    cleaned = cleaned.replace(/\n?```\s*$/gm, '');
  }

  // 3. 清理首尾空白
  cleaned = cleaned.trim();

  // 4. 确保代码以 HTML 内容开始
  const doctypeIndex = cleaned.indexOf('<!DOCTYPE');
  const htmlIndex = cleaned.indexOf('<html');
  const headIndex = cleaned.indexOf('<head');
  const bodyIndex = cleaned.indexOf('<body');
  const divIndex = cleaned.indexOf('<div');

  const indices = [doctypeIndex, htmlIndex, headIndex, bodyIndex, divIndex].filter(i => i >= 0);
  if (indices.length > 0) {
    const startIndex = Math.min(...indices);
    if (startIndex > 0 && startIndex < cleaned.length) {
      cleaned = cleaned.substring(startIndex);
    }
  }

  return cleaned;
}

export function useGeneration() {
  const images = useProjectStore((state) => state.images);
  const setStatus = useProjectStore((state) => state.setStatus);
  const setThinking = useProjectStore((state) => state.setThinking);
  const appendThinking = useProjectStore((state) => state.appendThinking);
  const setGeneratedCode = useProjectStore((state) => state.setGeneratedCode);
  const setImageIds = useProjectStore((state) => state.setImageIds);
  const setSessionId = useProjectStore((state) => state.setSessionId);
  const setError = useProjectStore((state) => state.setError);

  // 对话历史操作
  const addUserMessage = useProjectStore((state) => state.addUserMessage);
  const addAssistantMessage = useProjectStore((state) => state.addAssistantMessage);
  const appendToLastAssistantMessage = useProjectStore((state) => state.appendToLastAssistantMessage);

  const [uploadProgress, setUploadProgress] = useState(0);
  const abortControllerRef = useRef<AbortController | null>(null);

  const generate = useCallback(async (promptText?: string) => {
    if (images.length === 0) {
      setError('请先上传设计稿');
      return;
    }

    try {
      // 重置生成状态（保留对话历史）
      setUploadProgress(0);
      setStatus('uploading');
      setThinking('');

      // 添加用户消息到对话历史
      const imagePreviews = images.map((img) => img.preview);
      addUserMessage(promptText || '生成交互原型', imagePreviews);

      // 添加 AI 消息占位
      addAssistantMessage('');

      // 创建可中断的控制器
      abortControllerRef.current = new AbortController();

      // 1. 上传图片
      const files = images.map((img) => img.file);

      // 模拟上传进度
      const progressInterval = setInterval(() => {
        setUploadProgress((prev) => Math.min(prev + 10, 90));
      }, 200);

      const uploadResult = await api.upload(files);

      clearInterval(progressInterval);
      setUploadProgress(100);

      // 保存 sessionId 和 imageIds
      setSessionId(uploadResult.session_id);
      const imageIds = uploadResult.images.map((img) => img.id);
      setImageIds(imageIds);

      // 2. 请求生成代码
      setStatus('generating');
      // 更新对话框状态提示
      appendToLastAssistantMessage('🔄 正在生成代码...\n\n');

      const response = await api.generate(uploadResult.session_id, imageIds, 'react');

      // 3. 读取 SSE 流
      let codeBuffer = '';

      await readSSEStream(response, {
        onThinking: (content) => {
          appendThinking(content);
          // 更新对话历史中的 AI 消息
          appendToLastAssistantMessage(content);
        },
        onCode: (content) => {
          codeBuffer += content;
          // 流式更新代码，让用户实时看到生成结果（清理 markdown 标记）
          setGeneratedCode({
            code: cleanCodeContent(codeBuffer),
            timestamp: Date.now(),
          });
        },
        onDone: () => {
          if (codeBuffer) {
            // 最终清理确保没有残留的 markdown 标记
            const finalCode = cleanCodeContent(codeBuffer);
            setGeneratedCode({
              code: finalCode,
              timestamp: Date.now(),
            });
            // 添加完成提示（总是显示，让用户知道生成完成）
            appendToLastAssistantMessage('\n\n✅ 原型生成完成！请在右侧预览区体验交互效果');
          } else {
            // 没有代码生成
            appendToLastAssistantMessage('\n\n⚠️ 未能生成代码，请重试或调整设计稿');
          }
          setStatus('completed');
        },
        onError: (error) => {
          appendToLastAssistantMessage(`生成失败: ${error}`);
          setError(error);
        },
      });
    } catch (error) {
      if ((error as Error).name === 'AbortError') {
        setStatus('idle');
        return;
      }
      const message =
        error instanceof Error ? error.message : '生成失败，请重试';
      appendToLastAssistantMessage(`错误: ${message}`);
      setError(message);
    }
  }, [
    images,
    setStatus,
    setThinking,
    appendThinking,
    setGeneratedCode,
    setImageIds,
    setSessionId,
    setError,
    addUserMessage,
    addAssistantMessage,
    appendToLastAssistantMessage,
  ]);

  const cancel = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
    }
    setStatus('idle');
    setUploadProgress(0);
  }, [setStatus]);

  return { generate, cancel, uploadProgress };
}
