'use client';

import { useState, useRef, useCallback, KeyboardEvent, ClipboardEvent } from 'react';
import { Button } from '@/components/ui';

interface UnifiedInputProps {
  onSend: (text: string) => void;
  onImagesPaste?: (files: File[]) => void; // 粘贴图片回调，直接添加到上传区域
  disabled?: boolean;
  placeholder?: string;
  buttonText?: string;
  hasUploadedImages?: boolean; // 是否已有上传的图片（在 store 中）
}

export function UnifiedInput({
  onSend,
  onImagesPaste,
  disabled = false,
  placeholder = '输入需求描述，支持粘贴图片...',
  buttonText = '发送',
  hasUploadedImages = false,
}: UnifiedInputProps) {
  const [text, setText] = useState('');
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const handlePaste = useCallback((e: ClipboardEvent<HTMLTextAreaElement>) => {
    const items = e.clipboardData?.items;
    if (!items) return;

    const imageFiles: File[] = [];
    for (const item of Array.from(items)) {
      if (item.type.startsWith('image/')) {
        const file = item.getAsFile();
        if (file) {
          imageFiles.push(file);
        }
      }
    }

    if (imageFiles.length > 0) {
      e.preventDefault();
      // 粘贴的图片直接添加到上传区域（store），不在输入框内显示
      onImagesPaste?.(imageFiles);
    }
  }, [onImagesPaste]);

  const handleSend = useCallback(() => {
    const trimmedText = text.trim();

    // 如果没有文字、也没有已上传的图片，则不发送
    if (!trimmedText && !hasUploadedImages) return;

    onSend(trimmedText);

    // 清空文字
    setText('');
  }, [text, onSend, hasUploadedImages]);

  const handleKeyDown = useCallback(
    (e: KeyboardEvent<HTMLTextAreaElement>) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        handleSend();
      }
    },
    [handleSend]
  );

  // 可以发送的条件：有文字或已有上传的图片
  const canSend = text.trim() || hasUploadedImages;

  return (
    <div className="border-t border-slate-200 bg-white p-3">
      {/* 输入区域 */}
      <div className="flex items-end gap-2">
        <div className="flex-1">
          <textarea
            ref={textareaRef}
            value={text}
            onChange={(e) => setText(e.target.value)}
            onPaste={handlePaste}
            onKeyDown={handleKeyDown}
            placeholder={placeholder}
            disabled={disabled}
            rows={1}
            className="w-full resize-none rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm text-slate-900 placeholder-slate-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 disabled:cursor-not-allowed disabled:bg-slate-50 disabled:opacity-60"
            style={{ minHeight: '40px', maxHeight: '120px' }}
          />
        </div>
        <Button
          onClick={handleSend}
          disabled={disabled || !canSend}
          size="sm"
          className="h-10 px-4"
        >
          {buttonText}
        </Button>
      </div>

      <p className="mt-1 text-xs text-slate-400">
        Enter 发送，Shift+Enter 换行，支持 Ctrl+V 粘贴图片
      </p>
    </div>
  );
}
