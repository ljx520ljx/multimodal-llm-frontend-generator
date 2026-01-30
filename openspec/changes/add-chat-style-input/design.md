# Design: add-chat-style-input

## Architecture Overview

### Before (当前布局)
```
┌──────────┬────────────────────────┬──────────┐
│  上传区  │      预览区             │  对话区  │
│  (左侧)  │      (中间)             │  (右侧)  │
└──────────┴────────────────────────┴──────────┘
```

### After (新布局)
```
┌──────────────────┬───────────────────────────┐
│  统一交互面板     │        预览区             │
│  (上传+对话)      │        (中间)             │
└──────────────────┴───────────────────────────┘
```

## New Panel Structure

```
┌─────────────────────────────────────────────┐
│ 设计稿上传                                   │
│ 上传 UI 设计稿序列，系统将自动推断交互逻辑    │
├─────────────────────────────────────────────┤
│ ┌─────────────────────────────────────────┐ │
│ │     拖拽或点击上传设计稿                  │ │
│ │     支持 PNG, JPG, GIF, WebP            │ │
│ └─────────────────────────────────────────┘ │
├─────────────────────────────────────────────┤
│ 已上传 3 张图片                   清空全部   │
│ ┌─────┐ ┌─────┐ ┌─────┐                    │
│ │ 1   │ │ 2   │ │ 3   │  (横向排列，可拖拽) │
│ └─────┘ └─────┘ └─────┘                    │
├─────────────────────────────────────────────┤
│ 对话历史 (flex-1, 可滚动)                    │
│ ┌─────────────────────────────────────────┐ │
│ │ 👤 [图片x3] 生成一个电商首页              │ │
│ │ 🤖 正在分析设计稿...                     │ │
│ │ 🤖 已生成原型代码                        │ │
│ │ 👤 把按钮改成蓝色                        │ │
│ │ 🤖 已更新代码                           │ │
│ └─────────────────────────────────────────┘ │
├─────────────────────────────────────────────┤
│ ┌──────────────────────────────┐ ┌───────┐ │
│ │ 输入需求，支持粘贴图片...     │ │  ➤   │ │
│ │ [粘贴的图片预览]             │ │ 发送  │ │
│ └──────────────────────────────┘ └───────┘ │
└─────────────────────────────────────────────┘
```

## Component Design

### 1. UnifiedInput (统一输入组件)

```typescript
interface UnifiedInputProps {
  onSend: (text: string, images: File[]) => void;
  disabled?: boolean;
  placeholder?: string;
  buttonText?: string; // "生成原型" 或 "发送"
}
```

**功能特性：**
- 多行文本输入 (textarea)
- 监听 paste 事件提取图片
- 粘贴图片的预览 + 删除
- 右侧发送按钮
- Enter 发送，Shift+Enter 换行

### 2. ChatHistory (对话历史组件)

```typescript
interface ChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  images?: string[]; // 图片预览 URL
  timestamp: number;
}

interface ChatHistoryProps {
  messages: ChatMessage[];
}
```

### 3. 重构 UploadPanel → InteractionPanel

```typescript
// 新的统一交互面板
function InteractionPanel() {
  return (
    <div className="flex h-full flex-col">
      {/* 标题 */}
      <Header />

      {/* 拖拽上传区 */}
      <ImageDropzone />

      {/* 已上传图片列表 (横向) */}
      <ImageList layout="horizontal" />

      {/* 对话历史 (flex-1) */}
      <ChatHistory />

      {/* 统一输入框 */}
      <UnifiedInput />
    </div>
  );
}
```

## State Management Changes

```typescript
interface ProjectState {
  // 新增
  chatMessages: ChatMessage[];
  addUserMessage: (text: string, images?: string[]) => void;
  addAssistantMessage: (content: string) => void;

  // 移除 (整合到 chatMessages)
  // thinkingContent: string; → 改为 assistant message
}
```

## Interaction Flow

### 首次生成
1. 用户拖拽/粘贴图片 → 添加到图片列表
2. 用户输入需求描述（可选）
3. 点击发送 →
   - 添加 user message (带图片缩略图)
   - 调用 generate API
   - 流式显示 assistant message

### 对话修改
1. 用户在输入框输入修改指令
2. 点击发送 →
   - 添加 user message
   - 调用 chat API
   - 流式显示 assistant message

## Layout Changes

### MainLayout 修改

```typescript
// 体验模式：移除右侧 chatPanel
if (viewMode === 'experience') {
  return (
    <div className="flex h-full">
      <aside className="w-[280px]">  {/* 加宽左侧 */}
        {interactionPanel}  {/* 原 uploadPanel + chatPanel */}
      </aside>
      <main className="flex-1">
        {previewPanel}
      </main>
    </div>
  );
}
```

## Image List Layout Change

当前：垂直列表
新设计：横向滚动列表（节省空间）

```
当前:
┌────────────┐
│ 1  [img]   │
├────────────┤
│ 2  [img]   │
├────────────┤
│ 3  [img]   │
└────────────┘

新设计:
┌────┐ ┌────┐ ┌────┐ →
│ 1  │ │ 2  │ │ 3  │
└────┘ └────┘ └────┘
```

## Testing Strategy

1. **单元测试**：UnifiedInput 组件、ChatHistory 组件
2. **集成测试**：粘贴图片 → 添加到列表 → 发送生成
3. **E2E 测试**：完整的上传 → 生成 → 对话修改流程
