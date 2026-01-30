import imageCompression from 'browser-image-compression';

export interface CompressionOptions {
  maxSizeMB: number;
  maxWidthOrHeight: number;
  quality: number;
  useWebWorker: boolean;
}

export const DEFAULT_COMPRESSION_OPTIONS: CompressionOptions = {
  maxSizeMB: 2,
  maxWidthOrHeight: 2048,
  quality: 0.8,
  useWebWorker: true,
};

export const COMPRESSION_THRESHOLD_MB = 2;

export interface CompressionResult {
  file: File;
  wasCompressed: boolean;
  originalSize: number;
  compressedSize: number;
}

export async function compressImage(
  file: File,
  options: Partial<CompressionOptions> = {},
  onProgress?: (progress: number) => void
): Promise<CompressionResult> {
  const originalSize = file.size;

  if (originalSize <= COMPRESSION_THRESHOLD_MB * 1024 * 1024) {
    return {
      file,
      wasCompressed: false,
      originalSize,
      compressedSize: originalSize,
    };
  }

  const compressionOptions = {
    ...DEFAULT_COMPRESSION_OPTIONS,
    ...options,
    onProgress: onProgress
      ? (progress: number) => onProgress(Math.round(progress))
      : undefined,
  };

  const adjustedOptions = {
    maxSizeMB: compressionOptions.maxSizeMB,
    maxWidthOrHeight: compressionOptions.maxWidthOrHeight,
    useWebWorker: compressionOptions.useWebWorker,
    initialQuality: compressionOptions.quality,
    onProgress: compressionOptions.onProgress,
  };

  try {
    const compressedFile = await imageCompression(file, adjustedOptions);

    return {
      file: compressedFile,
      wasCompressed: true,
      originalSize,
      compressedSize: compressedFile.size,
    };
  } catch (error) {
    console.error('Image compression failed:', error);
    throw error;
  }
}

export async function compressImages(
  files: File[],
  options: Partial<CompressionOptions> = {},
  onFileProgress?: (fileIndex: number, progress: number) => void
): Promise<CompressionResult[]> {
  const results: CompressionResult[] = [];

  for (let i = 0; i < files.length; i++) {
    const result = await compressImage(files[i], options, (progress) =>
      onFileProgress?.(i, progress)
    );
    results.push(result);
  }

  return results;
}

export function formatFileSize(bytes: number): string {
  if (bytes < 1024) {
    return `${bytes} B`;
  }
  if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(1)} KB`;
  }
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
