'use client';

import { useEffect, useRef, useState, useCallback } from 'react';
import {
  SandpackProvider,
  SandpackPreview as SandpackPreviewComponent,
  SandpackConsole,
  useSandpack,
} from '@codesandbox/sandpack-react';
import { useProjectStore } from '@/stores/useProjectStore';
import { useDebouncedValue } from '@/lib/hooks';
import { api, readSSEStream } from '@/lib/api';

const DEFAULT_CODE = `export default function App() {
  return (
    <div className="min-h-screen bg-gray-100 flex items-center justify-center">
      <div className="text-center">
        <h1 className="text-2xl font-bold text-gray-800 mb-4">
          上传设计稿开始生成
        </h1>
        <p className="text-gray-600">
          支持交互预览，点击体验生成的原型
        </p>
      </div>
    </div>
  );
}`;

const DEFAULT_CSS = `@tailwind base;
@tailwind components;
@tailwind utilities;`;

interface PreviewContentProps {
  code: string;
  css: string;
  onRefresh?: () => void;
  onAutoFix?: (errorMessage: string) => void;
  isAutoFixing?: boolean;
  autoFixAttempts?: number;
  isGenerating?: boolean;
}

function PreviewContent({ code, css, onRefresh, onAutoFix, isAutoFixing, autoFixAttempts = 0, isGenerating = false }: PreviewContentProps) {
  const { sandpack } = useSandpack();
  const { error, status } = sandpack;
  const prevCodeRef = useRef(code);
  const prevCssRef = useRef(css);
  const lastErrorRef = useRef<string | null>(null);

  // 同步代码到 Sandpack
  useEffect(() => {
    if (code !== prevCodeRef.current) {
      sandpack.updateFile('/App.tsx', code);
      prevCodeRef.current = code;
    }
    if (css !== prevCssRef.current) {
      sandpack.updateFile('/styles.css', css);
      prevCssRef.current = css;
    }
  }, [code, css, sandpack]);

  // 自动修复编译错误（最多尝试 3 次，生成过程中不触发）
  useEffect(() => {
    if (error && onAutoFix && !isAutoFixing && !isGenerating && autoFixAttempts < 3) {
      const errorMessage = error.message;
      // 避免对同一个错误重复触发
      if (lastErrorRef.current !== errorMessage) {
        lastErrorRef.current = errorMessage;
        // 延迟触发，避免在渲染期间更新状态，增加延迟以确保代码稳定
        const timer = setTimeout(() => {
          onAutoFix(errorMessage);
        }, 1500);
        return () => clearTimeout(timer);
      }
    }
    return undefined;
  }, [error, onAutoFix, isAutoFixing, isGenerating, autoFixAttempts]);

  // 清除错误记录（当代码变化时）
  useEffect(() => {
    lastErrorRef.current = null;
  }, [code]);

  if (error) {
    return (
      <div className="flex h-full flex-col items-center justify-center bg-red-50 p-4">
        {isAutoFixing ? (
          <>
            <svg className="mb-3 h-10 w-10 animate-spin text-blue-500" viewBox="0 0 24 24">
              <circle
                className="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                strokeWidth="4"
                fill="none"
              />
              <path
                className="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
              />
            </svg>
            <p className="text-sm font-medium text-blue-700">正在自动修复...</p>
            <p className="mt-2 text-xs text-blue-600">
              AI 正在分析并修复编译错误 ({autoFixAttempts + 1}/3)
            </p>
          </>
        ) : autoFixAttempts >= 3 ? (
          <>
            <svg
              className="mb-3 h-10 w-10 text-orange-400"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
            >
              <circle cx="12" cy="12" r="10" />
              <path d="M12 8v4M12 16h.01" />
            </svg>
            <p className="text-sm font-medium text-orange-700">自动修复失败</p>
            <p className="mt-2 max-w-md text-center text-xs text-orange-600">
              已尝试 3 次自动修复，请尝试手动描述问题或重新生成
            </p>
            <p className="mt-1 max-w-md text-center text-xs text-red-500">
              {error.message}
            </p>
          </>
        ) : (
          <>
            <svg
              className="mb-3 h-10 w-10 text-red-400"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
            >
              <circle cx="12" cy="12" r="10" />
              <path d="M12 8v4M12 16h.01" />
            </svg>
            <p className="text-sm font-medium text-red-700">编译错误</p>
            <p className="mt-2 max-w-md text-center text-xs text-red-600">
              {error.message}
            </p>
            {onRefresh && !isAutoFixing && (
              <button
                onClick={onRefresh}
                className="mt-3 rounded bg-red-100 px-3 py-1 text-xs text-red-700 hover:bg-red-200"
              >
                重新加载
              </button>
            )}
          </>
        )}
      </div>
    );
  }

  return (
    <div className="relative h-full w-full flex flex-col" style={{ minHeight: '200px' }}>
      {status === 'idle' && (
        <div className="absolute inset-0 flex flex-col items-center justify-center bg-white/90 z-10">
          <svg className="h-8 w-8 animate-spin text-blue-500 mb-3" viewBox="0 0 24 24">
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
              fill="none"
            />
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
            />
          </svg>
          <p className="text-sm text-slate-600">正在启动预览...</p>
        </div>
      )}
      <div className="flex-1 min-h-0">
        <SandpackPreviewComponent
          showNavigator={false}
          showRefreshButton={false}
          showOpenInCodeSandbox={false}
          style={{ height: '100%', width: '100%' }}
        />
      </div>
    </div>
  );
}

