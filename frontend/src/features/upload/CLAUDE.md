# L3 - 图片上传功能文档 | Upload Feature

> DocOps 层级: **L3 (Implementation Level)**
> 父文档: [L2 前端模块文档](../../../CLAUDE.md)

## 功能职责

处理 UI 设计稿的上传、预览、排序和管理。

## 目录结构

```
features/upload/
├── components/
│   ├── ImageUploader.tsx      # 上传区域组件
│   ├── ImagePreview.tsx       # 图片预览卡片
│   ├── ImageSorter.tsx        # 拖拽排序列表
│   └── UploadProgress.tsx     # 上传进度指示
│
├── hooks/
│   ├── useImageUpload.ts      # 上传逻辑 Hook
│   └── useImageSorter.ts      # 排序逻辑 Hook
│
├── utils/
│   ├── imageCompressor.ts     # 图片压缩
│   └── fileValidator.ts       # 文件验证
│
├── types.ts                    # 类型定义
├── index.ts                    # 导出汇总
└── CLAUDE.md                   # 本文档
```

## 核心实现

### 类型定义

```typescript
// types.ts
export interface ImageFile {
  id: string;
  file: File;
  preview: string;        // blob URL
  order: number;
  status: 'pending' | 'uploading' | 'uploaded' | 'error';
  uploadProgress: number;
}

export interface UploadResult {
  sessionId: string;
  images: Array<{
    id: string;
    url: string;
    order: number;
  }>;
}
```

### ImageUploader 组件

```typescript
// components/ImageUploader.tsx
'use client';

import { useCallback } from 'react';
import { useDropzone } from 'react-dropzone';
import { useImageUpload } from '../hooks/useImageUpload';

export interface ImageUploaderProps {
  onUpload: (files: ImageFile[]) => void;
  maxFiles?: number;
  maxSize?: number;  // bytes
}

export function ImageUploader({
  onUpload,
  maxFiles = 10,
  maxSize = 10 * 1024 * 1024,  // 10MB
}: ImageUploaderProps) {
  const { processFiles } = useImageUpload();

  const onDrop = useCallback(async (acceptedFiles: File[]) => {
    const processed = await processFiles(acceptedFiles);
    onUpload(processed);
  }, [processFiles, onUpload]);

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: {
      'image/png': ['.png'],
      'image/jpeg': ['.jpg', '.jpeg'],
      'image/webp': ['.webp'],
    },
    maxFiles,
    maxSize,
  });

  return (
    <div
      {...getRootProps()}
      className={`
        border-2 border-dashed rounded-lg p-8 text-center cursor-pointer
        transition-colors duration-200
        ${isDragActive ? 'border-blue-500 bg-blue-50' : 'border-gray-300 hover:border-gray-400'}
      `}
    >
      <input {...getInputProps()} />
      <div className="text-gray-600">
        {isDragActive ? (
          <p>释放以上传图片...</p>
        ) : (
          <>
            <p>拖拽 UI 设计稿到此处，或点击选择文件</p>
            <p className="text-sm text-gray-400 mt-2">
              支持 PNG, JPG, WebP，最大 {maxSize / 1024 / 1024}MB
            </p>
          </>
        )}
      </div>
    </div>
  );
}
```

### useImageUpload Hook

```typescript
// hooks/useImageUpload.ts
import { useCallback, useState } from 'react';
import { nanoid } from 'nanoid';
import { compressImage } from '../utils/imageCompressor';
import { validateFile } from '../utils/fileValidator';
import type { ImageFile } from '../types';

export function useImageUpload() {
  const [isProcessing, setIsProcessing] = useState(false);

  const processFiles = useCallback(async (files: File[]): Promise<ImageFile[]> => {
    setIsProcessing(true);

    try {
      const processed = await Promise.all(
        files.map(async (file, index) => {
          // 验证文件
          const validation = validateFile(file);
          if (!validation.valid) {
            throw new Error(validation.error);
          }

          // 压缩图片
          const compressed = await compressImage(file, {
            maxWidth: 2048,
            maxHeight: 2048,
            quality: 0.8,
          });

          return {
            id: nanoid(),
            file: compressed,
            preview: URL.createObjectURL(compressed),
            order: index,
            status: 'pending' as const,
            uploadProgress: 0,
          };
        })
      );

      return processed;
    } finally {
      setIsProcessing(false);
    }
  }, []);

  return { processFiles, isProcessing };
}
```

