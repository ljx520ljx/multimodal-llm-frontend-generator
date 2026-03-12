import { SSEEventSchema } from './schemas';

interface SSECallbacks {
  onThinking?: (content: string) => void;
  onCode?: (content: string) => void;
  onDone?: () => void;
  onError?: (error: string) => void;
  onAgentStart?: (agent: string, content: string) => void;
  onAgentResult?: (agent: string, content: string) => void;
}

/**
 * Timeout in ms — if no data received within this window, consider connection stalled.
 * This is the outermost timeout in the chain. Under normal conditions, inner layers
 * (Python LLM 120s → Go Agent 360s → Go Handler 480s) should timeout first and
 * return meaningful errors. This 600s value is a last-resort safety net.
 * Quality 模式需要 4+ 个 Agent 串行调用，总时间可达 200-400s。
 */
const SSE_READ_TIMEOUT = 600_000; // 10 minutes (outermost timeout)

export async function readSSEStream(
  response: Response,
  callbacks: SSECallbacks
): Promise<void> {
  if (!response.ok) {
    const errorText = await response.text().catch(() => 'Unknown error');
    throw new Error(`SSE request failed (${response.status}): ${errorText}`);
  }

  const reader = response.body?.getReader();
  if (!reader) {
    throw new Error('Response body is not readable');
  }

  const decoder = new TextDecoder();
  let buffer = '';
  let doneFired = false;

  try {
    while (true) {
      // Race between read and timeout (with cleanup to avoid timer leak)
      const readPromise = reader.read();
      let timeoutId: ReturnType<typeof setTimeout>;
      const timeoutPromise = new Promise<never>((_, reject) => {
        timeoutId = setTimeout(() => reject(new Error('SSE read timeout')), SSE_READ_TIMEOUT);
      });

      let result: ReadableStreamReadResult<Uint8Array>;
      try {
        result = await Promise.race([readPromise, timeoutPromise]);
      } catch (err) {
        clearTimeout(timeoutId!);
        callbacks.onError?.(err instanceof Error ? err.message : 'SSE 连接超时');
        doneFired = true;
        return;
      }
      clearTimeout(timeoutId!);

      const { done, value } = result;

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
            doneFired = true;
            callbacks.onDone?.();
            return;
          }

          try {
            const raw = JSON.parse(data);
            const result = SSEEventSchema.safeParse(raw);

            if (!result.success) {
              // schema 验证失败：只有当数据看起来像 HTML 时才当 code 处理，
              // 否则忽略，避免 JSON 垃圾污染 codeBuffer
              if (typeof raw === 'string' || (typeof data === 'string' && data.trimStart().startsWith('<'))) {
                callbacks.onCode?.(data);
              }
              continue;
            }

            const event = result.data;

            switch (event.type) {
              case 'thinking':
                callbacks.onThinking?.(event.content);
                break;
              case 'code':
                callbacks.onCode?.(event.content);
                break;
              case 'agent_start':
                callbacks.onAgentStart?.(event.agent || '', event.content);
                break;
              case 'agent_result':
                callbacks.onAgentResult?.(event.agent || '', event.content);
                break;
              case 'done':
                doneFired = true;
                callbacks.onDone?.();
                return;
              case 'error':
                doneFired = true;
                callbacks.onError?.(event.content);
                return;
              // tool_call / tool_result 事件：静默忽略（不污染 codeBuffer）
              case 'tool_call':
              case 'tool_result':
                break;
            }
          } catch {
            // 非 JSON 数据：只有看起来像 HTML 的内容才当 code 处理
            if (data.trimStart().startsWith('<')) {
              callbacks.onCode?.(data);
            }
          }
        }
      }
    }

    // 处理剩余的 buffer
    if (buffer.startsWith('data: ')) {
      const data = buffer.slice(6);
      if (data !== '[DONE]') {
        try {
          const raw = JSON.parse(data);
          const result = SSEEventSchema.safeParse(raw);
          if (result.success) {
            if (result.data.type === 'code') {
              callbacks.onCode?.(result.data.content);
            } else if (result.data.type === 'done') {
              doneFired = true;
              callbacks.onDone?.();
              return;
            }
          }
          // schema 验证失败的残留 buffer 数据直接忽略，不当 code 处理
        } catch {
          // 非 JSON 残留 buffer 忽略
        }
      }
    }
  } finally {
    reader.releaseLock();
    // 兜底：如果整个流程中 onDone/onError 都未被触发，确保 onDone 被调用
    // 防止 UI 永远卡在 "正在生成" 状态
    if (!doneFired) {
      callbacks.onDone?.();
    }
  }
}
