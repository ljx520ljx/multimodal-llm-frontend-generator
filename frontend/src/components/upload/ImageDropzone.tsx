'use client';

import { useCallback, useState } from 'react';
import { useDropzone } from 'react-dropzone';
import { useProjectStore } from '@/stores/useProjectStore';
import {
  compressImage,
  COMPRESSION_THRESHOLD_MB,
} from '@/lib/utils/imageCompression';
import type { ImageFile } from '@/types';

function generateId() {
  return Math.random().toString(36).substring(2, 11);
}

interface ImageDropzoneProps {
  disabled?: boolean;
}

interface CompressionStatus {
  isCompressing: boolean;
  currentFile: string;
  progress: number;
  totalFiles: number;
  processedFiles: number;
}

export function ImageDropzone({ disabled = false }: ImageDropzoneProps) {
  const addImages = useProjectStore((state) => state.addImages);
  const images = useProjectStore((state) => state.images);
  const [compressionStatus, setCompressionStatus] =
    useState<CompressionStatus | null>(null);
  const [compressionError, setCompressionError] = useState<string | null>(null);

  const onDrop = useCallback(
    async (acceptedFiles: File[]) => {
      if (disabled) return;

      setCompressionError(null);
      const newImages: ImageFile[] = [];
      const filesToCompress = acceptedFiles.filter(
        (f) => f.size > COMPRESSION_THRESHOLD_MB * 1024 * 1024
      );

      if (filesToCompress.length > 0) {
        setCompressionStatus({
          isCompressing: true,
          currentFile: '',
          progress: 0,
          totalFiles: acceptedFiles.length,
          processedFiles: 0,
        });
      }

      for (let i = 0; i < acceptedFiles.length; i++) {
        const file = acceptedFiles[i];

        try {
          const needsCompression =
            file.size > COMPRESSION_THRESHOLD_MB * 1024 * 1024;

          if (needsCompression) {
            setCompressionStatus((prev) =>
              prev
                ? {
                    ...prev,
                    currentFile: file.name,
                    progress: 0,
                    processedFiles: i,
                  }
                : null
            );
          }

          const result = await compressImage(file, {}, (progress) => {
            setCompressionStatus((prev) =>
              prev ? { ...prev, progress } : null
            );
          });

          newImages.push({
            id: generateId(),
            file: result.file,
            preview: URL.createObjectURL(result.file),
            order: images.length + i,
          });
        } catch {
          setCompressionError(
            `压缩失败: ${file.name}。将使用原图上传。`
          );
          newImages.push({
            id: generateId(),
            file,
            preview: URL.createObjectURL(file),
            order: images.length + i,
          });
        }
      }

      setCompressionStatus(null);
      addImages(newImages);
    },
    [addImages, images.length, disabled]
  );

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: {
      'image/*': ['.png', '.jpg', '.jpeg', '.gif', '.webp'],
    },
    multiple: true,
    disabled,
  });

  const isProcessing = compressionStatus?.isCompressing;

  return (
    <div className="space-y-2">
      <div
        {...getRootProps()}
        className={`
          flex flex-col items-center justify-center
          rounded-lg border-2 border-dashed p-3 transition-colors
          ${disabled || isProcessing
            ? 'cursor-not-allowed border-slate-200 bg-slate-50 opacity-60'
            : isDragActive
              ? 'cursor-pointer border-blue-500 bg-blue-50'
              : 'cursor-pointer border-slate-300 bg-white hover:border-blue-400 hover:bg-slate-50'
          }
        `}
      >
        <input {...getInputProps()} disabled={isProcessing} />
        {isProcessing ? (
          <>
            <div className="mb-2 h-6 w-6 animate-spin rounded-full border-2 border-blue-200 border-t-blue-500" />
            <p className="mb-1 text-xs font-medium text-slate-700">
              压缩中: {compressionStatus.currentFile}
            </p>
            <div className="w-full max-w-xs">
              <div className="mb-1 h-1.5 w-full overflow-hidden rounded-full bg-slate-200">
                <div
                  className="h-full bg-blue-500 transition-all duration-300"
                  style={{ width: `${compressionStatus.progress}%` }}
                />
              </div>
              <p className="text-center text-xs text-slate-500">
                {compressionStatus.processedFiles + 1} /{' '}
                {compressionStatus.totalFiles} 文件
              </p>
            </div>
          </>
        ) : (
          <div className="flex items-center gap-3">
            <svg
              className={`h-6 w-6 flex-shrink-0 ${isDragActive ? 'text-blue-500' : 'text-slate-400'}`}
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
            >
              <path d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14" />
              <path d="M14 8h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            <div>
              <p className="text-xs font-medium text-slate-700">
                {isDragActive ? '释放以上传图片' : '拖拽或点击上传设计稿'}
              </p>
              <p className="text-xs text-slate-500">
                PNG, JPG, GIF, WebP（大于 2MB 自动压缩）
              </p>
            </div>
          </div>
        )}
      </div>
      {compressionError && (
        <p className="text-center text-xs text-amber-600">{compressionError}</p>
      )}
    </div>
  );
}
