'use client';

import { useEffect, useRef, useState, useCallback } from 'react';
import { useProjectStore } from '@/stores/useProjectStore';
import { useDebouncedValue } from '@/lib/hooks';

const DEFAULT_HTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>原型预览</title>
  <script src="https://cdn.tailwindcss.com"></script>
  <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body>
  <div class="min-h-screen bg-gray-100 flex items-center justify-center">
    <div class="text-center">
      <h1 class="text-2xl font-bold text-gray-800 mb-4">
        上传设计稿开始生成
      </h1>
      <p class="text-gray-600">
        支持交互预览，点击体验生成的原型
      </p>
    </div>
  </div>
</body>
</html>`;

interface HtmlPreviewProps {
  onRefresh?: () => void;
  isGenerating?: boolean;
}

export function HtmlPreview({ onRefresh, isGenerating: _isGenerating = false }: HtmlPreviewProps) {
  const generatedCode = useProjectStore((state) => state.generatedCode);
  const iframeRef = useRef<HTMLIFrameElement>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const rawCode = generatedCode?.code || DEFAULT_HTML;

  // 使用防抖值，避免频繁重渲染
  const debouncedCode = useDebouncedValue(rawCode, 300);

  // 验证 HTML 代码
  const validateHtml = useCallback((html: string): boolean => {
    // 基本验证：检查是否包含必要的 HTML 结构或有效的 HTML 标签
    const hasStructure = html.includes('<html') || html.includes('<body') || html.includes('<!DOCTYPE');
    const hasValidTags = /<\w+[^>]*>/.test(html) && html.length > 50;
    return hasStructure || hasValidTags;
  }, []);

  // 更新 iframe 内容
  useEffect(() => {
    if (!iframeRef.current) return;

    setIsLoading(true);
    setError(null);

    try {
      // 验证代码
      if (!validateHtml(debouncedCode)) {
        setError('代码格式无效');
        setIsLoading(false);
        return;
      }

      // 确保代码是完整的 HTML
      let finalHtml = debouncedCode;

      // 如果代码不包含完整的 HTML 结构，包装它
      if (!finalHtml.includes('<!DOCTYPE') && !finalHtml.includes('<html')) {
        finalHtml = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <script src="https://cdn.tailwindcss.com"></script>
  <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body>
${finalHtml}
</body>
</html>`;
      }

      // 使用 srcdoc 设置 iframe 内容
      iframeRef.current.srcdoc = finalHtml;
    } catch (err) {
      setError(err instanceof Error ? err.message : '渲染失败');
    }
  }, [debouncedCode, validateHtml]);

  // 监听 iframe 加载完成
  const handleLoad = useCallback(() => {
    setIsLoading(false);
  }, []);

  return (
    <div className="relative h-full w-full flex flex-col" style={{ minHeight: '200px' }}>
      {/* 加载状态 */}
      {isLoading && (
        <div className="absolute inset-0 flex flex-col items-center justify-center bg-white/90 z-10">
          <svg className="h-8 w-8 animate-spin text-blue-500 mb-3" viewBox="0 0 24 24">
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
          <p className="text-sm text-slate-600">正在加载预览...</p>
        </div>
      )}

      {/* 错误状态 */}
      {error && (
        <div className="absolute inset-0 flex flex-col items-center justify-center bg-red-50 z-10">
          <svg
            className="mb-3 h-10 w-10 text-red-400"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
          >
            <circle cx="12" cy="12" r="10" />
            <path d="M12 8v4M12 16h.01" />
          </svg>
          <p className="text-sm font-medium text-red-700">渲染错误</p>
          <p className="mt-2 max-w-md text-center text-xs text-red-600">{error}</p>
          {onRefresh && (
            <button
              onClick={onRefresh}
              className="mt-3 rounded bg-red-100 px-3 py-1 text-xs text-red-700 hover:bg-red-200"
            >
              重新加载
            </button>
          )}
        </div>
      )}

      {/* iframe 预览 */}
      <iframe
        ref={iframeRef}
        className="flex-1 w-full border-0"
        sandbox="allow-scripts allow-same-origin"
        onLoad={handleLoad}
        title="原型预览"
      />
    </div>
  );
}
