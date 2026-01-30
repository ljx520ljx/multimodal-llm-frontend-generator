# Design: 交互预览 & 编辑器增强

## 架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                          Frontend                                │
│  ┌─────────────┐  ┌─────────────┐  ┌───────────────────────────┐ │
│  │ UploadPanel │  │ EditorPanel │  │      PreviewPanel         │ │
│  │   (图片)     │──│  (Monaco)   │──│      (Sandpack)           │ │
│  └──────┬──────┘  └──────┬──────┘  └───────────┬───────────────┘ │
│         │                │                      │                 │
│         ▼                ▼                      ▼                 │
│  ┌──────────────────────────────────────────────────────────────┐ │
│  │                  Zustand Store (useProjectStore)              │ │
│  │  images[] │ generatedCode │ status │ thinkingContent         │ │
│  └──────────────────────────────────────────────────────────────┘ │
│         │                                                         │
│         ▼                                                         │
│  ┌──────────────────────────────────────────────────────────────┐ │
│  │                     API Client (lib/api)                      │ │
│  │  upload() ──► POST /api/upload                                │ │
│  │  generate() ─► POST /api/generate (SSE)                       │ │
│  └──────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Backend (Go)                             │
│  POST /api/upload ──► 图片处理 ──► imageIds                      │
│  POST /api/generate ─► LLM Gateway ──► SSE (thinking/code)       │
└─────────────────────────────────────────────────────────────────┘
```

## 核心技术决策

### 1. 编辑器-预览联动

**方案**：通过 Zustand store 单向数据流

```
CodeEditor onChange ──► updateCode() ──► generatedCode ──► SandpackPreview
                          │
                       debounce(300ms)
```

**原因**：
- 避免频繁重渲染 Sandpack
- 保持状态集中管理
- 便于后续添加撤销/重做功能

### 2. Sandpack 文件同步

**问题**：Sandpack 不会自动响应外部 props 变化

**方案**：使用 `useSandpack()` hook 的 `updateFile()` 方法

```typescript
const { sandpack } = useSandpack();
useEffect(() => {
  sandpack.updateFile('/App.tsx', code);
}, [code]);
```

### 3. 上传流程

```
用户选择图片
     │
     ▼
addImages() ──► store.images[]
     │
     ▼
点击"生成交互原型"
     │
     ▼
setStatus('uploading')
     │
     ▼
api.upload(files) ──► POST /api/upload
     │
     ▼
setStatus('generating')
     │
     ▼
api.generate(imageIds) ──► SSE stream
     │
     ├─► onThinking ──► appendThinking()
     │
     └─► onCode ──► setGeneratedCode()
```

### 4. 预览体验优化

**加载状态**：
- 初始化时显示骨架屏
- 编译中显示 loading overlay
- 使用 `SandpackPreviewRef` 获取编译状态

**错误处理**：
- 语法错误：显示错误行号和提示
- 运行时错误：显示错误堆栈
- 提供"恢复上次有效代码"按钮

**全屏模式**：
- 使用 `dialog` 元素实现模态全屏
- 保留刷新和关闭按钮
- ESC 键退出

## 文件变更清单

### 新增文件
- `src/components/preview/PreviewToolbar.tsx` - 预览工具栏
- `src/components/preview/FullscreenPreview.tsx` - 全屏预览
- `src/components/editor/EditorToolbar.tsx` - 编辑器工具栏
- `src/lib/hooks/useDebouncedCode.ts` - 防抖 Hook
- `.env.local` - 环境变量

### 修改文件
- `src/components/preview/SandpackPreview.tsx` - 添加文件同步
- `src/components/preview/PreviewPanel.tsx` - 添加工具栏和全屏
- `src/components/editor/CodeEditor.tsx` - 优化配置
- `src/components/editor/EditorPanel.tsx` - 添加工具栏
- `src/components/upload/UploadPanel.tsx` - 对接后端上传
- `src/lib/hooks/useGeneration.ts` - 完善生成流程
- `src/stores/useProjectStore.ts` - 添加 imageIds 状态

## 性能考虑

1. **Sandpack 重渲染**：使用 debounce 避免频繁更新
2. **Monaco 动态加载**：保持 SSR false 配置
3. **图片处理**：前端压缩超大图片（>5MB）后再上传
