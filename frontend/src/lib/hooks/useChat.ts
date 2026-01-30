'use client';

import { useCallback, useRef } from 'react';
import { useProjectStore } from '@/stores/useProjectStore';
import { api, readSSEStream } from '@/lib/api';

export function useChat() {
  const generatedCode = useProjectStore((state) => state.generatedCode);
  const updateCode = useProjectStore((state) => state.updateCode);
  const setStatus = useProjectStore((state) => state.setStatus);
  const setError = useProjectStore((state) => state.setError);

  // 对话历史操作
  const addUserMessage = useProjectStore((state) => state.addUserMessage);
  const addAssistantMessage = useProjectStore((state) => state.addAssistantMessage);
  const appendToLastAssistantMessage = useProjectStore((state) => state.appendToLastAssistantMessage);

  const abortControllerRef = useRef<AbortController | null>(null);

  const sendMessage = useCallback(
    async (sessionId: string, message: string) => {
      if (!generatedCode?.code) {
        setError('请先生成代码');
        return;
      }

      if (!message.trim()) {
        return;
      }

      // 添加用户消息
      addUserMessage(message);

      // 添加 AI 消息占位
      addAssistantMessage('');

      setStatus('generating');

      // 创建 abort controller
      abortControllerRef.current = new AbortController();

      try {
        const response = await api.chat(sessionId, message);

        let codeBuffer = '';

        await readSSEStream(response, {
          onThinking: (content) => {
            // 只追加思考内容到对话，不追加代码
            appendToLastAssistantMessage(content);
          },
          onCode: (content) => {
            // 代码只存入 buffer，不追加到对话历史
            codeBuffer += content;
          },
          onDone: () => {
            // 提取代码并更新编辑器
            const extractedCode = extractCodeFromResponse(codeBuffer);
            if (extractedCode) {
              updateCode(extractedCode);
              // 始终显示成功消息
              appendToLastAssistantMessage('\n\n✅ 代码已更新，请在右侧预览区查看效果');
            } else if (codeBuffer) {
              // 有代码内容但提取失败，尝试直接使用
              const cleanedCode = codeBuffer
                .replace(/```[\w]*\s*\n?/g, '')
                .replace(/\n?```\s*$/g, '')
                .trim();
              if (cleanedCode.length > 50) {
                updateCode(cleanedCode);
                appendToLastAssistantMessage('\n\n✅ 代码已更新，请在右侧预览区查看效果');
              } else {
                appendToLastAssistantMessage('\n\n⚠️ 代码格式异常，请重试');
              }
            } else {
              // 没有代码返回
              appendToLastAssistantMessage('\n\n⚠️ 未能生成修改后的代码，请重新描述需求');
            }
            setStatus('completed');
          },
          onError: (err) => {
            appendToLastAssistantMessage(`错误: ${err}`);
            setError(err);
            setStatus('error');
          },
        });
      } catch (error) {
        if ((error as Error).name === 'AbortError') {
          setStatus('completed');
          return;
        }
        const message = error instanceof Error ? error.message : '发送失败';
        appendToLastAssistantMessage(`错误: ${message}`);
        setError(message);
        setStatus('error');
      }
    },
    [generatedCode, updateCode, setStatus, setError, addUserMessage, addAssistantMessage, appendToLastAssistantMessage]
  );

  const cancel = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
    }
    setStatus('completed');
  }, [setStatus]);

  return { sendMessage, cancel };
}

// 从响应中提取代码，支持 HTML 和 React 代码
function extractCodeFromResponse(content: string): string | null {
  if (!content || content.trim().length < 10) {
    return null;
  }

  let cleaned = content;

  // 1. 首先移除 thinking 标签及其内容
  cleaned = cleaned.replace(/<thinking>[\s\S]*?<\/thinking>/gi, '');
  cleaned = cleaned.replace(/<\/?thinking>/gi, '');

  // 2. 尝试匹配 markdown 代码块（支持各种语言标记）
  const pattern = /```[\w]*\s*\n?([\s\S]*?)```/g;
  const matches: RegExpExecArray[] = [];
  let match: RegExpExecArray | null;
  while ((match = pattern.exec(cleaned)) !== null) {
    matches.push(match);
  }

  let code: string;

  if (matches.length > 0) {
    // 查找包含完整 HTML 结构的代码块
    let bestMatch: string | null = null;
    for (const m of matches) {
      const blockContent = m[1] || '';
      // 优先选择包含完整 HTML 结构的代码块
      if (blockContent.includes('<!DOCTYPE') || blockContent.includes('<html')) {
        bestMatch = blockContent;
        break;
      }
      // 次选包含 div 或 body 结构的代码块
      if (!bestMatch && (blockContent.includes('<div') || blockContent.includes('<body'))) {
        bestMatch = blockContent;
      }
    }
    // 如果没有找到结构化的代码块，使用最后一个代码块
    code = bestMatch || matches[matches.length - 1][1] || '';
  } else {
    // 没有代码块标记，尝试找代码起始点
    // 支持 HTML 和 React 两种格式
    const startMarkers = [
      { marker: '<!DOCTYPE', index: cleaned.indexOf('<!DOCTYPE') },
      { marker: '<html', index: cleaned.indexOf('<html') },
      { marker: '<head', index: cleaned.indexOf('<head') },
      { marker: '<body', index: cleaned.indexOf('<body') },
      { marker: 'import ', index: cleaned.indexOf('import ') },
      { marker: 'export ', index: cleaned.indexOf('export ') },
    ];

    const validStarts = startMarkers.filter(m => m.index >= 0);
    if (validStarts.length === 0) {
      return null;
    }

    const earliest = validStarts.reduce((min, curr) =>
      curr.index < min.index ? curr : min
    );

    code = cleaned.substring(earliest.index);
  }

  // 3. 清理可能残留的代码块标记
  code = code
    .replace(/^```[\w]*\s*\n?/gm, '')
    .replace(/\n?```\s*$/gm, '')
    .trim();

  // 4. 再次确保没有 thinking 标签残留
  code = code.replace(/<\/?thinking>/gi, '');

  // 5. 确保代码有效
  if (!code || code.length < 10) {
    return null;
  }

  // 6. 验证代码特征（支持 HTML 和 React）
  const hasValidCode =
    // 完整 HTML 特征
    code.includes('<!DOCTYPE') ||
    code.includes('<html') ||
    // HTML 片段特征（需要有标签和属性）
    (code.includes('<div') && (code.includes('class=') || code.includes('x-'))) ||
    (code.includes('<body') && code.includes('<')) ||
    // React 特征
    code.includes('import') ||
    code.includes('export') ||
    code.includes('function') ||
    (code.includes('return') && code.includes('<'));

  if (!hasValidCode) {
    return null;
  }

  return code;
}
