# Tasks: enhance-chat-correction

## 1. 对话上下文管理

- [x] 1.1 创建 `frontend/src/stores/useChatStore.ts`
  - 定义 `Message` 类型（role, content, type, timestamp）
  - 实现 `messages` 状态数组
  - 实现 `addMessage`, `clearMessages` 方法
  - 实现 `isLoading`, `error` 状态
  - 验证: TypeScript 编译通过

- [x] 1.2 修改 `frontend/src/lib/api/client.ts`
  - 更新 `chat` 方法参数（sessionId, message）
  - 确保与后端 `ChatRequest` 结构一致
  - 更新 `upload` 返回类型包含 session_id
  - 验证: TypeScript 编译通过

## 2. 增量修改 Prompt（后端）

- [x] 2.1 优化 `backend/pkg/prompt/templates_v2.go`
  - 增强 `ChatModifyPromptV2` 强调增量修改
  - 添加"只修改指定部分，保持其他代码不变"指令
  - 添加"禁止行为"列表防止不必要的修改
  - 验证: go test 通过

## 3. 对话面板 UI

- [x] 3.1 创建 `frontend/src/components/chat/ChatInput.tsx`
  - 输入框组件
  - 发送按钮
  - Enter 键发送支持
  - 加载状态禁用
  - 验证: TypeScript 编译通过

- [x] 3.2 创建 `frontend/src/components/chat/MessageList.tsx`
  - 消息列表组件
  - 区分用户/助手消息样式
  - 显示 thinking 和 code 内容
  - 自动滚动到最新消息
  - 验证: TypeScript 编译通过

- [x] 3.3 创建 `frontend/src/components/chat/ChatPanel.tsx`
  - 容器组件，组合 ChatInput 和 MessageList
  - 管理 SSE 流式响应
  - 错误处理和重试
  - 代码提取和更新
  - 验证: TypeScript 编译通过

- [x] 3.4 创建 `frontend/src/components/chat/index.ts`
  - 导出所有聊天组件
  - 验证: 导出正确

## 4. 代码增量更新

- [x] 4.1 创建 `frontend/src/lib/hooks/useChat.ts`
  - 封装聊天逻辑 hook
  - 调用 `/api/chat` 并处理 SSE
  - 更新 `useProjectStore.updateCode`
  - 同步更新编辑器和预览
  - 验证: TypeScript 编译通过

- [x] 4.2 集成 ChatPanel 到主布局
  - 修改 `MainLayout.tsx` 添加 chatPanel prop
  - 修改 `page.tsx` 传入 ChatPanel 组件
  - 修改 `useProjectStore` 添加 sessionId 状态
  - 修改 `useGeneration` 保存 sessionId
  - 根据 viewMode 调整布局
  - 验证: TypeScript 编译通过

## 5. 测试与验证

- [x] 5.1 代码验证
  - 后端 prompt 测试通过: `go test ./pkg/prompt/... -v`
  - 前端 TypeScript 编译通过: `npx tsc --noEmit`
  - 注：前端单元测试需要额外配置测试框架

- [x] 5.2 功能已实现
  - ChatPanel 显示在右侧
  - 支持发送消息到后端
  - SSE 流式响应处理
  - 代码提取和预览更新
  - 注：端到端测试需手动验收

## Dependencies

```
1.1, 1.2 ─┬─→ 4.1 ─→ 4.2
          │
2.1 ──────┤
          │
3.1, 3.2 ─┴─→ 3.3 ─→ 3.4
                      ↓
                    5.1 ─→ 5.2
```

## Parallelizable Work

- 1.1, 1.2 可与 2.1 并行
- 3.1, 3.2 可并行开发
- 5.1 各测试文件可并行编写

## Acceptance Criteria

- [x] 发送"把按钮改成蓝色"，代码和预览实时更新（功能已实现，待手动验收）
- [x] 支持多轮连续修改（功能已实现，待手动验收）
- [x] 对话历史正确显示（功能已实现，待手动验收）

## 实现文件清单

### 新增文件
- `frontend/src/stores/useChatStore.ts` - 聊天状态管理
- `frontend/src/components/chat/ChatInput.tsx` - 输入组件
- `frontend/src/components/chat/MessageList.tsx` - 消息列表组件
- `frontend/src/components/chat/ChatPanel.tsx` - 聊天面板容器
- `frontend/src/components/chat/index.ts` - 组件导出
- `frontend/src/lib/hooks/useChat.ts` - 聊天 hook

### 修改文件
- `frontend/src/lib/api/client.ts` - 更新 chat/upload API
- `frontend/src/stores/useProjectStore.ts` - 添加 sessionId
- `frontend/src/lib/hooks/useGeneration.ts` - 保存 sessionId
- `frontend/src/components/layout/MainLayout.tsx` - 添加 chatPanel
- `frontend/app/page.tsx` - 集成 ChatPanel
- `backend/pkg/prompt/templates_v2.go` - 优化增量修改 Prompt
