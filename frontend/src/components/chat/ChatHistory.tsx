'use client';

import { useEffect, useRef, useMemo } from 'react';
import Image from 'next/image';
import type { ChatMessage } from '@/types';
import { ThinkingSteps } from '@/components/ui';

// 简单的 Markdown 渲染函数
function renderMarkdown(text: string): React.ReactNode {
  // 分割文本，处理 **粗体** 格式
  const parts = text.split(/(\*\*[^*]+\*\*)/g);

  return parts.map((part, index) => {
    // 检查是否是粗体文本
    if (part.startsWith('**') && part.endsWith('**')) {
      const content = part.slice(2, -2);
      return (
        <strong key={index} className="font-semibold text-slate-800">
          {content}
        </strong>
      );
    }
    return part;
  });
}

interface ChatHistoryProps {
  messages: ChatMessage[];
  isGenerating?: boolean;
}

interface ParsedContent {
  thinkingContent: string;
  mainContent: string;
  hasCode: boolean;
  isGeneratingCode: boolean;  // 正在生成代码（代码块未闭合）
}

// 解析消息内容，分离思考部分，过滤代码
function parseContent(content: string): ParsedContent {
  // 检查是否有完整的代码块
  const hasCompleteCodeBlock = /```[\s\S]*?```/.test(content);
  const hasHtmlCode = /<!DOCTYPE|<html|<head|<body/.test(content);
  const hasCode = hasCompleteCodeBlock || hasHtmlCode;

  // 检测是否正在生成代码（有代码块开始但未闭合，或有 HTML 但未结束）
  const codeBlockStarts = (content.match(/```/g) || []).length;
  const hasUnclosedCodeBlock = codeBlockStarts % 2 === 1;  // 奇数个 ``` 表示未闭合
  const hasUnclosedHtml = hasHtmlCode && !content.includes('</html>');
  const isGeneratingCode = hasUnclosedCodeBlock || hasUnclosedHtml;

  // 移除所有代码内容
  let textContent = content;

  // 1. 移除 markdown 代码块（包括未闭合的）
  textContent = textContent.replace(/```[\w]*\s*[\s\S]*?```/g, '');
  // 移除未闭合的代码块开始标记及其后内容
  textContent = textContent.replace(/```[\w]*[\s\S]*$/g, '');
  // 移除孤立的代码块标记
  textContent = textContent.replace(/```/g, '');

  // 2. 移除裸露的 HTML 代码（从 <!DOCTYPE 或 <html 开始到结尾）
  const doctypeIndex = textContent.indexOf('<!DOCTYPE');
  const htmlTagIndex = textContent.indexOf('<html');
  const codeStartIndex = Math.min(
    ...[doctypeIndex, htmlTagIndex].filter(i => i >= 0).concat([textContent.length])
  );

  if (codeStartIndex < textContent.length) {
    textContent = textContent.substring(0, codeStartIndex);
  }

  // 3. 移除残留的 HTML 标签（所有标签，保留 <thinking>）
  textContent = textContent.replace(/<\/?(?!thinking)[a-zA-Z][^>]*>/gi, '');

  // 3.5 清理孤立的 HTML 标签片段（如 "main>", "div>", "section>" 等）
  // 这些是未被上面正则匹配到的不完整标签残留
  textContent = textContent.replace(/\b[a-zA-Z]{1,10}>\s*$/gm, '');

  // 4. 移除 AI 常见的过渡语句（包括不完整的）
  textContent = textContent
    .replace(/下面是.*?代码[：:.]*\s*/gi, '')
    .replace(/以下是.*?代码[：:.]*\s*/gi, '')
    .replace(/完整代码如下[：:.]*\s*/gi, '')
    .replace(/代码如下[：:.]*\s*/gi, '')
    .replace(/下面.*?实现[：:.>]*\s*/gi, '')  // 处理"下面是实现代>"
    .replace(/以下.*?实现[：:.>]*\s*/gi, '')
    .replace(/下面生成.*?[：:.]*\s*/gi, '')
    .replace(/现在生成.*?[：:.]*\s*/gi, '');

  // 5. 移除状态前缀
  textContent = textContent
    .replace(/^🔄\s*正在生成代码[^\n]*\n*/gm, '')
    .replace(/^✅\s*代码生成完毕[^\n]*\n*/gm, '')
    .replace(/^✅\s*代码已更新[^\n]*\n*/gm, '')
    .trim();

  // 清理多余空行
  textContent = textContent.replace(/\n{3,}/g, '\n\n').trim();

  // 6. 提取完成/状态消息（无论何种格式，这些始终显示为 mainContent）
  const donePattern = /^.*(原型.*生成完成|代码.*已生成|代码已更新|请在右侧预览|未能生成代码|请重试或调整|生成失败|重新生成失败|Request cancelled|SSE.*超时|连接超时).*$/gm;
  const doneMatches = textContent.match(donePattern) || [];
  if (doneMatches.length > 0) {
    textContent = textContent.replace(donePattern, '').replace(/\n{3,}/g, '\n').trim();
  }
  const doneText = doneMatches.join('\n').trim();

  // 提取 <thinking> 标签内容
  const thinkingMatch = textContent.match(/<thinking>([\s\S]*?)(?:<\/thinking>|$)/);
  if (thinkingMatch) {
    const remainingContent = textContent
      .replace(/<thinking>[\s\S]*?(?:<\/thinking>|$)/, '')
      .trim();
    return {
      thinkingContent: thinkingMatch[0],
      mainContent: (remainingContent + (doneText ? '\n' + doneText : '')).trim(),
      hasCode,
      isGeneratingCode,
    };
  }

  // 检测是否包含 Step 格式的分析内容
  const hasStepContent = /Step\s*\d+[：:]/i.test(textContent);
  if (hasStepContent) {
    // Step 内容 → thinkingContent (由 ThinkingSteps 渲染)
    // done 消息 → mainContent (始终显示)
    return {
      thinkingContent: textContent,
      mainContent: doneText,
      hasCode,
      isGeneratingCode,
    };
  }

  // 没有找到思考内容格式
  // 如果内容很短或是状态消息，作为主内容显示
  // 否则可能是未格式化的分析内容，作为思考内容
  const isStatusMessage = /[✅✓⚠️❌]/.test(textContent) || textContent.includes('错误') || textContent.includes('失败');
  const isShortMessage = textContent.length < 200;

  if (isStatusMessage || isShortMessage) {
    return {
      thinkingContent: '',
      mainContent: (textContent + (doneText ? '\n' + doneText : '')).trim(),
      hasCode,
      isGeneratingCode,
    };
  }

  // 较长的内容可能是分析过程，作为思考内容；done 消息始终作为 mainContent
  return {
    thinkingContent: textContent,
    mainContent: doneText,
    hasCode,
    isGeneratingCode,
  };
}

function UserMessage({ message }: { message: ChatMessage }) {
  return (
    <div className="flex gap-2">
      <div className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-blue-100 text-xs">
        <svg className="h-3.5 w-3.5 text-blue-600" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z" />
        </svg>
      </div>
      <div className="flex-1 space-y-2">
        {message.images && message.images.length > 0 && (
          <div className="flex flex-wrap gap-1.5">
            {message.images.map((src, index) => (
              <div
                key={index}
                className="relative h-12 w-12 overflow-hidden rounded border border-slate-200"
              >
                <Image
                  src={src}
                  alt={`Image ${index + 1}`}
                  fill
                  className="object-cover"
                  unoptimized
                />
              </div>
            ))}
          </div>
        )}
        {message.content && (
          <p className="text-sm text-slate-700">{message.content}</p>
        )}
      </div>
    </div>
  );
}

// 检测消息类型
function getMessageType(content: string): 'success' | 'warning' | 'error' | 'normal' {
  if (content.includes('✅') || content.includes('代码已更新') || content.includes('代码生成完成') || content.includes('原型生成完成') || content.includes('原型重新生成完成')) {
    return 'success';
  }
  if (content.includes('⚠️') || content.includes('格式异常') || content.includes('未能生成')) {
    return 'warning';
  }
  if (content.includes('错误') || content.includes('失败') || content.includes('超时') || content.includes('cancelled')) {
    return 'error';
  }
  return 'normal';
}

function AssistantMessage({ message, isStreaming }: { message: ChatMessage; isStreaming?: boolean }) {
  const parsed = useMemo(() => parseContent(message.content), [message.content]);
  const messageType = useMemo(() => getMessageType(parsed.mainContent), [parsed.mainContent]);

  return (
    <div className="flex gap-2">
      <div className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-emerald-100 text-xs">
        <svg className="h-3.5 w-3.5 text-emerald-600" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z" />
        </svg>
      </div>
      <div className="flex-1 min-w-0">
        {/* 思考步骤展示 */}
        {parsed.thinkingContent && (
          <ThinkingSteps
            content={parsed.thinkingContent}
            isStreaming={!!isStreaming}
            variant="card"
          />
        )}

        {/* 状态消息（成功/警告/错误） */}
        {!isStreaming && parsed.mainContent && messageType !== 'normal' && (
          <div className={`flex items-center gap-1.5 rounded-md px-2 py-1.5 text-xs ${
            messageType === 'success' ? 'bg-emerald-50 text-emerald-700' :
            messageType === 'warning' ? 'bg-amber-50 text-amber-700' :
            'bg-red-50 text-red-700'
          }`}>
            {messageType === 'success' && (
              <svg className="h-3.5 w-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            )}
            {messageType === 'warning' && (
              <svg className="h-3.5 w-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            )}
            {messageType === 'error' && (
              <svg className="h-3.5 w-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            )}
            {renderMarkdown(parsed.mainContent)}
          </div>
        )}

        {/* 普通文本内容 - 支持基本 Markdown 格式 */}
        {parsed.mainContent && messageType === 'normal' && (
          <p className="whitespace-pre-wrap text-sm text-slate-700">
            {renderMarkdown(parsed.mainContent)}
            {isStreaming && (
              <span className="ml-0.5 inline-block h-3 w-1 animate-pulse bg-slate-400" />
            )}
          </p>
        )}

        {/* 正在生成代码状态 */}
        {isStreaming && parsed.isGeneratingCode && (
          <div className="flex items-center gap-2 rounded-md bg-blue-50 px-3 py-2 text-xs text-blue-700">
            <svg className="h-4 w-4 animate-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="12" cy="12" r="10" strokeOpacity="0.25" />
              <path d="M12 2a10 10 0 0 1 10 10" strokeLinecap="round" />
            </svg>
            正在生成代码...
          </div>
        )}

        {/* 代码生成成功提示（当没有其他消息时） */}
        {!isStreaming && parsed.hasCode && !parsed.mainContent && (
          <div className="flex items-center gap-1.5 rounded-md bg-emerald-50 px-2 py-1.5 text-xs text-emerald-700">
            <svg className="h-3.5 w-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            代码已生成，请在右侧预览区查看效果
          </div>
        )}

        {/* 如果正在流式输出且没有主要内容，显示加载状态 */}
        {isStreaming && !parsed.mainContent && !parsed.thinkingContent && (
          <div className="flex items-center gap-1.5 text-xs text-slate-400">
            <span className="inline-block h-1.5 w-1.5 animate-pulse rounded-full bg-slate-400" />
            正在分析...
          </div>
        )}
      </div>
    </div>
  );
}

export function ChatHistory({ messages, isGenerating = false }: ChatHistoryProps) {
  const scrollRef = useRef<HTMLDivElement>(null);

  // 自动滚动到底部
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages]);

  if (messages.length === 0) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center p-4 text-center text-slate-400">
        <svg
          className="mb-2 h-8 w-8 text-slate-300"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.5"
        >
          <path d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
        </svg>
        <p className="text-xs">上传设计稿并点击发送开始生成</p>
      </div>
    );
  }

  return (
    <div
      ref={scrollRef}
      className="flex-1 space-y-3 overflow-y-auto p-3"
    >
      {messages.map((message, index) => (
        <div key={message.id}>
          {message.role === 'user' ? (
            <UserMessage message={message} />
          ) : (
            <AssistantMessage
              message={message}
              isStreaming={isGenerating && index === messages.length - 1}
            />
          )}
        </div>
      ))}
    </div>
  );
}
