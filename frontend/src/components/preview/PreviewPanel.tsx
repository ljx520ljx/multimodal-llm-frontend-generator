'use client';

import { useState, useCallback } from 'react';
import { useProjectStore } from '@/stores/useProjectStore';
import { PreviewToolbar } from './PreviewToolbar';
import { FullscreenPreview } from './FullscreenPreview';
import { CodePanel } from '@/components/editor';
import { Skeleton } from '@/components/ui';
import { HtmlPreview } from './HtmlPreview';

export function PreviewPanel() {
  const status = useProjectStore((state) => state.status);
  const generatedCode = useProjectStore((state) => state.generatedCode);
  const errorMessage = useProjectStore((state) => state.errorMessage);

  const [isFullscreen, setIsFullscreen] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);

  const hasCode = !!generatedCode?.code;
  const isGenerating = status === 'generating';
  const isError = status === 'error';

  const handleRefresh = useCallback(() => {
    setRefreshKey((k) => k + 1);
  }, []);

  const handleFullscreen = useCallback(() => {
    setIsFullscreen(true);
  }, []);

  const handleCloseFullscreen = useCallback(() => {
    setIsFullscreen(false);
  }, []);

  return (
    <div className="flex h-full flex-col">
      {/* 全屏预览弹窗 */}
      <FullscreenPreview isOpen={isFullscreen} onClose={handleCloseFullscreen} />

      {/* 标题栏 */}
      <div className="flex items-center justify-between border-b border-slate-200 bg-white px-4 py-2">
        <span className="text-sm font-medium text-slate-700">交互预览</span>
        <div className="flex items-center gap-2">
          {isGenerating && (
            <span className="flex items-center gap-1 text-xs text-blue-600">
              <svg className="h-3 w-3 animate-spin" viewBox="0 0 24 24">
                <circle
                  className="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  strokeWidth="4"
                  fill="none"
                />
                <path
                  className="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                />
              </svg>
              生成中...
            </span>
          )}
          {hasCode && !isGenerating && (
            <>
              <span className="rounded bg-green-100 px-1.5 py-0.5 text-xs text-green-700">
                可交互
              </span>
              <PreviewToolbar
                onRefresh={handleRefresh}
                onFullscreen={handleFullscreen}
              />
            </>
          )}
        </div>
      </div>

      {/* 预览区域 */}
      <div className="flex-1 overflow-hidden bg-white" style={{ minHeight: '300px' }}>
        {isError ? (
          <div className="flex h-full flex-col items-center justify-center p-4">
            <svg
              className="mb-3 h-12 w-12 text-red-400"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
            >
              <circle cx="12" cy="12" r="10" />
              <path d="M12 8v4M12 16h.01" />
            </svg>
            <p className="text-sm font-medium text-red-700">生成失败</p>
            <p className="mt-2 max-w-md text-center text-xs text-red-600">
              {errorMessage || '请检查网络连接后重试'}
            </p>
          </div>
        ) : isGenerating && !hasCode ? (
          <div className="flex h-full flex-col items-center justify-center">
            <div className="mb-6 flex items-center gap-2">
              <svg className="h-6 w-6 animate-spin text-blue-500" viewBox="0 0 24 24">
                <circle
                  className="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  strokeWidth="4"
                  fill="none"
                />
                <path
                  className="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                />
              </svg>
              <span className="text-sm font-medium text-slate-700">
                AI 正在分析设计稿并生成代码...
              </span>
            </div>
            <div className="w-64 space-y-3">
              <Skeleton height={12} className="rounded" />
              <Skeleton height={12} width="80%" className="rounded" />
              <Skeleton height={12} width="60%" className="rounded" />
            </div>
          </div>
        ) : (
          <HtmlPreview key={refreshKey} onRefresh={handleRefresh} isGenerating={isGenerating} />
        )}
      </div>

      {/* 代码折叠面板（只读） */}
      <CodePanel />
    </div>
  );
}
