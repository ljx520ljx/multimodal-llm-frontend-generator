'use client';

export function Header() {
  return (
    <header className="flex h-14 items-center justify-between border-b border-slate-200 bg-white px-4">
      <div className="flex items-center gap-3">
        <svg
          className="h-8 w-8 text-blue-600"
          viewBox="0 0 24 24"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <rect
            x="3"
            y="3"
            width="18"
            height="18"
            rx="2"
            stroke="currentColor"
            strokeWidth="2"
          />
          <path
            d="M3 9h18M9 21V9"
            stroke="currentColor"
            strokeWidth="2"
          />
        </svg>
        <h1 className="text-lg font-semibold text-slate-900">
          UI Prototype Generator
        </h1>
      </div>

      {/* 右侧可以添加其他功能按钮 */}
      <div className="flex items-center gap-2">
        <span className="text-xs text-slate-400">AI 驱动的交互原型生成</span>
      </div>
    </header>
  );
}
