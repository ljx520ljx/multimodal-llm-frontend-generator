'use client';

import Link from 'next/link';
import { useAuthStore } from '@/stores/useAuthStore';

export function Header() {
  const user = useAuthStore((state) => state.user);
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

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

      <div className="flex items-center gap-3">
        <span className="text-xs text-slate-400">AI 驱动的交互原型生成</span>
        {isAuthenticated ? (
          <Link
            href="/dashboard"
            className="flex items-center gap-1.5 rounded-md border border-slate-200 px-2.5 py-1 text-xs text-slate-600 hover:bg-slate-50 transition-colors"
          >
            {user?.avatar_url ? (
              <img src={user.avatar_url} alt="" className="h-4 w-4 rounded-full" />
            ) : (
              <div className="flex h-4 w-4 items-center justify-center rounded-full bg-blue-100 text-[10px] font-medium text-blue-600">
                {(user?.display_name || user?.email || '?')[0].toUpperCase()}
              </div>
            )}
            {user?.display_name || '我的项目'}
          </Link>
        ) : (
          <Link
            href="/login"
            className="rounded-md bg-slate-100 px-2.5 py-1 text-xs text-slate-600 hover:bg-slate-200 transition-colors"
          >
            登录
          </Link>
        )}
      </div>
    </header>
  );
}
