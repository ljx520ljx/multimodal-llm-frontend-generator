'use client';

import { useState, useCallback } from 'react';
import { api } from '@/lib/api/client';
import { useProjectStore } from '@/stores/useProjectStore';

export function ShareButton() {
  const sessionId = useProjectStore((state) => state.sessionId);
  const shareUrl = useProjectStore((state) => state.shareUrl);
  const setShareUrl = useProjectStore((state) => state.setShareUrl);

  const [isLoading, setIsLoading] = useState(false);
  const [copied, setCopied] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleShare = useCallback(async () => {
    if (!sessionId) return;
    setIsLoading(true);
    setError(null);

    try {
      const result = await api.share(sessionId);
      setShareUrl(result.url);
    } catch (e) {
      setError(e instanceof Error ? e.message : '分享失败');
    } finally {
      setIsLoading(false);
    }
  }, [sessionId, setShareUrl]);

  const handleCopy = useCallback(async () => {
    if (!shareUrl) return;
    try {
      await navigator.clipboard.writeText(shareUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Fallback for older browsers
      const input = document.createElement('input');
      input.value = shareUrl;
      document.body.appendChild(input);
      input.select();
      document.execCommand('copy');
      document.body.removeChild(input);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  }, [shareUrl]);

  if (!sessionId) return null;

  // Already shared - show link
  if (shareUrl) {
    return (
      <div className="flex items-center gap-1.5">
        <input
          type="text"
          value={shareUrl}
          readOnly
          className="h-7 w-36 rounded border border-slate-200 px-2 text-xs text-slate-600 focus:outline-none"
          onClick={(e) => (e.target as HTMLInputElement).select()}
        />
        <button
          onClick={handleCopy}
          className="flex h-7 items-center gap-1 rounded bg-slate-100 px-2 text-xs text-slate-600 hover:bg-slate-200 transition-colors"
        >
          {copied ? (
            <>
              <svg className="h-3 w-3 text-green-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M20 6L9 17l-5-5" />
              </svg>
              <span className="text-green-600">已复制</span>
            </>
          ) : (
            <>
              <svg className="h-3 w-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" />
              </svg>
              复制
            </>
          )}
        </button>
      </div>
    );
  }

  // Not shared yet - show share button
  return (
    <div className="flex items-center gap-1.5">
      <button
        onClick={handleShare}
        disabled={isLoading}
        className="flex h-7 items-center gap-1 rounded bg-blue-500 px-2.5 text-xs text-white hover:bg-blue-600 disabled:opacity-50 transition-colors"
      >
        {isLoading ? (
          <svg className="h-3 w-3 animate-spin" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
        ) : (
          <svg className="h-3 w-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M4 12v8a2 2 0 002 2h12a2 2 0 002-2v-8M16 6l-4-4-4 4M12 2v13" />
          </svg>
        )}
        分享
      </button>
      {error && <span className="text-xs text-red-500">{error}</span>}
    </div>
  );
}