### ImageSorter 组件

```typescript
// components/ImageSorter.tsx
'use client';

import { DragDropContext, Droppable, Draggable, type DropResult } from '@hello-pangea/dnd';
import { ImagePreview } from './ImagePreview';
import type { ImageFile } from '../types';

export interface ImageSorterProps {
  images: ImageFile[];
  onReorder: (images: ImageFile[]) => void;
  onRemove: (id: string) => void;
}

export function ImageSorter({ images, onReorder, onRemove }: ImageSorterProps) {
  const handleDragEnd = (result: DropResult) => {
    if (!result.destination) return;

    const items = Array.from(images);
    const [reordered] = items.splice(result.source.index, 1);
    items.splice(result.destination.index, 0, reordered);

    // 更新 order 字段
    const updated = items.map((item, index) => ({
      ...item,
      order: index,
    }));

    onReorder(updated);
  };

  return (
    <DragDropContext onDragEnd={handleDragEnd}>
      <Droppable droppableId="images" direction="horizontal">
        {(provided) => (
          <div
            ref={provided.innerRef}
            {...provided.droppableProps}
            className="flex gap-4 overflow-x-auto p-4"
          >
            {images.map((image, index) => (
              <Draggable key={image.id} draggableId={image.id} index={index}>
                {(provided, snapshot) => (
                  <div
                    ref={provided.innerRef}
                    {...provided.draggableProps}
                    {...provided.dragHandleProps}
                  >
                    <ImagePreview
                      image={image}
                      isDragging={snapshot.isDragging}
                      onRemove={() => onRemove(image.id)}
                    />
                  </div>
                )}
              </Draggable>
            ))}
            {provided.placeholder}
          </div>
        )}
      </Droppable>
    </DragDropContext>
  );
}
```

### 图片压缩工具

```typescript
// utils/imageCompressor.ts
export interface CompressOptions {
  maxWidth: number;
  maxHeight: number;
  quality: number;
}

export async function compressImage(
  file: File,
  options: CompressOptions
): Promise<File> {
  const { maxWidth, maxHeight, quality } = options;

  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => {
      let { width, height } = img;

      // 计算缩放比例
      if (width > maxWidth || height > maxHeight) {
        const ratio = Math.min(maxWidth / width, maxHeight / height);
        width *= ratio;
        height *= ratio;
      }

      // 使用 Canvas 压缩
      const canvas = document.createElement('canvas');
      canvas.width = width;
      canvas.height = height;

      const ctx = canvas.getContext('2d')!;
      ctx.drawImage(img, 0, 0, width, height);

      canvas.toBlob(
        (blob) => {
          if (blob) {
            resolve(new File([blob], file.name, { type: 'image/webp' }));
          } else {
            reject(new Error('Failed to compress image'));
          }
        },
        'image/webp',
        quality
      );
    };

    img.onerror = () => reject(new Error('Failed to load image'));
    img.src = URL.createObjectURL(file);
  });
}
```

## 测试用例

```typescript
// __tests__/useImageUpload.test.ts
import { renderHook, act } from '@testing-library/react';
import { useImageUpload } from '../hooks/useImageUpload';

describe('useImageUpload', () => {
  it('should process files and return ImageFile array', async () => {
    const { result } = renderHook(() => useImageUpload());

    const mockFile = new File(['test'], 'test.png', { type: 'image/png' });

    await act(async () => {
      const processed = await result.current.processFiles([mockFile]);
      expect(processed).toHaveLength(1);
      expect(processed[0]).toHaveProperty('id');
      expect(processed[0]).toHaveProperty('preview');
      expect(processed[0].order).toBe(0);
    });
  });
});
```
