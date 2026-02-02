'use client';

import Image from 'next/image';
import { useProjectStore } from '@/stores/useProjectStore';
import {
  DragDropContext,
  Droppable,
  Draggable,
  DropResult,
} from '@hello-pangea/dnd';
import type { ImageFile } from '@/types';

interface ImageListProps {
  layout?: 'vertical' | 'horizontal';
  disabled?: boolean;
}

export function ImageList({ layout = 'vertical', disabled = false }: ImageListProps) {
  const images = useProjectStore((state) => state.images);
  const reorderImages = useProjectStore((state) => state.reorderImages);
  const removeImage = useProjectStore((state) => state.removeImage);

  const handleDragEnd = (result: DropResult) => {
    if (!result.destination || disabled) return;

    const items = Array.from(images);
    const [reorderedItem] = items.splice(result.source.index, 1);
    items.splice(result.destination.index, 0, reorderedItem);

    const reordered = items.map((item, index) => ({
      ...item,
      order: index,
    }));

    reorderImages(reordered);
  };

  if (images.length === 0) {
    return null;
  }

  const isHorizontal = layout === 'horizontal';

  return (
    <DragDropContext onDragEnd={handleDragEnd}>
      <Droppable droppableId="images" direction={isHorizontal ? 'horizontal' : 'vertical'}>
        {(provided) => (
          <div
            {...provided.droppableProps}
            ref={provided.innerRef}
            className={
              isHorizontal
                ? 'flex gap-2 overflow-x-auto pb-1'
                : 'space-y-2'
            }
          >
            {images.map((image, index) => (
              isHorizontal ? (
                <HorizontalImageItem
                  key={image.id}
                  image={image}
                  index={index}
                  onRemove={() => removeImage(image.id)}
                  disabled={disabled}
                />
              ) : (
                <VerticalImageItem
                  key={image.id}
                  image={image}
                  index={index}
                  onRemove={() => removeImage(image.id)}
                  disabled={disabled}
                />
              )
            ))}
            {provided.placeholder}
          </div>
        )}
      </Droppable>
    </DragDropContext>
  );
}

interface ImageItemProps {
  image: ImageFile;
  index: number;
  onRemove: () => void;
  disabled?: boolean;
}

// 横向布局的图片卡片（更紧凑）
function HorizontalImageItem({ image, index, onRemove, disabled = false }: ImageItemProps) {
  return (
    <Draggable draggableId={image.id} index={index} isDragDisabled={disabled}>
      {(provided, snapshot) => (
        <div
          ref={provided.innerRef}
          {...provided.draggableProps}
          {...provided.dragHandleProps}
          className={`
            group relative flex-shrink-0
            ${disabled ? 'cursor-not-allowed opacity-60' : 'cursor-grab'}
            ${snapshot.isDragging ? 'z-10' : ''}
          `}
        >
          {/* 序号标签 */}
          <span className="absolute -left-1 -top-1 z-10 flex h-5 w-5 items-center justify-center rounded-full bg-blue-500 text-xs font-medium text-white shadow-sm">
            {index + 1}
          </span>

          {/* 预览图 */}
          <div
            className={`
              relative h-16 w-16 overflow-hidden rounded-lg border-2
              ${snapshot.isDragging ? 'border-blue-400 shadow-lg' : 'border-slate-200'}
            `}
          >
            <Image
              src={image.preview}
              alt={`Image ${index + 1}`}
              fill
              sizes="64px"
              className="object-cover"
              unoptimized
            />

            {/* 删除按钮 - 禁用时隐藏 */}
            {!disabled && (
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  onRemove();
                }}
                className="absolute -right-1 -top-1 z-20 flex h-5 w-5 items-center justify-center rounded-full bg-red-500 text-white opacity-0 shadow-sm transition-opacity group-hover:opacity-100"
              >
                <svg className="h-3 w-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            )}
          </div>
        </div>
      )}
    </Draggable>
  );
}

// 纵向布局的图片卡片（原设计）
function VerticalImageItem({ image, index, onRemove, disabled = false }: ImageItemProps) {
  return (
    <Draggable draggableId={image.id} index={index} isDragDisabled={disabled}>
      {(provided, snapshot) => (
        <div
          ref={provided.innerRef}
          {...provided.draggableProps}
          className={`
            group relative flex items-center gap-3 rounded-lg border bg-white p-2
            ${snapshot.isDragging ? 'border-blue-400 shadow-lg' : 'border-slate-200'}
            ${disabled ? 'opacity-60' : ''}
          `}
        >
          {/* 拖拽手柄 */}
          <div
            {...provided.dragHandleProps}
            className={`flex items-center text-slate-400 ${disabled ? 'cursor-not-allowed' : 'cursor-grab hover:text-slate-600'}`}
          >
            <svg className="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
              <circle cx="9" cy="6" r="1.5" />
              <circle cx="15" cy="6" r="1.5" />
              <circle cx="9" cy="12" r="1.5" />
              <circle cx="15" cy="12" r="1.5" />
              <circle cx="9" cy="18" r="1.5" />
              <circle cx="15" cy="18" r="1.5" />
            </svg>
          </div>

          {/* 序号 */}
          <span className="flex h-6 w-6 items-center justify-center rounded-full bg-blue-100 text-xs font-medium text-blue-700">
            {index + 1}
          </span>

          {/* 预览图 */}
          <div className="relative h-12 w-12 flex-shrink-0 overflow-hidden rounded border border-slate-200">
            <Image
              src={image.preview}
              alt={`Image ${index + 1}`}
              fill
              sizes="48px"
              className="object-cover"
              unoptimized
            />
          </div>

          {/* 文件名 */}
          <span className="flex-1 truncate text-sm text-slate-600">
            {image.file.name}
          </span>

          {/* 删除按钮 - 禁用时隐藏 */}
          {!disabled && (
            <button
              onClick={onRemove}
              className="rounded p-1 text-slate-400 opacity-0 transition-opacity hover:bg-slate-100 hover:text-red-500 group-hover:opacity-100"
            >
              <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          )}
        </div>
      )}
    </Draggable>
  );
}