interface SandpackPreviewProps {
  showConsole?: boolean;
  onRefresh?: () => void;
  isGenerating?: boolean;
}

// 从响应中提取代码，移除 markdown 代码块标记和 thinking 标签
function extractCodeFromResponse(content: string): string | null {
  let cleaned = content;

  // 1. 首先移除 thinking 标签及其内容
  cleaned = cleaned.replace(/<thinking>[\s\S]*?<\/thinking>/gi, '');
  // 移除可能残留的不完整 thinking 标签
  cleaned = cleaned.replace(/<\/?thinking>/gi, '');

  // 2. 尝试匹配 markdown 代码块
  const pattern = /```[\w]*\s*\n?([\s\S]*?)```/g;
  const matches: RegExpExecArray[] = [];
  let match: RegExpExecArray | null;
  while ((match = pattern.exec(cleaned)) !== null) {
    matches.push(match);
  }

  let code: string;

  if (matches.length > 0) {
    // 返回最后一个代码块的内容
    const lastMatch = matches[matches.length - 1];
    code = lastMatch[1] || '';
  } else {
    // 没有代码块标记，尝试从 import/export 开始提取
    const importIndex = cleaned.indexOf('import ');
    const exportIndex = cleaned.indexOf('export ');
    const startIndex = Math.min(
      ...[importIndex, exportIndex].filter(i => i >= 0)
    );

    if (startIndex >= 0 && startIndex < cleaned.length) {
      code = cleaned.substring(startIndex);
    } else {
      code = cleaned;
    }
  }

  // 3. 清理可能残留的代码块标记
  code = code
    .replace(/^```[\w]*\s*\n?/gm, '')
    .replace(/\n?```\s*$/gm, '')
    .trim();

  // 4. 再次确保没有 thinking 标签残留
  code = code.replace(/<\/?thinking>/gi, '');

  // 5. 确保代码以有效的 React 组件开始
  if (!code || code.length < 10) {
    return null;
  }

  // 6. 验证代码包含 React 特征
  const hasReactFeatures =
    code.includes('import') ||
    code.includes('export') ||
    code.includes('function') ||
    code.includes('const');

  if (!hasReactFeatures) {
    return null;
  }

  return code;
}

