'use client';

import { ReactNode } from 'react';

interface MainLayoutProps {
  interactionPanel: ReactNode;  // 统一的交互面板（上传 + 对话）
  previewPanel: ReactNode;
}

export function MainLayout({
  interactionPanel,
  previewPanel,
}: MainLayoutProps) {
  // 响应式布局：小屏幕上下排列，大屏幕左右排列
  return (
    <div className="flex h-full flex-col overflow-hidden lg:flex-row">
      {/* 交互面板 - 小屏上方/大屏左侧 */}
      <aside className="max-h-[40vh] w-full flex-shrink-0 overflow-hidden border-b border-slate-200 lg:max-h-full lg:w-[320px] lg:min-w-[280px] lg:max-w-[360px] lg:border-b-0 lg:border-r">
        {interactionPanel}
      </aside>

      {/* 预览面板 */}
      <main className="min-h-0 flex-1 overflow-hidden bg-white">
        {previewPanel}
      </main>
    </div>
  );
}
