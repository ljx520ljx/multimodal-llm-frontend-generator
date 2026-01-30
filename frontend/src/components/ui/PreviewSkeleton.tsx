'use client';

export function PreviewSkeleton() {
  return (
    <div className="flex h-full flex-col items-center justify-center bg-slate-50 p-8">
      <div className="mb-4 h-10 w-10 animate-spin rounded-full border-4 border-blue-200 border-t-blue-500" />
      <div className="text-sm font-medium text-slate-600">正在初始化预览环境...</div>
      <div className="mt-2 text-xs text-slate-400">首次加载可能需要几秒钟</div>
    </div>
  );
}