export function SandpackPreview({ showConsole = false, onRefresh, isGenerating = false }: SandpackPreviewProps) {
  const generatedCode = useProjectStore((state) => state.generatedCode);
  const generatedCss = useProjectStore((state) => state.generatedCss);
  const sessionId = useProjectStore((state) => state.sessionId);
  const updateCode = useProjectStore((state) => state.updateCode);
  const addAssistantMessage = useProjectStore((state) => state.addAssistantMessage);
  const appendToLastAssistantMessage = useProjectStore((state) => state.appendToLastAssistantMessage);

  const [isAutoFixing, setIsAutoFixing] = useState(false);
  const [autoFixAttempts, setAutoFixAttempts] = useState(0);

  const rawCode = generatedCode?.code || DEFAULT_CODE;
  const rawCss = generatedCss || DEFAULT_CSS;

  // 使用防抖值，避免频繁重渲染
  const debouncedCode = useDebouncedValue(rawCode, 300);
  const debouncedCss = useDebouncedValue(rawCss, 300);

  // 使用 sessionId 和 timestamp 确保每次代码更新都会重新挂载
  const codeTimestamp = generatedCode?.timestamp || 0;
  const sandpackKey = `${sessionId || 'default'}-${codeTimestamp}`;

  // 当代码成功更新后，重置修复尝试次数
  useEffect(() => {
    setAutoFixAttempts(0);
  }, [generatedCode?.timestamp]);

  // 获取 store 的 updateLastAssistantMessage
  const updateLastAssistantMessage = useProjectStore((state) => state.updateLastAssistantMessage);

  // 自动修复编译错误
  const handleAutoFix = useCallback(async (errorMessage: string) => {
    // 生成过程中不触发自动修复
    if (!sessionId || isAutoFixing || isGenerating) return;

    setIsAutoFixing(true);
    setAutoFixAttempts((prev) => prev + 1);

    // 添加"正在修复"的消息
    addAssistantMessage('🔄 AI 正在自动修复编译错误...');

    try {
      const fixPrompt = `代码编译出现错误，请修复：\n\n错误信息：${errorMessage}\n\n请直接输出修复后的完整代码，不需要解释。`;
      const response = await api.chat(sessionId, fixPrompt);

      let codeBuffer = '';
      let hasThinking = false;

      await readSSEStream(response, {
        onThinking: (content) => {
          // 第一次收到 thinking 内容时，更新消息
          if (!hasThinking) {
            hasThinking = true;
            updateLastAssistantMessage('🔄 AI 正在分析错误...\n\n' + content);
          } else {
            appendToLastAssistantMessage(content);
          }
        },
        onCode: (content) => {
          // 代码只存入 buffer，不追加到对话
          codeBuffer += content;
        },
        onDone: () => {
          const extractedCode = extractCodeFromResponse(codeBuffer);
          if (extractedCode) {
            updateCode(extractedCode);
            updateLastAssistantMessage('✅ 代码已自动修复，请查看预览');
          } else if (codeBuffer) {
            updateLastAssistantMessage('⚠️ 修复代码提取失败，请重试');
          } else {
            updateLastAssistantMessage('⚠️ 未能生成修复代码');
          }
          setIsAutoFixing(false);
        },
        onError: (err) => {
          updateLastAssistantMessage(`❌ 修复失败: ${err}`);
          setIsAutoFixing(false);
        },
      });
    } catch (error) {
      const message = error instanceof Error ? error.message : '网络错误';
      updateLastAssistantMessage(`❌ 修复请求失败: ${message}`);
      setIsAutoFixing(false);
    }
  }, [sessionId, isAutoFixing, isGenerating, addAssistantMessage, updateLastAssistantMessage, appendToLastAssistantMessage, updateCode]);

  // 准备 Sandpack 初始文件
  const files = {
    '/App.tsx': debouncedCode,
    '/index.tsx': `import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import './styles.css';

const root = ReactDOM.createRoot(document.getElementById('root')!);
root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);`,
    '/styles.css': debouncedCss,
  };

  return (
    <SandpackProvider
      key={sandpackKey}
      template="react-ts"
      theme="light"
      files={files}
      customSetup={{
        dependencies: {
          'react': '^18.2.0',
          'react-dom': '^18.2.0',
        },
      }}
      options={{
        externalResources: [
          'https://cdn.tailwindcss.com',
        ],
        recompileMode: 'delayed',
        recompileDelay: 300,
      }}
    >
      <div className="flex h-full w-full flex-col" style={{ minHeight: '300px' }}>
        <div className={showConsole ? 'h-[70%]' : 'flex-1'} style={{ minHeight: '200px' }}>
          <PreviewContent
            code={debouncedCode}
            css={debouncedCss}
            onRefresh={onRefresh}
            onAutoFix={sessionId ? handleAutoFix : undefined}
            isAutoFixing={isAutoFixing}
            autoFixAttempts={autoFixAttempts}
            isGenerating={isGenerating}
          />
        </div>
        {showConsole && (
          <div className="h-[30%] border-t border-slate-200">
            <SandpackConsole />
          </div>
        )}
      </div>
    </SandpackProvider>
  );
}
