'use client';

import { useState, useRef, useCallback, KeyboardEvent, ClipboardEvent } from 'react';
import Image from 'next/image';
import { Button } from '@/components/ui';

interface PastedImage {
  id: string;
  file: File;
  preview: string;
}

interface UnifiedInputProps {
  onSend: (text: string, images: File[]) => void;
  disabled?: boolean;
  placeholder?: string;
  buttonText?: string;
  hasUploadedImages?: boolean; // 是否已有上传的图片（在 store 中）
}

function generateId() {
  return Math.random().toString(36).substring(2, 11);
}

export function UnifiedInput({
  onSend,
  disabled = false,
  placeholder = '输入需求描述，支持粘贴图片...',
  buttonText = '发送',
  hasUploadedImages = false,
}: UnifiedInputProps) {
  const [text, setText] = useState('');
  const [pastedImages, setPastedImages] = useState<PastedImage[]>([]);
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
      const newImages = imageFiles.map((file) => ({
        id: generateId(),
        file,
        preview: URL.createObjectURL(file),
      }));
      setPastedImages((prev) => [...prev, ...newImages]);
    }
  }, []);

  const handleRemoveImage = useCallback((id: string) => {
    setPastedImages((prev) => {
      const removed = prev.find((img) => img.id === id);
      if (removed) {
        URL.revokeObjectURL(removed.preview);
      }
      return prev.filter((img) => img.id !== id);
    });
  }, []);

  const handleSend = useCallback(() => {
    const trimmedText = text.trim();
    const files = pastedImages.map((img) => img.file);

    // 如果没有文字、没有粘贴图片、也没有已上传的图片，则不发送
    if (!trimmedText && files.length === 0 && !hasUploadedImages) return;

    onSend(trimmedText, files);

    // 清空状态
    setText('');
    pastedImages.forEach((img) => URL.revokeObjectURL(img.preview));
    setPastedImages([]);
  }, [text, pastedImages, onSend, hasUploadedImages]);

  const handleKeyDown = useCallback(
    (e: KeyboardEvent<HTMLTextAreaElement>) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        handleSend();
      }
    },
    [handleSend]
  );

  // 可以发送的条件：有文字、有粘贴图片、或已有上传的图片
  const canSend = text.trim() || pastedImages.length > 0 || hasUploadedImages;

  return (
    <div className="border-t border-slate-200 bg-white p-3">
      {/* 粘贴图片预览 */}
      {pastedImages.length > 0 && (
        <div className="mb-2 flex flex-wrap gap-2">
          {pastedImages.map((img) => (
            <div
              key={img.id}
              className="group relative h-16 w-16 overflow-hidden rounded-lg border border-slate-200"
            >
              <Image
                src={img.preview}
                alt="Pasted image"
                fill
                className="object-cover"
                unoptimized
              />
              <button
                onClick={() => handleRemoveImage(img.id)}
                className="absolute -right-1 -top-1 flex h-5 w-5 items-center justify-center rounded-full bg-red-500 text-white opacity-0 transition-opacity group-hover:opacity-100"
              >
                <svg className="h-3 w-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          ))}
        </div>
      )}

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
