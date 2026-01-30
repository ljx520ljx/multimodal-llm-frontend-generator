# L3 - 组件库文档 | Components

> DocOps 层级: **L3 (Implementation Level)**
> 父文档: [L2 前端模块文档](../../CLAUDE.md)

## 模块职责

提供可复用的 UI 组件，包括基础 UI 组件、编辑器组件、预览组件和对话组件。

## 目录结构

```
components/
├── ui/                         # 基础 UI 组件
│   ├── Button.tsx
│   ├── Card.tsx
│   ├── ThinkingSteps.tsx      # 思考步骤展示组件
│   ├── PreviewSkeleton.tsx    # 预览加载骨架
│   └── index.ts               # 导出汇总
│
├── editor/                     # 编辑器组件
│   ├── CodeEditor.tsx         # Monaco Editor 封装
│   ├── EditorToolbar.tsx      # 编辑器工具栏
│   ├── CodePanel.tsx          # 代码面板
│   └── index.ts
│
├── preview/                    # 预览组件
│   ├── HtmlPreview.tsx        # HTML iframe 预览（当前使用）
│   ├── SandpackPreview.tsx    # Sandpack 预览（备用）
│   ├── PreviewPanel.tsx       # 预览面板容器
│   ├── PreviewToolbar.tsx     # 预览工具栏
│   ├── FullscreenPreview.tsx  # 全屏预览
│   └── index.ts
│
├── chat/                       # 对话组件
│   └── ChatHistory.tsx        # 对话历史展示
│
├── upload/                     # 上传组件
│   ├── ImageDropzone.tsx      # 图片拖拽上传
│   ├── ImageList.tsx          # 图片列表
│   └── UnifiedInput.tsx       # 统一输入框
│
├── interaction/                # 交互面板
│   └── InteractionPanel.tsx   # 左侧交互区域
│
└── CLAUDE.md                   # 本文档
```

## 核心组件

### HtmlPreview（当前使用）

使用 iframe 直接渲染 HTML + Alpine.js 代码：

```typescript
// components/preview/HtmlPreview.tsx
export function HtmlPreview({ onRefresh, isGenerating }: HtmlPreviewProps) {
  const iframeRef = useRef<HTMLIFrameElement>(null);
  const code = useProjectStore((state) => state.generatedCode?.code);

  // 构建完整 HTML（注入 Tailwind + Alpine.js CDN）
  const finalHtml = useMemo(() => buildFullHtml(code), [code]);

  // 使用 srcdoc 渲染
  useEffect(() => {
    if (iframeRef.current && finalHtml) {
      iframeRef.current.srcdoc = finalHtml;
    }
  }, [finalHtml]);

  return <iframe ref={iframeRef} className="h-full w-full" />;
}
```

**优势**：
- 无编译延迟，即时渲染
- 稳定性高，避免 Sandpack 编译问题
- 支持 Alpine.js 交互

### ThinkingSteps

展示 AI 分析步骤，支持流式显示：

```typescript
// components/ui/ThinkingSteps.tsx
export function ThinkingSteps({ content, isStreaming, variant }: ThinkingStepsProps) {
  const { steps, otherContent } = parseThinkingContent(content);

  return (
    <div className="space-y-3">
      {steps.map((step) => (
        <div key={step.number} className="flex gap-2.5">
          <StepIcon number={step.number} isComplete={!isStreaming} />
          <div>
            <span className="font-medium">{step.title}</span>
            <p>{step.content}</p>
          </div>
        </div>
      ))}
    </div>
  );
}
```

**支持的步骤格式**：
- `### Step X: Title`
- `**Step X: Title**`
- `Step X: Title`（无标记）

### ChatHistory

展示对话历史，自动识别思考内容：

```typescript
// components/chat/ChatHistory.tsx
function parseContent(content: string): ParsedContent {
  // 检测 Step 格式内容作为思考过程
  const hasStepContent = /Step\s*\d+[：:]/i.test(content);

  if (hasStepContent) {
    // 分离思考内容和状态消息
    return { thinkingContent, mainContent, hasCode };
  }
  // ...
}
```

**功能**：
- 自动识别 `Step X:` 格式的分析内容
- 流式显示思考步骤
- 过滤代码块，只显示文字说明

## 预览方案对比

| 方案 | 组件 | 优势 | 劣势 |
|------|------|------|------|
| **HTML + iframe** | `HtmlPreview` | 稳定、无编译、即时 | 功能相对简单 |
| React + Sandpack | `SandpackPreview` | 功能丰富 | 编译慢、可能报错 |

当前默认使用 `HtmlPreview`，`SandpackPreview` 保留作为备用。

## 组件规范

### 组件模板

```typescript
// components/ui/Button.tsx
import { type ButtonHTMLAttributes, forwardRef } from 'react';
import { cn } from '@/lib/utils';

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
  loading?: boolean;
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'primary', size = 'md', loading, children, ...props }, ref) => {
    return (
      <button
        ref={ref}
        className={cn(
          'inline-flex items-center justify-center rounded-md font-medium',
          // ... variant and size styles
          className
        )}
        {...props}
      >
        {loading && <Spinner className="mr-2 h-4 w-4" />}
        {children}
      </button>
    );
  }
);
```

### 样式约定

- 使用 Tailwind CSS 类名
- 使用 `cn()` 工具函数合并类名
- 支持 `className` prop 覆盖样式
- 响应式设计使用 Tailwind 断点

## 更新历史

- **2026-01-15**:
  - 添加 `HtmlPreview` 组件，替代 Sandpack
  - 优化 `ThinkingSteps` 支持多种步骤格式
  - 优化 `ChatHistory` 自动识别思考内容
- **2025-xx-xx**: 初始版本

## 相关文档

- [Prompt 模板文档](../../../backend/pkg/prompt/CLAUDE.md)
- [useGeneration Hook](../lib/hooks/useGeneration.ts)
