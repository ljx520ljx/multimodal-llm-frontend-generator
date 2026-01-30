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
  // 单一布局：交互面板 + 预览
  return (
    <div className="flex h-full overflow-hidden">
      {/* 交互面板 - 左侧 */}
      <aside className="w-[320px] min-w-[280px] max-w-[360px] flex-shrink-0 overflow-hidden border-r border-slate-200">
        {interactionPanel}
      </aside>

      {/* 预览面板 */}
      <main className="flex-1 overflow-hidden bg-white">
        {previewPanel}
      </main>
    </div>
  );
}
