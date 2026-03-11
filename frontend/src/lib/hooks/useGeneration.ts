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
  const rafIdsRef = useRef<number[]>([]);

  // 公共 SSE 流处理逻辑
  const processSSEStream = useCallback(async (
    response: Response,
    doneMessage: string,
    errorPrefix: string,
  ) => {
    let codeBuffer = '';
    let thinkingBuffer = '';
    let thinkingRafPending = false;

    // Agent step counter for quality mode (maps agent events to Step format)
    let agentStepCounter = 0;
    const AGENT_DISPLAY_NAMES: Record<string, string> = {
      'LayoutAnalyzer': '布局分析',
      'ComponentDetector': '组件识别',
      'InteractionInfer': '交互推理',
      'CodeGenerator': '代码生成',
    };

    await readSSEStream(response, {
      onThinking: (content) => {
        thinkingBuffer += content;
        if (!thinkingRafPending) {
          thinkingRafPending = true;
          const id = requestAnimationFrame(() => {
            // 写入 chatMessages 以便 ChatHistory 渲染 ThinkingSteps
            appendToLastAssistantMessage(thinkingBuffer);
            appendThinking(thinkingBuffer);
            thinkingBuffer = '';
            thinkingRafPending = false;
          });
          rafIdsRef.current.push(id);
        }
      },
      onCode: (content) => {
        // Fast 模式: CODE 事件是流式 token 片段，需要累加
        // Quality 模式: CODE 事件是完整 HTML，需要替换（避免双重 HTML）
        const mode = useProjectStore.getState().generationMode;
        if (mode === 'quality') {
          codeBuffer = content;
        } else {
          codeBuffer += content;
        }
        // 立即更新预览，让用户在生成过程中就能看到原型
        const previewCode = cleanCodeContent(codeBuffer);
        if (previewCode.length > 50) {
          setGeneratedCode({
            code: previewCode,
            timestamp: Date.now(),
          });
        }
      },
      onAgentStart: (agent, content) => {
        // Format as Step to reuse ThinkingSteps rendering
        agentStepCounter++;
        const displayName = AGENT_DISPLAY_NAMES[agent] || agent;
        const msg = `Step ${agentStepCounter}: ${displayName}\n${content}\n`;
        appendToLastAssistantMessage(msg);
        appendThinking(msg);
      },
      onAgentResult: (_agent, content) => {
        // Append result summary under the current step
        if (content) {
          const msg = `${content}\n\n`;
          appendToLastAssistantMessage(msg);
          appendThinking(msg);
        }
      },
      onDone: () => {
        if (thinkingBuffer) {
          appendToLastAssistantMessage(thinkingBuffer);
          appendThinking(thinkingBuffer);
          thinkingBuffer = '';
        }
        if (codeBuffer) {
          const finalCode = cleanCodeContent(codeBuffer);
          setGeneratedCode({
            code: finalCode,
            timestamp: Date.now(),
          });
          appendToLastAssistantMessage('\n\n' + doneMessage);
        } else {
          appendToLastAssistantMessage('\n\n未能生成代码，请重试或调整设计稿');
        }
        setStatus('completed');
      },
      onError: (error) => {
        if (thinkingBuffer) {
          appendToLastAssistantMessage(thinkingBuffer);
          appendThinking(thinkingBuffer);
          thinkingBuffer = '';
        }
        appendToLastAssistantMessage(`${errorPrefix}: ${error}`);
        setError(error);
        setStatus('error');
      },
    });
  }, [appendThinking, setGeneratedCode, setStatus, setError, appendToLastAssistantMessage]);

  const generate = useCallback(async (promptText?: string) => {
    const currentImages = useProjectStore.getState().images;
    const mode = useProjectStore.getState().generationMode;

    // Text-only generation: no images but has description, requires quality mode
    const isTextOnly = currentImages.length === 0 && !!promptText && mode === 'quality';

    if (currentImages.length === 0 && !isTextOnly) {
      setError('请先上传设计稿');
      return;
    }

    try {
      setUploadProgress(0);
      setThinking('');

      abortControllerRef.current = new AbortController();

      if (isTextOnly) {
        // Text-to-UI: skip upload, create session via upload with empty files
        // then call agent generate with description
        addUserMessage(promptText || '根据描述生成 UI');
        addAssistantMessage('');

        setStatus('generating');

        // For text-only, we need a session. Create one via a lightweight upload or use existing.
        let currentSessionId = useProjectStore.getState().sessionId;
        if (!currentSessionId) {
          // Create a session by uploading an empty set (backend should handle this)
          // Actually, we need to create a session first. Use a small dummy approach:
          // The agent generate endpoint will work with session_id from a previous upload
          // or we'll need to handle creating sessions server-side.
          // For now, generate a client-side session and let the agent service handle it.
          const sessionId = `text_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`;
          setSessionId(sessionId);
          currentSessionId = sessionId;
        }

        const response = await api.generate(currentSessionId, [], 'react', mode, promptText, abortControllerRef.current.signal);

        await processSSEStream(
          response,
          '✅ 原型生成完成，请在右侧预览区体验交互效果',
          '生成失败',
        );
      } else {
        // Image-based generation: upload then generate
        setStatus('uploading');

        const imagePreviews = currentImages.map((img) => img.preview);
        addUserMessage(promptText || '生成交互原型', imagePreviews);
        addAssistantMessage('');

        const files = currentImages.map((img) => img.file);

        const uploadResult = await api.upload(files, (percent) => {
          setUploadProgress(percent);
        });

        setSessionId(uploadResult.session_id);
        const imageIds = uploadResult.images.map((img) => img.id);
        setImageIds(imageIds);

        setStatus('generating');

        const response = await api.generate(uploadResult.session_id, imageIds, 'react', mode, undefined, abortControllerRef.current!.signal);

        await processSSEStream(
          response,
          '✅ 原型生成完成，请在右侧预览区体验交互效果',
          '生成失败',
        );
      }
    } catch (error) {
      if ((error as Error).name === 'AbortError') {
        setStatus('idle');
        return;
      }
      const message =
        error instanceof Error ? error.message : '生成失败，请重试';
      setError(message);
    }
  }, [
    setStatus,
    setThinking,
    setGeneratedCode,
    setImageIds,
    setSessionId,
    setError,
    addUserMessage,
    addAssistantMessage,
    processSSEStream,
  ]);

  const regenerate = useCallback(async () => {
    const { sessionId, imageIds, generationMode } = useProjectStore.getState();
    if (!sessionId || imageIds.length === 0) {
      setError('无法重新生成：缺少会话信息，请重新上传设计稿');
      return;
    }

    try {
      setStatus('generating');
      setThinking('');
      addAssistantMessage('');

      abortControllerRef.current = new AbortController();

      const response = await api.generate(sessionId, imageIds, 'react', generationMode, undefined, abortControllerRef.current!.signal);

      await processSSEStream(
        response,
        '✅ 原型重新生成完成，请在右侧预览区体验交互效果',
        '重新生成失败',
      );
    } catch (error) {
      if ((error as Error).name === 'AbortError') {
        setStatus('idle');
        return;
      }
      const message =
        error instanceof Error ? error.message : '重新生成失败，请重试';
      setError(message);
    }
  }, [
    setStatus,
    setThinking,
    setError,
    addAssistantMessage,
    processSSEStream,
  ]);

  const cancel = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
    }
    // 清理所有待执行的 requestAnimationFrame
    rafIdsRef.current.forEach((id) => cancelAnimationFrame(id));
    rafIdsRef.current = [];
    setStatus('idle');
    setUploadProgress(0);
  }, [setStatus]);

  return { generate, regenerate, cancel, uploadProgress };
}
