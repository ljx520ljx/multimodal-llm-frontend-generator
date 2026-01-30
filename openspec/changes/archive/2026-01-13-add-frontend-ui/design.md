# Design: Phase 3 前端基础 UI

## Context

后端服务（Phase 0-2）已完成，提供以下 API：
- `POST /api/upload` - 图片上传
- `POST /api/generate` - 代码生成（SSE 流式输出）
- `POST /api/chat` - 对话修正（SSE）

**核心定位**：交互原型验证工具，让产品/设计师体验和验证 UI 交互流程。

前端需要构建以预览为核心的界面，与后端 API 集成，实现设计稿到可交互原型的完整工作流。

## Goals / Non-Goals

### Goals
- 实现双模式布局（体验模式 + 开发模式）
- **体验模式**：预览区为主，专注交互体验验证
- **开发模式**：传统三栏布局，便于开发调试
- 支持拖拽上传和排序 UI 设计稿
- 集成 Sandpack 实时预览，支持交互操作
- 代码编辑器可折叠，默认隐藏

### Non-Goals
- 对话修正功能（Phase 7）
- 流程录制/回放功能
- 用户认证和历史记录
- Vue 框架支持
- 移动端完整适配

## Decisions

### 1. 布局架构

**决定**: 采用双模式可切换布局

#### 体验模式（默认）- 产品/设计师使用
```
┌─────────────────────────────────────────────────────────────┐
│  Header                              [体验模式 ▼] [开发模式] │
├─────────────────┬───────────────────────────────────────────┤
│   Upload        │                                           │
│   Panel         │           Preview (Sandpack)              │
│   (侧边栏)      │           全屏交互预览区                   │
│   (~20%)        │           (~80%)                          │
│                 │                                           │
│  [图片列表]     │     [可点击交互的原型界面]                │
│  [生成按钮]     │                                           │
│                 ├───────────────────────────────────────────┤
│                 │   [展开代码 ▾] (折叠状态)                 │
└─────────────────┴───────────────────────────────────────────┘
```

#### 开发模式 - 开发者使用
```
┌─────────────────────────────────────────────────────────────┐
│  Header                              [体验模式] [开发模式 ▼] │
├──────────────┬──────────────────────┬───────────────────────┤
│   Upload     │    Code Editor       │    Preview            │
│   Panel      │    (Monaco)          │    (Sandpack)         │
│   (~20%)     │    (~45%)            │    (~35%)             │
│              │                      │                       │
│  [图片列表]  │  [代码编辑区]        │  [实时预览]           │
│  [生成按钮]  │                      │  [错误提示]           │
└──────────────┴──────────────────────┴───────────────────────┘
```

**理由**:
- 体验模式：让产品/设计师专注于交互体验，不被代码干扰
- 开发模式：开发者可查看和调试代码
- 一键切换，适应不同用户场景
- 使用 CSS Grid 实现，状态控制布局切换

### 2. 组件库方案

**决定**: 自建轻量组件 + Tailwind CSS

**备选方案**:
| 方案 | 优点 | 缺点 |
|------|------|------|
| shadcn/ui | 成熟、可定制 | 引入额外复杂度 |
| 自建组件 | 轻量、可控 | 需要更多开发 |
| Headless UI | 无样式冲突 | 需要自己写样式 |

**理由**:
- 项目组件数量有限（Button, Card, Input 等）
- 自建可完全控制，避免依赖升级问题
- L3 文档已定义组件规范

### 3. 状态管理

**决定**: Zustand

```typescript
interface ProjectStore {
  // 视图状态
  viewMode: 'experience' | 'developer';  // 新增：模式切换
  codeExpanded: boolean;                  // 新增：代码面板展开状态

  // 项目状态
  images: ImageFile[];
  sessionId: string | null;
  generatedCode: string;
  status: 'idle' | 'uploading' | 'generating' | 'completed' | 'error';
  error: string | null;

  // Actions
  setViewMode: (mode: 'experience' | 'developer') => void;
  toggleCodeExpanded: () => void;
  addImages: (files: File[]) => void;
  removeImage: (id: string) => void;
  reorderImages: (fromIndex: number, toIndex: number) => void;
  setSessionId: (id: string) => void;
  setGeneratedCode: (code: string) => void;
  appendCode: (chunk: string) => void;
  setStatus: (status: Status) => void;
  reset: () => void;
}
```

**理由**:
- 比 Redux 轻量
- 无需 Provider 包裹
- TypeScript 友好
- viewMode 支持模式切换持久化（localStorage）

### 4. SSE 流式处理

**决定**: 原生 fetch + ReadableStream

