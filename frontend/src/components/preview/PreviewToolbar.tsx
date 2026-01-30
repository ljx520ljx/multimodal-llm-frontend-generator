'use client';

interface PreviewToolbarProps {
  onRefresh: () => void;
  onFullscreen: () => void;
  isCompiling?: boolean;
}

export function PreviewToolbar({
  onRefresh,
  onFullscreen,
  isCompiling = false,
}: PreviewToolbarProps) {
  return (
    <div className="flex items-center gap-1">
      {/* 编译状态指示器 */}
      {isCompiling && (
        <span className="flex items-center gap-1 text-xs text-amber-600">
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
          编译中
        </span>
      )}

      {/* 刷新按钮 */}
      <button
        onClick={onRefresh}
        className="rounded p-1.5 text-slate-500 hover:bg-slate-100 hover:text-slate-700"
        title="刷新预览"
      >
        <svg
          className="h-4 w-4"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
        >
          <path d="M21 12a9 9 0 11-2.636-6.364" />
          <path d="M21 3v6h-6" />
        </svg>
      </button>

      {/* 全屏按钮 */}
      <button
        onClick={onFullscreen}
        className="rounded p-1.5 text-slate-500 hover:bg-slate-100 hover:text-slate-700"
        title="全屏预览"
      >
        <svg
          className="h-4 w-4"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
        >
          <path d="M8 3H5a2 2 0 00-2 2v3m18 0V5a2 2 0 00-2-2h-3m0 18h3a2 2 0 002-2v-3M3 16v3a2 2 0 002 2h3" />
        </svg>
      </button>
    </div>
  );
}
