'use client';

import { useEffect } from 'react';
import { Header, MainLayout } from '@/components/layout';
import { InteractionPanel } from '@/components/interaction';
import { PreviewPanel } from '@/components/preview';
import { useAuthStore } from '@/stores/useAuthStore';
import { api } from '@/lib/api/client';

export default function Home() {
  const setAuth = useAuthStore((state) => state.setAuth);

  // Handle GitHub OAuth callback: read token from URL fragment
  useEffect(() => {
    const hash = window.location.hash;
    if (!hash.includes('access_token=')) return;

    const params = new URLSearchParams(hash.slice(1));
    const accessToken = params.get('access_token');
    if (!accessToken) return;

    // Clear the hash from URL
    window.history.replaceState(null, '', window.location.pathname);

    // Fetch user info with the token and store auth state
    const token = { access_token: accessToken, expires_at: 0 };
    api.auth.me(accessToken).then((res) => {
      setAuth(res.user, token);
    }).catch(() => {
      // Token invalid, ignore
    });
  }, [setAuth]);

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
