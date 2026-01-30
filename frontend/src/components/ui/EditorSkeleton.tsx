'use client';

export function EditorSkeleton() {
  return (
    <div className="flex h-full flex-col bg-slate-900 p-4">
      <div className="mb-4 flex items-center gap-2">
        <div className="h-4 w-4 animate-pulse rounded bg-slate-700" />
        <div className="h-3 w-24 animate-pulse rounded bg-slate-700" />
      </div>
      <div className="space-y-2">
        {Array.from({ length: 15 }).map((_, i) => (
          <div
            key={i}
            className="h-4 animate-pulse rounded bg-slate-700"
            style={{
              width: `${Math.random() * 40 + 30}%`,
              animationDelay: `${i * 50}ms`,
            }}
          />
        ))}
      </div>
    </div>
  );
}
