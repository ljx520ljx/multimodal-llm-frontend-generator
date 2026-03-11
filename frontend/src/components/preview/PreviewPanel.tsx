'use client';

import { useState, useCallback, useRef } from 'react';
import { useProjectStore } from '@/stores/useProjectStore';
import { PreviewToolbar } from './PreviewToolbar';
import { FullscreenPreview } from './FullscreenPreview';
import { CodePanel } from '@/components/editor';
import { Skeleton } from '@/components/ui';
import { HtmlPreview, type SelectedElementInfo } from './HtmlPreview';
import { useChat } from '@/lib/hooks/useChat';
import { useGeneration } from '@/lib/hooks/useGeneration';
import { ShareButton } from './ShareButton';

export function PreviewPanel() {
  const status = useProjectStore((state) => state.status);
  const generatedCode = useProjectStore((state) => state.generatedCode);
  const errorMessage = useProjectStore((state) => state.errorMessage);
  const sessionId = useProjectStore((state) => state.sessionId);

  const [isFullscreen, setIsFullscreen] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);
  const [annotationMode, setAnnotationMode] = useState(false);
  const [selectedElement, setSelectedElement] = useState<SelectedElementInfo | null>(null);
  const [modifyInput, setModifyInput] = useState('');
  const inputRef = useRef<HTMLInputElement>(null);

  const { sendMessage } = useChat();
  const { regenerate } = useGeneration();

  const hasCode = !!generatedCode?.code;
  const isGenerating = status === 'generating';
  const isError = status === 'error';

  const handleRefresh = useCallback(() => {
    setRefreshKey((k) => k + 1);
  }, []);

  const handleFullscreen = useCallback(() => {
    setIsFullscreen(true);
  }, []);

  const handleCloseFullscreen = useCallback(() => {
    setIsFullscreen(false);
  }, []);

  // 切换标注模式
  const handleToggleAnnotation = useCallback(() => {
    setAnnotationMode((prev) => !prev);
    setSelectedElement(null);
    setModifyInput('');
  }, []);

  // 元素选中回调
  const handleElementSelect = useCallback((info: SelectedElementInfo) => {
    setSelectedElement(info);
    setModifyInput('');
    // 自动聚焦输入框
    setTimeout(() => inputRef.current?.focus(), 100);
  }, []);

  // 关闭修改弹窗
  const handleCloseModifyDialog = useCallback(() => {
    setSelectedElement(null);
    setModifyInput('');
  }, []);

  // 发送修改请求
  const handleSendModify = useCallback(async () => {
    if (!selectedElement || !modifyInput.trim() || !sessionId) return;

    // 构建描述：包含元素信息和用户需求
    const elementDesc = [
      `元素: ${selectedElement.tagName}`,
      selectedElement.classList.length > 0 ? `类名: ${selectedElement.classList.join(' ')}` : '',
      selectedElement.text ? `文本: "${selectedElement.text.slice(0, 50)}"` : '',
    ].filter(Boolean).join(', ');

    const message = `请修改这个元素（${elementDesc}）：${modifyInput}`;

    // 关闭弹窗和标注模式
    setSelectedElement(null);
    setModifyInput('');
    setAnnotationMode(false);

    // 发送修改请求
    await sendMessage(sessionId, message);
  }, [selectedElement, modifyInput, sessionId, sendMessage]);

  return (
    <div className="flex h-full flex-col">
      {/* 全屏预览弹窗 */}
      <FullscreenPreview isOpen={isFullscreen} onClose={handleCloseFullscreen} />

      {/* 标题栏 */}
      <div className="flex items-center justify-between border-b border-slate-200 bg-white px-4 py-2">
        <span className="text-sm font-medium text-slate-700">交互预览</span>
        <div className="flex items-center gap-2">
          {isGenerating && (
            <span className="flex items-center gap-1 text-xs text-blue-600">
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
              生成中...
            </span>
          )}
          {hasCode && !isGenerating && (
            <>
              <span className="rounded bg-green-100 px-1.5 py-0.5 text-xs text-green-700">
                可交互
              </span>
              <ShareButton />
              {/* 标注模式按钮 */}
              <button
                onClick={handleToggleAnnotation}
                className={`flex items-center gap-1 rounded px-2 py-1 text-xs transition-colors ${
                  annotationMode
                    ? 'bg-blue-500 text-white'
                    : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
                }`}
                title={annotationMode ? '退出标注模式' : '点击元素进行标注修改'}
              >
                <svg className="h-3.5 w-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.12 2.122" />
                </svg>
                {annotationMode ? '标注中' : '标注修改'}
              </button>
              <PreviewToolbar
                onRefresh={handleRefresh}
                onFullscreen={handleFullscreen}
              />
            </>
          )}
        </div>
      </div>

      {/* 预览区域 */}
      <div className="relative flex-1 overflow-hidden bg-white" style={{ minHeight: '300px' }}>
        {isError ? (
          <div className="flex h-full flex-col items-center justify-center p-4">
            <svg
              className="mb-3 h-12 w-12 text-red-400"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
            >
              <circle cx="12" cy="12" r="10" />
              <path d="M12 8v4M12 16h.01" />
            </svg>
            <p className="text-sm font-medium text-red-700">生成失败</p>
            <p className="mt-2 max-w-md text-center text-xs text-red-600">
              {errorMessage || '请检查网络连接后重试'}
            </p>
            <button
              onClick={regenerate}
              className="mt-4 flex items-center gap-1.5 rounded-md bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 transition-colors"
            >
              <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M1 4v6h6M23 20v-6h-6" />
                <path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 0 1 3.51 15" />
              </svg>
              重新生成
            </button>
          </div>
        ) : isGenerating && !hasCode ? (
          <div className="flex h-full flex-col items-center justify-center">
            <div className="mb-6 flex items-center gap-2">
              <svg className="h-6 w-6 animate-spin text-blue-500" viewBox="0 0 24 24">
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
              <span className="text-sm font-medium text-slate-700">
                AI 正在分析设计稿并生成代码...
              </span>
            </div>
            <div className="w-64 space-y-3">
              <Skeleton height={12} className="rounded" />
              <Skeleton height={12} width="80%" className="rounded" />
              <Skeleton height={12} width="60%" className="rounded" />
            </div>
          </div>
        ) : (
          <HtmlPreview
            key={refreshKey}
            onRefresh={handleRefresh}
            isGenerating={isGenerating}
            annotationMode={annotationMode}
            onElementSelect={handleElementSelect}
          />
        )}

        {/* 标注模式提示 */}
        {annotationMode && !selectedElement && (
          <div className="absolute bottom-4 left-1/2 -translate-x-1/2 rounded-lg bg-blue-500 px-4 py-2 text-sm text-white shadow-lg">
            点击页面中的元素进行标注修改
          </div>
        )}

        {/* 元素修改弹窗 */}
        {selectedElement && (
          <div className="absolute inset-0 flex items-center justify-center bg-black/30">
            <div className="mx-4 w-full max-w-md rounded-lg bg-white p-4 shadow-xl">
              <div className="mb-3 flex items-center justify-between">
                <h3 className="font-medium text-slate-900">标注修改</h3>
                <button
                  onClick={handleCloseModifyDialog}
                  className="text-slate-400 hover:text-slate-600"
                >
                  <svg className="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>

              {/* 选中元素信息 */}
              <div className="mb-3 rounded bg-slate-50 p-2 text-xs text-slate-600">
                <div className="flex items-center gap-2">
                  <span className="rounded bg-slate-200 px-1.5 py-0.5 font-mono">
                    {selectedElement.tagName}
                  </span>
                  {selectedElement.classList.length > 0 && (
                    <span className="truncate text-slate-500">
                      .{selectedElement.classList.slice(0, 3).join('.')}
                    </span>
                  )}
                </div>
                {selectedElement.text && (
                  <p className="mt-1 truncate text-slate-500">
                    &ldquo;{selectedElement.text.slice(0, 50)}{selectedElement.text.length > 50 ? '...' : ''}&rdquo;
                  </p>
                )}
              </div>

              {/* 修改输入 */}
              <input
                ref={inputRef}
                type="text"
                value={modifyInput}
                onChange={(e) => setModifyInput(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleSendModify()}
                placeholder="描述修改需求，如：改成红色、字体变大..."
                className="mb-3 w-full rounded-md border border-slate-200 px-3 py-2 text-sm text-slate-900 placeholder:text-slate-500 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                autoFocus
              />

              {/* 操作按钮 */}
              <div className="flex justify-end gap-2">
                <button
                  onClick={handleCloseModifyDialog}
                  className="rounded-md px-3 py-1.5 text-sm text-slate-600 hover:bg-slate-100"
                >
                  取消
                </button>
                <button
                  onClick={handleSendModify}
                  disabled={!modifyInput.trim()}
                  className="rounded-md bg-blue-500 px-3 py-1.5 text-sm text-white hover:bg-blue-600 disabled:opacity-50"
                >
                  发送修改
                </button>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* 代码折叠面板（只读） */}
      <CodePanel />
    </div>
  );
}
