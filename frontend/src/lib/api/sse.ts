import type { SSEEvent } from '@/types';

interface SSECallbacks {
  onThinking?: (content: string) => void;
  onCode?: (content: string) => void;
  onDone?: () => void;
  onError?: (error: string) => void;
}

export async function readSSEStream(
  response: Response,
  callbacks: SSECallbacks
): Promise<void> {
  const reader = response.body?.getReader();
  if (!reader) {
    throw new Error('Response body is not readable');
  }

  const decoder = new TextDecoder();
  let buffer = '';

  try {
    while (true) {
      const { done, value } = await reader.read();

      if (done) {
        break;
      }

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop() || '';

      for (const line of lines) {
        if (line.startsWith('data: ')) {
          const data = line.slice(6);

          if (data === '[DONE]') {
            callbacks.onDone?.();
            return;
          }

          try {
            const event: SSEEvent = JSON.parse(data);

            switch (event.type) {
              case 'thinking':
                callbacks.onThinking?.(event.content);
                break;
              case 'code':
                callbacks.onCode?.(event.content);
                break;
              case 'done':
                callbacks.onDone?.();
                return;
              case 'error':
                callbacks.onError?.(event.content);
                return;
            }
          } catch {
            // 非 JSON 数据，作为纯文本代码处理
            callbacks.onCode?.(data);
          }
        }
      }
    }

    // 处理剩余的 buffer
    if (buffer.startsWith('data: ')) {
      const data = buffer.slice(6);
      if (data !== '[DONE]') {
        try {
          const event: SSEEvent = JSON.parse(data);
          if (event.type === 'code') {
            callbacks.onCode?.(event.content);
          }
        } catch {
          callbacks.onCode?.(data);
        }
      }
    }

    callbacks.onDone?.();
  } finally {
    reader.releaseLock();
  }
}

export function createSSEConnection(
  url: string,
  callbacks: SSECallbacks
): EventSource {
  const eventSource = new EventSource(url);

  eventSource.addEventListener('thinking', (event) => {
    callbacks.onThinking?.(event.data);
  });

  eventSource.addEventListener('code', (event) => {
    callbacks.onCode?.(event.data);
  });

  eventSource.addEventListener('done', () => {
    callbacks.onDone?.();
    eventSource.close();
  });

  eventSource.addEventListener('error', () => {
    callbacks.onError?.('SSE connection error');
    eventSource.close();
  });

  return eventSource;
}
