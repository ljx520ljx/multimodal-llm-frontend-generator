'use client';

import { useState } from 'react';
import dynamic from 'next/dynamic';
import { useProjectStore } from '@/stores/useProjectStore';
import { EditorToolbar } from './EditorToolbar';
import { Skeleton, EditorSkeleton, ThinkingSteps } from '@/components/ui';

const CodeEditor = dynamic(
  () => import('./CodeEditor').then((mod) => mod.CodeEditor),
  {
    loading: () => <EditorSkeleton />,
    ssr: false,
  }
);

export function EditorPanel() {
  const generatedCode = useProjectStore((state) => state.generatedCode);
  const thinkingContent = useProjectStore((state) => state.thinkingContent);
  const status = useProjectStore((state) => state.status);

  const [thinkingExpanded, setThinkingExpanded] = useState(true);

  const hasCode = !!generatedCode?.code;
  const isGenerating = status === 'generating';

  if (!hasCode && !isGenerating) {
    return (
      <div className="flex h-full flex-col items-center justify-center text-slate-500">
        <svg
          className="mb-4 h-12 w-12 text-slate-300"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.5"
        >
          <polyline points="16,18 22,12 16,6" />
          <polyline points="8,6 2,12 8,18" />
        </svg>
        <p className="text-sm">上传设计稿后生成代码</p>
      </div>
    );
  }

  return (
    <div className="flex h-full flex-col">
      {/* 标题栏 */}
      <div className="flex items-center justify-between border-b border-slate-200 px-4 py-2">
        <span className="text-sm font-medium text-slate-700">代码编辑器</span>
        <div className="flex items-center gap-2">
          {generatedCode?.timestamp && (
            <span className="text-xs text-slate-400">
              {new Date(generatedCode.timestamp).toLocaleTimeString()}
            </span>
          )}
        </div>
      </div>

      {/* 思考过程展示（可折叠） */}
      {thinkingContent && (
        <div className="border-b border-slate-200 bg-amber-50">
          <button
            onClick={() => setThinkingExpanded(!thinkingExpanded)}
            className="flex w-full items-center justify-between px-3 py-2 text-left hover:bg-amber-100/50"
          >
            <div className="flex items-center gap-2">
              {isGenerating && (
                <svg
                  className="h-4 w-4 animate-pulse text-amber-500"
                  viewBox="0 0 24 24"
                  fill="currentColor"
                >
                  <path d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
                </svg>
              )}
              <span className="text-xs font-medium text-amber-700">
                {isGenerating ? 'AI 正在思考...' : 'AI 思考过程'}
              </span>
            </div>
            <svg
              className={`h-4 w-4 text-amber-500 transition-transform ${
                thinkingExpanded ? 'rotate-180' : ''
              }`}
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
            >
              <path d="M19 9l-7 7-7-7" />
            </svg>
          </button>
          {thinkingExpanded && (
            <div className="max-h-40 overflow-y-auto px-3 pb-3">
              <ThinkingSteps
                content={thinkingContent}
                isStreaming={isGenerating}
                variant="inline"
              />
            </div>
          )}
        </div>
      )}

      {/* 编辑器工具栏 */}
      {hasCode && <EditorToolbar />}

      {/* 编辑器 */}
      <div className="flex-1">
        {isGenerating && !hasCode ? (
          <div className="flex h-full flex-col gap-2 bg-slate-900 p-4">
            <Skeleton height={16} className="bg-slate-700" />
            <Skeleton height={16} width="80%" className="bg-slate-700" />
            <Skeleton height={16} width="60%" className="bg-slate-700" />
          </div>
        ) : (
          <CodeEditor />
        )}
      </div>
    </div>
  );
}
