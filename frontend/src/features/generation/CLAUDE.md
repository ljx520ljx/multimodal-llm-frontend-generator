# L3 - 代码生成功能文档 | Generation Feature

> DocOps 层级: **L3 (Implementation Level)**
> 父文档: [L2 前端模块文档](../../../CLAUDE.md)

## 功能职责

调用后端 API 生成代码，处理流式响应，展示生成进度和结果。

## 目录结构

```
features/generation/
├── components/
│   ├── GenerateButton.tsx     # 生成触发按钮
│   ├── GenerationStatus.tsx   # 生成状态指示
│   └── ThinkingDisplay.tsx    # 思考过程展示
│
├── hooks/
│   ├── useCodeGeneration.ts   # 生成逻辑 Hook
│   └── useStreamReader.ts     # SSE 流式读取
│
├── types.ts
├── index.ts
└── CLAUDE.md
```

## 核心实现

### 类型定义

```typescript
// types.ts
export interface GenerationRequest {
  sessionId: string;
  imageIds: string[];
  framework: 'react' | 'vue';
}

export interface GenerationEvent {
  type: 'thinking' | 'code' | 'error' | 'done';
  content: string;
}

export interface GenerationState {
  status: 'idle' | 'generating' | 'completed' | 'error';
  thinking: string;
  code: string;
  error: string | null;
}
```

### useCodeGeneration Hook

```typescript
// hooks/useCodeGeneration.ts
'use client';

import { useState, useCallback } from 'react';
import { useStreamReader } from './useStreamReader';
import type { GenerationRequest, GenerationState, GenerationEvent } from '../types';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export function useCodeGeneration() {
  const [state, setState] = useState<GenerationState>({
    status: 'idle',
    thinking: '',
    code: '',
    error: null,
  });

  const { readStream } = useStreamReader();

  const generate = useCallback(async (request: GenerationRequest) => {
    setState({
      status: 'generating',
      thinking: '',
      code: '',
      error: null,
    });

    try {
      const response = await fetch(`${API_URL}/api/generate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(request),
      });

      if (!response.ok) {
        throw new Error(`Generation failed: ${response.statusText}`);
      }

      await readStream(response.body!, (event: GenerationEvent) => {
        setState((prev) => {
          switch (event.type) {
            case 'thinking':
              return { ...prev, thinking: prev.thinking + event.content };
            case 'code':
              return { ...prev, code: prev.code + event.content };
            case 'error':
              return { ...prev, status: 'error', error: event.content };
            case 'done':
              return { ...prev, status: 'completed' };
            default:
              return prev;
          }
        });
      });

    } catch (error) {
      setState((prev) => ({
        ...prev,
        status: 'error',
        error: error instanceof Error ? error.message : 'Unknown error',
      }));
    }
  }, [readStream]);

  const reset = useCallback(() => {
    setState({
      status: 'idle',
      thinking: '',
      code: '',
      error: null,
    });
  }, []);

  return { state, generate, reset };
}
```

### useStreamReader Hook

```typescript
// hooks/useStreamReader.ts
import { useCallback } from 'react';
import type { GenerationEvent } from '../types';

export function useStreamReader() {
  const readStream = useCallback(async (
    stream: ReadableStream<Uint8Array>,
    onEvent: (event: GenerationEvent) => void
  ) => {
    const reader = stream.getReader();
    const decoder = new TextDecoder();
    let buffer = '';

    try {
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });

        // 解析 SSE 格式
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';  // 保留不完整的行

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const data = line.slice(6);
            if (data === '[DONE]') {
              onEvent({ type: 'done', content: '' });
              return;
            }

            try {
              const event: GenerationEvent = JSON.parse(data);
              onEvent(event);
            } catch {
              // 忽略解析错误
            }
          }
        }
      }
    } finally {
      reader.releaseLock();
    }
  }, []);

  return { readStream };
}
```

### GenerationStatus 组件

```typescript
// components/GenerationStatus.tsx
import type { GenerationState } from '../types';

export interface GenerationStatusProps {
  state: GenerationState;
}

export function GenerationStatus({ state }: GenerationStatusProps) {
  const { status, thinking } = state;

  if (status === 'idle') return null;

  return (
    <div className="p-4 rounded-lg bg-gray-800">
      <div className="flex items-center gap-2 mb-2">
        {status === 'generating' && (
          <>
            <div className="w-3 h-3 bg-blue-500 rounded-full animate-pulse" />
            <span className="text-blue-400">正在生成代码...</span>
          </>
        )}
        {status === 'completed' && (
          <>
            <div className="w-3 h-3 bg-green-500 rounded-full" />
            <span className="text-green-400">生成完成</span>
          </>
        )}
        {status === 'error' && (
          <>
            <div className="w-3 h-3 bg-red-500 rounded-full" />
            <span className="text-red-400">生成失败</span>
          </>
        )}
      </div>

      {thinking && (
        <div className="mt-4 p-3 bg-gray-900 rounded text-sm text-gray-400">
          <p className="font-medium text-gray-300 mb-2">AI 思考过程：</p>
          <p className="whitespace-pre-wrap">{thinking}</p>
        </div>
      )}
    </div>
  );
}
```

### GenerateButton 组件

```typescript
// components/GenerateButton.tsx
import { Button } from '@/components/ui';
import type { GenerationState } from '../types';

export interface GenerateButtonProps {
  disabled: boolean;
  state: GenerationState;
  onGenerate: () => void;
}

export function GenerateButton({ disabled, state, onGenerate }: GenerateButtonProps) {
  const isGenerating = state.status === 'generating';

  return (
    <Button
      onClick={onGenerate}
      disabled={disabled || isGenerating}
      loading={isGenerating}
      size="lg"
      className="w-full"
    >
      {isGenerating ? '正在生成...' : '生成代码'}
    </Button>
  );
}
```

## 状态流转

```
idle → generating → completed
         ↓
       error
```

## 错误处理

| 错误类型 | 处理方式 |
|----------|----------|
| 网络错误 | 显示重试按钮 |
| API 错误 | 显示错误信息 |
| 流解析错误 | 忽略，继续读取 |
| 超时 | 显示超时提示 |

## 测试用例

```typescript
// __tests__/useCodeGeneration.test.ts
import { renderHook, act, waitFor } from '@testing-library/react';
import { useCodeGeneration } from '../hooks/useCodeGeneration';

describe('useCodeGeneration', () => {
  it('should handle successful generation', async () => {
    const { result } = renderHook(() => useCodeGeneration());

    // Mock fetch
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      body: createMockStream([
        { type: 'thinking', content: 'Analyzing...' },
        { type: 'code', content: 'const App = () => {}' },
        { type: 'done', content: '' },
      ]),
    });

    await act(async () => {
      await result.current.generate({
        sessionId: 'test',
        imageIds: ['1'],
        framework: 'react',
      });
    });

    await waitFor(() => {
      expect(result.current.state.status).toBe('completed');
      expect(result.current.state.code).toContain('App');
    });
  });
});
```
