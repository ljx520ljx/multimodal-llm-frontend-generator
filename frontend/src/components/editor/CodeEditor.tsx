'use client';

import { useCallback } from 'react';
import dynamic from 'next/dynamic';
import type { OnMount } from '@monaco-editor/react';
import { useProjectStore } from '@/stores/useProjectStore';
import { Skeleton } from '@/components/ui';

const MonacoEditor = dynamic(() => import('@monaco-editor/react'), {
  ssr: false,
  loading: () => (
    <div className="flex h-full items-center justify-center bg-slate-900">
      <Skeleton width={200} height={24} className="bg-slate-700" />
    </div>
  ),
});

interface CodeEditorProps {
  readOnly?: boolean;
}

export function CodeEditor({ readOnly = false }: CodeEditorProps) {
  const generatedCode = useProjectStore((state) => state.generatedCode);
  const generatedCss = useProjectStore((state) => state.generatedCss);
  const activeFile = useProjectStore((state) => state.activeFile);
  const updateCode = useProjectStore((state) => state.updateCode);
  const updateCss = useProjectStore((state) => state.updateCss);

  const isTypescript = activeFile === 'App.tsx';
  const code = isTypescript ? (generatedCode?.code || '') : (generatedCss || '');
  const language = isTypescript ? 'typescript' : 'css';

  const handleChange = useCallback(
    (value: string | undefined) => {
      if (value !== undefined) {
        if (isTypescript) {
          updateCode(value);
        } else {
          updateCss(value);
        }
      }
    },
    [isTypescript, updateCode, updateCss]
  );

  // 配置 Monaco Editor 的 TypeScript/JSX 支持
  const handleMount: OnMount = useCallback((_editor, monaco) => {
    // 配置 TypeScript 编译器选项
    monaco.languages.typescript.typescriptDefaults.setCompilerOptions({
      jsx: monaco.languages.typescript.JsxEmit.React,
      jsxFactory: 'React.createElement',
      jsxFragmentFactory: 'React.Fragment',
      target: monaco.languages.typescript.ScriptTarget.ESNext,
      moduleResolution: monaco.languages.typescript.ModuleResolutionKind.NodeJs,
      allowSyntheticDefaultImports: true,
      esModuleInterop: true,
    });

    // 添加 React 类型声明
    monaco.languages.typescript.typescriptDefaults.addExtraLib(
      `declare module 'react' {
        export function useState<T>(initialState: T | (() => T)): [T, (value: T | ((prev: T) => T)) => void];
        export function useEffect(effect: () => void | (() => void), deps?: any[]): void;
        export function useCallback<T extends Function>(callback: T, deps: any[]): T;
        export function useRef<T>(initialValue: T): { current: T };
        export function useMemo<T>(factory: () => T, deps: any[]): T;
        export const Fragment: any;
        export default any;
      }`,
      'react.d.ts'
    );
  }, []);

  return (
    <div className="h-full w-full">
      <MonacoEditor
        height="100%"
        language={language}
        theme="vs-dark"
        value={code}
        path={activeFile}
        onChange={readOnly ? undefined : handleChange}
        onMount={handleMount}
        options={{
          readOnly,
          readOnlyMessage: { value: '当前为预览模式，如需修改代码请在对话框中描述您的需求' },
          minimap: { enabled: false },
          fontSize: 14,
          lineNumbers: 'on',
          scrollBeyondLastLine: false,
          wordWrap: 'on',
          automaticLayout: true,
          tabSize: 2,
          formatOnPaste: true,
          formatOnType: true,
        }}
      />
    </div>
  );
}
