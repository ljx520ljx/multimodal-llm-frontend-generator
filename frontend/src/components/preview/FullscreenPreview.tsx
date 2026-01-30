'use client';

import { useEffect, useState, useCallback } from 'react';
import { HtmlPreview } from './HtmlPreview';

interface FullscreenPreviewProps {
  isOpen: boolean;
  onClose: () => void;
}

export function FullscreenPreview({ isOpen, onClose }: FullscreenPreviewProps) {
  const [refreshKey, setRefreshKey] = useState(0);

  const handleRefresh = useCallback(() => {
    setRefreshKey((k) => k + 1);
  }, []);

  // ESC 键关闭
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        onClose();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, onClose]);

  // 禁止背景滚动
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => {
      document.body.style.overflow = '';
    };
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-sm">
      <div className="absolute inset-4 flex flex-col rounded-lg bg-white shadow-2xl">
        {/* 工具栏 */}
        <div className="flex items-center justify-between border-b border-slate-200 px-4 py-2">
          <span className="text-sm font-medium text-slate-700">全屏预览</span>
          <div className="flex items-center gap-2">
            <span className="rounded bg-green-100 px-2 py-0.5 text-xs text-green-700">
              可交互
            </span>
            {/* 刷新按钮 */}
            <button
              onClick={handleRefresh}
              className="rounded p-1.5 text-slate-500 hover:bg-slate-100 hover:text-slate-700"
              title="重置演示"
            >
              <svg
                className="h-5 w-5"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
              >
                <path d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
            </button>
            {/* 关闭按钮 */}
            <button
              onClick={onClose}
              className="rounded p-1.5 text-slate-500 hover:bg-slate-100 hover:text-slate-700"
              title="退出全屏 (ESC)"
            >
              <svg
                className="h-5 w-5"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
              >
                <path d="M18 6L6 18M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        {/* 预览区域 */}
        <div className="flex-1 overflow-hidden">
          <HtmlPreview key={refreshKey} onRefresh={handleRefresh} />
        </div>

        {/* 底部提示 */}
        <div className="border-t border-slate-200 px-4 py-2 text-center text-xs text-slate-500">
          按 ESC 键退出全屏模式
        </div>
      </div>
    </div>
  );
}