```typescript
async function* streamGeneration(sessionId: string, imageIds: string[]) {
  const response = await fetch('/api/generate', {
    method: 'POST',
    body: JSON.stringify({ session_id: sessionId, image_ids: imageIds }),
  });

  const reader = response.body!.getReader();
  const decoder = new TextDecoder();

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    const chunk = decoder.decode(value);
    // 解析 SSE 格式: data: {...}\n\n
    yield parseSSE(chunk);
  }
}
```

**理由**:
- 无需额外依赖
- 控制粒度细
- 便于错误处理

### 5. 图片拖拽排序

**决定**: @hello-pangea/dnd (react-beautiful-dnd 社区分支)

**理由**:
- react-beautiful-dnd 已不再维护
- @hello-pangea/dnd 是活跃的社区分支
- API 兼容，文档丰富

### 6. Monaco Editor 集成

**决定**: 动态导入 + 加载状态

```typescript
import dynamic from 'next/dynamic';

const CodeEditor = dynamic(
  () => import('@/components/editor/CodeEditor'),
  {
    loading: () => <EditorSkeleton />,
    ssr: false,
  }
);
```

**理由**:
- Monaco 包体积大（~2MB）
- 动态导入减少首屏加载
- SSR 禁用避免服务端渲染问题

## Component Architecture

```
src/
├── app/
│   ├── layout.tsx              # 根布局
│   └── page.tsx                # 主页面
│
├── components/
│   ├── ui/                     # 基础 UI
│   │   ├── Button.tsx
│   │   ├── Card.tsx
│   │   ├── Input.tsx
│   │   ├── Skeleton.tsx
│   │   └── index.ts
│   │
│   ├── editor/                 # 编辑器（可折叠）
│   │   ├── CodeEditor.tsx
│   │   ├── CodePanel.tsx       # 新增：可折叠代码面板
│   │   ├── EditorSkeleton.tsx
│   │   └── index.ts
│   │
│   ├── preview/                # 预览（核心）
│   │   ├── SandpackPreview.tsx
│   │   ├── PreviewError.tsx
│   │   └── index.ts
│   │
│   └── layout/                 # 布局
│       ├── Header.tsx
│       ├── ViewModeToggle.tsx  # 新增：模式切换组件
│       ├── MainLayout.tsx      # 双模式布局容器
│       └── index.ts
│
├── features/
│   ├── upload/                 # 上传功能
│   │   ├── components/
│   │   │   ├── ImageUploader.tsx
│   │   │   ├── ImagePreviewCard.tsx
│   │   │   └── ImageSortableList.tsx
│   │   ├── hooks/
│   │   │   └── useImageUpload.ts
│   │   └── index.ts
│   │
│   └── generation/             # 生成功能
│       ├── components/
│       │   ├── GenerateButton.tsx
│       │   └── GenerationStatus.tsx
│       ├── hooks/
│       │   └── useCodeGeneration.ts
│       └── index.ts
│
├── stores/
│   └── useProjectStore.ts      # Zustand store（含 viewMode）
│
├── lib/
│   ├── api/
│   │   ├── client.ts           # API 客户端
│   │   └── sse.ts              # SSE 工具
│   └── utils.ts                # cn() 等工具
│
└── types/
    └── index.ts                # 类型定义
```

## Data Flow

```
用户上传图片
     ↓
ImageUploader → useProjectStore.addImages()
     ↓
用户点击生成
     ↓
GenerateButton → uploadImages() → setSessionId()
     ↓
useCodeGeneration.generate() → SSE stream
     ↓
appendCode() → Sandpack 预览更新（核心体验）
     ↓
用户在预览区体验交互流程
     ↓
（可选）切换到开发模式查看/编辑代码
```

## User Journey

### 产品/设计师（主要用户）
1. 上传多张 UI 设计稿（代表不同状态）
2. 点击"生成原型"
3. 在预览区体验交互（点击按钮、切换状态）
4. 验证流程是否符合预期
5. （可选）展开代码查看实现

### 开发者（次要用户）
1. 切换到开发模式
2. 查看生成的代码
3. 手动调整代码
4. 预览区实时反映修改

## Risks / Trade-offs

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|----------|
| Monaco 加载慢 | 高 | 中 | 动态导入 + Skeleton UI |
| Sandpack 编译失败 | 中 | 中 | 错误边界 + 友好提示 |
| SSE 连接中断 | 低 | 高 | 自动重连 + 进度恢复 |
| 大图片上传慢 | 中 | 低 | 前端压缩 + 进度条 |

## Open Questions

1. **是否需要前端图片压缩？** - 后端已有压缩，前端可暂不实现
2. **多标签页支持？** - 留待后续版本
3. **代码历史版本？** - 留待后续版本
