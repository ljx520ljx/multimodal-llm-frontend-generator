'use client';

import { Header, MainLayout } from '@/components/layout';
import { InteractionPanel } from '@/components/interaction';
import { PreviewPanel } from '@/components/preview';

export default function Home() {
  return (
    <div className="flex h-screen flex-col bg-slate-50">
      <Header />
      <div className="flex-1 overflow-hidden">
        <MainLayout
          interactionPanel={<InteractionPanel />}
          previewPanel={<PreviewPanel />}
        />
      </div>
    </div>
  );
}
