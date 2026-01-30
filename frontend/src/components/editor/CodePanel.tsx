'use client';

import { useProjectStore } from '@/stores/useProjectStore';
import { CodeEditor } from './CodeEditor';

export function CodePanel() {
  const codeExpanded = useProjectStore((state) => state.codeExpanded);
  const toggleCodeExpanded = useProjectStore((state) => state.toggleCodeExpanded);
  const generatedCode = useProjectStore((state) => state.generatedCode);

  const hasCode = !!generatedCode?.code;

  return (
    <div className="border-t border-slate-200 bg-slate-900">
      {/* 折叠/展开按钮 */}
      <button
        onClick={toggleCodeExpanded}
        disabled={!hasCode}
        className={`
          flex w-full items-center justify-between px-4 py-2 text-sm
          transition-colors
          ${
            hasCode
              ? 'text-slate-300 hover:bg-slate-800'
              : 'cursor-not-allowed text-slate-500'
          }
        `}
      >
        <span className="flex items-center gap-2">
          <svg
            className="h-4 w-4"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
          >
            <polyline points="16,18 22,12 16,6" />
            <polyline points="8,6 2,12 8,18" />
          </svg>
          查看代码
        </span>
        <span className="flex items-center gap-1">
          {hasCode && (
            <span className="rounded bg-slate-700 px-1.5 py-0.5 text-xs">
              TSX
            </span>
          )}
          <svg
            className={`h-4 w-4 transition-transform ${codeExpanded ? 'rotate-180' : ''}`}
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
          >
            <polyline points="6 9 12 15 18 9" />
          </svg>
        </span>
      </button>

      {/* 代码查看器（只读） */}
      {codeExpanded && hasCode && (
        <div className="h-[300px] border-t border-slate-700">
          <CodeEditor readOnly />
        </div>
      )}
    </div>
  );
}
