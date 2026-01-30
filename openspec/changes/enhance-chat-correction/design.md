# Design: enhance-chat-correction

## 架构概览

```
┌─────────────────────────────────────────────────────────┐
│                    前端 (Next.js)                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ ChatPanel    │  │ ChatInput    │  │ MessageList  │  │
│  │ (容器组件)    │  │ (输入框)     │  │ (消息列表)    │  │
│  └──────┬───────┘  └──────────────┘  └──────────────┘  │
│         │                                               │
│         ▼                                               │
│  ┌──────────────────────────────────────────────────┐  │
│  │              useChatStore (Zustand)               │  │
│  │  - messages: Message[]                            │  │
│  │  - isLoading: boolean                             │  │
│  │  - sendMessage(text) → 调用 API                   │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
                           │
                           │ SSE Stream
                           ▼
┌─────────────────────────────────────────────────────────┐
│                    后端 (Go/Gin)                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ ChatHandler  │→ │GenerateService│→ │PromptBuilder │  │
│  │ POST /chat   │  │  .Chat()      │  │ .BuildChat() │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
│                           │                             │
│                           ▼                             │
│                    ┌──────────────┐                     │
│                    │ SessionStore │                     │
│                    │ - Code       │                     │
│                    │ - History    │                     │
│                    └──────────────┘                     │
└─────────────────────────────────────────────────────────┘
```

## 数据流

1. **用户输入** → ChatInput 组件捕获
2. **发送请求** → useChatStore.sendMessage() 调用 `/api/chat`
3. **SSE 流式响应** → 后端返回 thinking/code 事件
4. **实时更新** → 前端更新消息列表 + 代码编辑器 + 预览

## 组件设计

### ChatPanel (容器组件)

```typescript
interface ChatPanelProps {
  sessionId: string;
  onCodeUpdate: (code: string) => void;
}
```

**职责**：
- 管理对话状态
- 协调子组件
- 处理 SSE 响应

### ChatInput (输入组件)

```typescript
interface ChatInputProps {
  onSend: (message: string) => void;
  disabled: boolean;
  placeholder?: string;
}
```

**功能**：
- 文本输入框
- 发送按钮
- Enter 键发送
- 加载状态禁用

### MessageList (消息列表)

```typescript
interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  type: 'text' | 'thinking' | 'code';
  timestamp: Date;
}

interface MessageListProps {
  messages: Message[];
  isLoading: boolean;
}
```

## 后端增强

### 现有能力（无需修改）

- `/api/chat` 接口已存在
- `ChatModifyPromptV2` Prompt 模板已存在
- SessionStore 已支持历史记录

### 可选优化

1. **增量修改 Prompt 优化**：明确"只修改指定部分"的指令
2. **历史压缩**：超过 N 轮后压缩历史减少 token 消耗

## 状态管理

```typescript
// useChatStore.ts
interface ChatState {
  messages: Message[];
  isLoading: boolean;
  error: string | null;

  sendMessage: (sessionId: string, message: string) => Promise<void>;
  clearMessages: () => void;
  addMessage: (message: Message) => void;
}
```

## 错误处理

| 错误场景 | 处理方式 |
|----------|----------|
| 网络中断 | 显示重试按钮 |
| 无代码可修改 | 提示"请先生成代码" |
| LLM 超时 | 显示超时提示，支持重试 |

## 技术决策

### 为什么用 Zustand 而不是 Context？

- 对话状态需要跨组件共享
- 避免不必要的重渲染
- 与现有 `useProjectStore` 保持一致

### 为什么不做 diff 可视化？

- 增加复杂度
- 当前阶段聚焦核心功能
- 可作为后续迭代
