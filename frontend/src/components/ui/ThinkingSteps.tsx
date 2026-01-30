'use client';

import { useMemo } from 'react';

interface ParsedStep {
  number: number;
  title: string;
  content: string;
}

interface ThinkingStepsProps {
  content: string;
  isStreaming?: boolean;
  variant?: 'card' | 'inline'; // card: 对话框样式, inline: 编辑器样式
}

// 清理 markdown 格式和表格内容
function cleanMarkdownContent(text: string): string {
  return text
    .replace(/```[\w]*[\s\S]*?```/g, '')  // 移除完整代码块
    .replace(/```[\w]*[\s\S]*$/g, '')     // 移除未闭合的代码块
    .replace(/```/g, '')                   // 移除孤立的代码块标记
    .replace(/\|[^|]*\|/g, '')             // 移除表格单元格
    .replace(/\|-+\|/g, '')                // 移除表格分隔线
    .replace(/^[\s|:-]+$/gm, '')           // 移除只包含表格字符的行
    .replace(/\*\*([^*]+)\*\*/g, '$1')     // 移除粗体
    .replace(/\*([^*]+)\*/g, '$1')         // 移除斜体
    .replace(/`([^`]+)`/g, '$1')           // 移除行内代码
    .replace(/^#+\s*/gm, '')               // 移除标题标记
    .replace(/^[-*]\s*/gm, '• ')           // 将列表符号统一为 •
    .replace(/下面是.*代码[：:.]*/gi, '')   // 移除过渡语句
    .replace(/以下是.*代码[：:.]*/gi, '')
    .replace(/<\/?(html|head|body)[^>]*>/gi, '')  // 移除 HTML 标签残留
    .replace(/\n{3,}/g, '\n\n')            // 移除多余空行
    .trim();
}

// 解析思考内容
function parseThinkingContent(content: string): { steps: ParsedStep[]; otherContent: string } {
  const steps: ParsedStep[] = [];

  // 移除 <thinking> 标签和表格内容
  const cleanContent = content
    .replace(/<thinking>/gi, '')
    .replace(/<\/thinking>/gi, '')
    .replace(/\|[^\n]*\|/g, '')           // 移除表格行
    .replace(/^\s*[-|:]+\s*$/gm, '');     // 移除表格分隔线

  // 尝试多种步骤格式（按优先级排序）
  // 格式1: ### Step X: Title
  // 格式2: **Step X: Title**
  // 格式3: Step X: Title (无标记，最宽松)
  const patterns = [
    /###\s*Step\s*(\d+)[：:]\s*([^\n]+)\n([\s\S]*?)(?=###\s*Step|$)/gi,
    /\*\*Step\s*(\d+)[：:]\s*([^*]+)\*\*\s*([\s\S]*?)(?=\*\*Step\s*\d|$)/gi,
    /Step\s*(\d+)[：:]\s*([^\n]+?)(?:\s{2,}|\n)([\s\S]*?)(?=Step\s*\d+[：:]|$)/gi,
  ];

  for (const pattern of patterns) {
    let match;
    pattern.lastIndex = 0; // 重置正则状态

    while ((match = pattern.exec(cleanContent)) !== null) {
      // 清理内容中的 markdown 格式
      const stepContent = cleanMarkdownContent(match[3]);

      steps.push({
        number: parseInt(match[1], 10),
        title: match[2].trim(),
        content: stepContent,
      });
    }

    // 如果找到了步骤，就不再尝试其他格式
    if (steps.length > 0) {
      // 按步骤编号排序
      steps.sort((a, b) => a.number - b.number);
      break;
    }
  }

  // 如果没有找到步骤格式，返回清理后的内容
  const otherContent = steps.length === 0
    ? cleanMarkdownContent(cleanContent)
    : '';

  return { steps, otherContent };
}

// 步骤图标
function StepIcon({ number, isComplete, size = 'sm' }: { number: number; isComplete: boolean; size?: 'sm' | 'xs' }) {
  const sizeClass = size === 'sm' ? 'h-5 w-5 text-xs' : 'h-4 w-4 text-[10px]';
  const iconSize = size === 'sm' ? 'h-3 w-3' : 'h-2.5 w-2.5';

  return (
    <div className={`flex shrink-0 items-center justify-center rounded-full font-medium ${sizeClass} ${
      isComplete ? 'bg-emerald-500 text-white' : 'bg-blue-100 text-blue-600'
    }`}>
      {isComplete ? (
        <svg className={iconSize} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3">
          <path d="M5 13l4 4L19 7" />
        </svg>
      ) : (
        number
      )}
    </div>
  );
}

export function ThinkingSteps({ content, isStreaming = false, variant = 'card' }: ThinkingStepsProps) {
  const { steps, otherContent } = useMemo(() => parseThinkingContent(content), [content]);

  if (steps.length === 0 && !otherContent) {
    return null;
  }

  // 编辑器内联样式
  if (variant === 'inline') {
    if (steps.length === 0) {
      return (
        <p className="text-xs text-amber-600 whitespace-pre-wrap">
          {otherContent}
        </p>
      );
    }

    return (
      <div className="space-y-2">
        {steps.map((step, index) => {
          const isComplete = !isStreaming || index < steps.length - 1;
          return (
            <div key={step.number} className="flex gap-2">
              <StepIcon number={step.number} isComplete={isComplete} size="xs" />
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-1">
                  <span className="text-xs font-medium text-amber-700">{step.title}</span>
                  {!isComplete && (
                    <span className="inline-block h-1 w-1 animate-pulse rounded-full bg-amber-500" />
                  )}
                </div>
                {step.content && (
                  <p className="mt-0.5 text-xs text-amber-600 line-clamp-2">{step.content}</p>
                )}
              </div>
            </div>
          );
        })}
      </div>
    );
  }

  // 对话框卡片样式 - 详细展示思考过程
  if (steps.length === 0) {
    return (
      <div className="mb-3 rounded-lg border border-blue-100 bg-blue-50/50 p-3">
        <div className="mb-2 flex items-center gap-1.5 text-xs font-medium text-blue-600">
          <svg className="h-3.5 w-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
          </svg>
          AI 思考过程
          {isStreaming && (
            <span className="inline-block h-1.5 w-1.5 animate-pulse rounded-full bg-blue-500" />
          )}
        </div>
        <p className="whitespace-pre-wrap text-sm text-slate-600 leading-relaxed">{otherContent}</p>
      </div>
    );
  }

  return (
    <div className="mb-3 rounded-lg border border-blue-100 bg-blue-50/50 p-3">
      <div className="mb-3 flex items-center gap-1.5 text-xs font-medium text-blue-600">
        <svg className="h-3.5 w-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
        </svg>
        AI 分析过程
        {isStreaming && (
          <span className="ml-1 text-[10px] text-blue-500">(分析中...)</span>
        )}
      </div>
      <div className="space-y-3">
        {steps.map((step, index) => {
          const isComplete = !isStreaming || index < steps.length - 1;
          const isCurrentStep = isStreaming && index === steps.length - 1;
          return (
            <div
              key={step.number}
              className={`flex gap-2.5 rounded-md p-2 transition-colors ${
                isCurrentStep ? 'bg-white/80 shadow-sm' : ''
              }`}
            >
              <StepIcon number={step.number} isComplete={isComplete} />
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-1.5">
                  <span className={`text-sm font-medium ${isCurrentStep ? 'text-blue-700' : 'text-slate-700'}`}>
                    {step.title}
                  </span>
                  {!isComplete && (
                    <span className="inline-block h-1.5 w-1.5 animate-pulse rounded-full bg-blue-500" />
                  )}
                </div>
                {step.content && (
                  <p className="mt-1 text-sm text-slate-600 leading-relaxed whitespace-pre-wrap">
                    {step.content}
                  </p>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
