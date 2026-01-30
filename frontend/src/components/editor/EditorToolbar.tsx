'use client';

import { useState, useCallback } from 'react';
import { useProjectStore } from '@/stores/useProjectStore';

export function EditorToolbar() {
  const activeFile = useProjectStore((state) => state.activeFile);
  const setActiveFile = useProjectStore((state) => state.setActiveFile);
  const generatedCode = useProjectStore((state) => state.generatedCode);
  const generatedCss = useProjectStore((state) => state.generatedCss);

  const [copySuccess, setCopySuccess] = useState(false);

  const handleCopy = useCallback(async () => {
    const content = activeFile === 'App.tsx'
      ? generatedCode?.code || ''
      : generatedCss || '';

    try {
      await navigator.clipboard.writeText(content);
      setCopySuccess(true);
      setTimeout(() => setCopySuccess(false), 2000);
    } catch (err) {
      console.error('Failed to copy:', err);
    }
  }, [activeFile, generatedCode, generatedCss]);

  return (
    <div className="flex items-center justify-between border-b border-slate-200 bg-slate-50 px-2 py-1">
      {/* 文件标签页 */}
      <div className="flex items-center gap-1">
        <button
          onClick={() => setActiveFile('App.tsx')}
          className={`rounded px-2 py-1 text-xs font-medium transition-colors ${
            activeFile === 'App.tsx'
              ? 'bg-white text-slate-900 shadow-sm'
              : 'text-slate-500 hover:text-slate-700'
          }`}
        >
          App.tsx
        </button>
        <button
          onClick={() => setActiveFile('styles.css')}
          className={`rounded px-2 py-1 text-xs font-medium transition-colors ${
            activeFile === 'styles.css'
              ? 'bg-white text-slate-900 shadow-sm'
              : 'text-slate-500 hover:text-slate-700'
          }`}
        >
          styles.css
        </button>
      </div>

      {/* 工具按钮 */}
      <div className="flex items-center gap-1">
        {/* 复制按钮 */}
        <button
          onClick={handleCopy}
          className="flex items-center gap-1 rounded px-2 py-1 text-xs text-slate-500 hover:bg-white hover:text-slate-700"
          title="复制代码"
        >
          {copySuccess ? (
            <>
              <svg
                className="h-3.5 w-3.5 text-green-500"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
              >
                <path d="M20 6L9 17l-5-5" />
              </svg>
              已复制
            </>
          ) : (
            <>
              <svg
                className="h-3.5 w-3.5"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
              >
                <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" />
              </svg>
              复制
            </>
          )}
        </button>
      </div>
    </div>
  );
}
