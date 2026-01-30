# L2 - 前端模块文档 | Frontend Module

> DocOps 层级: **L2 (Module Level)**
> 父文档: [L1 项目级文档](../CLAUDE.md)

## 模块概述

基于 Next.js 14+ 的**交互原型验证平台**，让产品/设计师快速体验和验证 UI 交互流程。

**核心定位**：交互体验优先，对话式修改。

**布局设计**：
- **交互面板** (左侧 320px)：图片上传、对话输入、对话历史
- **预览面板** (右侧 flex-1)：HTML + Alpine.js 实时预览

## 架构设计

```
frontend/
├── src/
│   ├── app/                    # Next.js App Router
│   │   ├── layout.tsx         # 根布局
│   │   ├── page.tsx           # 首页入口
│   │   └── globals.css        # 全局样式
│   │
│   ├── components/             # 通用组件 (L3)
│   │   ├── ui/                # 基础 UI 组件
│   │   ├── layout/            # 布局组件
│   │   │   ├── Header.tsx
│   │   │   └── MainLayout.tsx      # 主布局（交互+预览）
│   │   ├── editor/            # 编辑器组件
│   │   │   ├── CodeEditor.tsx
│   │   │   └── CodePanel.tsx
│   │   ├── preview/           # 预览组件（核心）
│   │   │   ├── HtmlPreview.tsx     # HTML 预览（当前使用）
│   │   │   └── SandpackPreview.tsx # Sandpack 预览（备用）
│   │   ├── upload/            # 上传组件
│   │   ├── chat/              # 对话组件
│   │   └── interaction/       # 交互面板
│   │
│   ├── features/               # 功能模块 (L3)
│   │   ├── upload/            # 图片上传模块
│   │   ├── generation/        # 原型生成模块
│   │   └── chat/              # 对话修正模块
│   │
│   ├── lib/                    # 工具库 (L3)
│   │   ├── api/               # API 客户端
│   │   ├── hooks/             # 自定义 Hooks
│   │   └── utils/             # 工具函数
│   │
│   ├── stores/                 # 状态管理
│   │   └── useProjectStore.ts # Zustand Store
│   │
│   └── types/                  # TypeScript 类型
│       └── index.ts
│
├── public/                     # 静态资源
├── package.json
├── tailwind.config.ts
├── tsconfig.json
└── next.config.mjs
```

## 核心功能模块

### 1. 交互预览 (`components/preview`) - 核心
- **HtmlPreview**: 使用 iframe + srcdoc 渲染 HTML + Alpine.js（当前使用）
- **SandpackPreview**: Sandpack 沙箱预览（备用）
- **实时交互体验**：支持点击、状态切换等交互
- 热更新预览

### 2. 图片上传与排序 (`components/upload`)
- 拖拽上传多张 UI 设计稿
- 图片预览与顺序调整
- 图片压缩与格式转换

### 3. 原型生成 (`features/generation`)
- 调用后端 API 获取生成结果
- SSE 流式输出展示
- 生成进度与思考步骤展示

### 4. 对话修正 (`components/chat`)
- 自然语言输入修改需求
- 对话历史展示
- AI 分析步骤展示（ThinkingSteps）
- 增量代码更新

### 5. 交互面板 (`components/interaction`)
- 统一的左侧交互区域
- 整合上传、对话输入、对话历史

## 技术规范

### 依赖清单

| 依赖 | 版本 | 用途 |
|------|------|------|
| next | 14.x | React 框架 |
| react | 18.x | UI 库 |
| typescript | 5.x | 类型系统 |
| tailwindcss | 3.x | 样式方案 |
| @monaco-editor/react | latest | 代码编辑器 |
| @codesandbox/sandpack-react | latest | 代码预览沙箱 |
| @hello-pangea/dnd | latest | 拖拽排序 |
| zustand | latest | 状态管理 |
| zod | latest | 数据验证 |

### 组件设计原则

```typescript
// 1. 函数式组件 + TypeScript
interface ComponentProps {
  title: string;
  onAction: () => void;
}

export function Component({ title, onAction }: ComponentProps) {
  // 2. Hooks 在顶部声明
  const [state, setState] = useState<string>('');

  // 3. 事件处理函数使用 handle 前缀
  const handleClick = () => {
    onAction();
  };

  // 4. 条件渲染使用早返回
  if (!title) return null;

  return (
    <div className="..." onClick={handleClick}>
      {title}
    </div>
  );
}
```

### 状态管理模式

```typescript
// Zustand Store 模式
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface ProjectState {
  // 会话状态
  sessionId: string | null;

  // 图片状态
  images: ImageFile[];

  // 生成状态
  generatedCode: GeneratedCode | null;
  thinkingContent: string;
  status: 'idle' | 'uploading' | 'generating' | 'completed' | 'error';

  // 对话历史
  chatMessages: ChatMessage[];

  // Actions
  setSessionId: (id: string) => void;
  addImage: (image: ImageFile) => void;
  setGeneratedCode: (code: GeneratedCode) => void;
  addUserMessage: (text: string) => void;
  addAssistantMessage: (content: string) => void;
}

export const useProjectStore = create<ProjectState>()(
  persist(
    (set) => ({
      sessionId: null,
      images: [],
      generatedCode: null,
      thinkingContent: '',
      status: 'idle',
      chatMessages: [],
      setSessionId: (id) => set({ sessionId: id }),
      addImage: (image) => set((state) => ({
        images: [...state.images, image]
      })),
      setGeneratedCode: (code) => set({ generatedCode: code, status: 'completed' }),
      // ... 其他 actions
    }),
    { name: 'project-store', partialize: () => ({}) }
  )
);
```

### API 调用模式

```typescript
// lib/api/client.ts
const API_BASE = process.env.NEXT_PUBLIC_API_URL;

export async function generateCode(images: File[]): Promise<ReadableStream> {
  const formData = new FormData();
  images.forEach((img, i) => formData.append(`image_${i}`, img));

  const response = await fetch(`${API_BASE}/api/generate`, {
    method: 'POST',
    body: formData,
  });

  if (!response.ok) throw new Error('Generation failed');
  return response.body!;
}
```

## 文件命名规范

| 类型 | 命名规则 | 示例 |
|------|----------|------|
| 组件 | PascalCase | `ImageUploader.tsx` |
| Hook | camelCase + use 前缀 | `useImageUpload.ts` |
| 工具函数 | camelCase | `formatCode.ts` |
| 类型定义 | PascalCase | `ImageFile.ts` |
| 常量 | UPPER_SNAKE_CASE | `API_ENDPOINTS.ts` |

## 测试策略

```typescript
// __tests__/components/ImageUploader.test.tsx
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ImageUploader } from '../ImageUploader';

describe('ImageUploader', () => {
  it('should upload images on drop', async () => {
    const onUpload = jest.fn();
    render(<ImageUploader onUpload={onUpload} />);

    // 模拟拖拽上传
    const dropzone = screen.getByTestId('dropzone');
    // ...测试逻辑
  });
});
```

## 环境变量

```env
# .env.local
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws
```

## 相关 L3 文档

- [组件库文档](./src/components/CLAUDE.md)
- [上传功能文档](./src/features/upload/CLAUDE.md)
- [代码生成功能文档](./src/features/generation/CLAUDE.md)
- [对话修正功能文档](./src/features/chat/CLAUDE.md)
- [API 客户端文档](./src/lib/api/CLAUDE.md)
