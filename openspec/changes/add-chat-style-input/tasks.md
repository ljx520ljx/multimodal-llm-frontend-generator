# Tasks: add-chat-style-input

## Task Groups

任务按依赖关系分组，同组内可并行执行。

---

## Group A: 状态管理扩展

### A1. 扩展 useProjectStore 支持对话历史

**依赖**: 无

1. [x] 添加 `chatMessages: ChatMessage[]` 状态
2. [x] 添加 `addUserMessage(text, images?)` action
3. [x] 添加 `addAssistantMessage(content)` action
4. [x] 添加 `updateLastAssistantMessage(content)` 用于流式更新
5. [x] 在 `reset()` 中清空对话历史

**验收**: Store 对话历史可正常增删改 ✅

---

## Group B: 核心组件开发（可并行）

### B1. 创建 UnifiedInput 组件

**依赖**: 无

1. [x] 创建 `src/components/upload/UnifiedInput.tsx`
2. [x] 实现多行 textarea 输入
3. [x] 监听 paste 事件提取图片
4. [x] 显示粘贴图片的缩略图预览（可删除）
5. [x] 右侧发送按钮
6. [x] Enter 发送，Shift+Enter 换行
7. [x] 支持 disabled 状态

**验收**: 组件可独立工作，粘贴图片可预览 ✅

### B2. 创建 ChatHistory 组件

**依赖**: A1

1. [x] 创建 `src/components/chat/ChatHistory.tsx`
2. [x] 渲染用户消息（支持显示图片缩略图）
3. [x] 渲染 AI 消息（支持流式文字）
4. [x] 自动滚动到底部
5. [x] 空状态显示提示文字

**验收**: 对话历史正确显示，自动滚动 ✅

### B3. 修改 ImageList 支持横向布局

**依赖**: 无

1. [x] 添加 `layout?: 'vertical' | 'horizontal'` prop
2. [x] 实现横向滚动布局
3. [x] 调整卡片尺寸（更紧凑）
4. [x] 保持拖拽排序功能

**验收**: 图片列表可横向显示并支持拖拽排序 ✅

---

## Group C: 面板重构

### C1. 创建 InteractionPanel 统一面板

**依赖**: B1, B2, B3

1. [x] 创建 `src/components/interaction/InteractionPanel.tsx`
2. [x] 整合 ImageDropzone、ImageList、ChatHistory、UnifiedInput
3. [x] 实现发送逻辑：
   - 有图片未生成 → 调用 generate API
   - 已生成代码 → 调用 chat API
4. [x] 处理流式响应更新对话历史

**验收**: 统一面板完整工作 ✅

### C2. 修改 MainLayout 移除右侧 ChatPanel

**依赖**: C1

1. [x] 体验模式：使用 InteractionPanel 替代 UploadPanel
2. [x] 体验模式：移除右侧 ChatPanel
3. [x] 开发模式：调整布局比例
4. [x] 调整左侧面板宽度（加宽至 300-320px）

**验收**: 新布局正确显示，右侧对话区被移除 ✅

### C3. 更新 page.tsx

**依赖**: C2

1. [x] 替换 UploadPanel 为 InteractionPanel
2. [x] 移除 ChatPanel 的条件渲染
3. [x] 调整 props 传递

**验收**: 首页正确使用新组件 ✅

---

## Group D: Hooks 适配

### D1. 修改 useGeneration hook

**依赖**: A1, C1

1. [x] 生成前添加 user message 到对话历史
2. [x] 流式响应时更新 assistant message
3. [x] 支持接收 promptText 参数

**验收**: 生成流程正确更新对话历史 ✅

### D2. 修改 useChat hook

**依赖**: A1, C1

1. [x] 发送前添加 user message
2. [x] 流式响应时更新 assistant message
3. [x] 整合到 InteractionPanel

**验收**: 对话修改流程正确更新历史 ✅

---

## Group E: 清理与测试

### E1. 移除旧组件

**依赖**: C3

1. [x] 删除或标记废弃 `ChatPanel` 组件
2. [x] 删除或标记废弃 `ChatInput` 组件
3. [x] 清理未使用的导入

**验收**: 无未使用代码警告 ✅

### E2. 更新测试

**依赖**: C3

1. [x] 更新/添加 UnifiedInput 单元测试
2. [x] 更新/添加 ChatHistory 单元测试
3. [x] 更新 E2E 测试适配新布局

**验收**: 所有测试通过 ✅

---

## Summary

| Group | Tasks | 描述 | 状态 |
|-------|-------|------|------|
| A | 1 | 状态管理扩展 | ✅ 完成 |
| B | 3 | 核心组件开发（可并行） | ✅ 完成 |
| C | 3 | 面板重构 | ✅ 完成 |
| D | 2 | Hooks 适配 | ✅ 完成 |
| E | 2 | 清理与测试 | ✅ 完成 |

**总计**: 11 个任务 - 全部完成

**关键路径**: A1 → B2 → C1 → C2 → C3 → E2

## Implementation Notes

### 已创建/修改的文件

**新组件**:
- `frontend/src/components/upload/UnifiedInput.tsx` - 统一输入框（支持粘贴图片）
- `frontend/src/components/chat/ChatHistory.tsx` - 对话历史组件
- `frontend/src/components/interaction/InteractionPanel.tsx` - 统一交互面板
- `frontend/src/components/interaction/index.ts` - 导出

**修改的组件**:
- `frontend/src/components/upload/ImageList.tsx` - 添加横向布局支持
- `frontend/src/components/layout/MainLayout.tsx` - 重构为使用 interactionPanel
- `frontend/app/page.tsx` - 使用新的 InteractionPanel

**修改的 Hooks**:
- `frontend/src/lib/hooks/useGeneration.ts` - 支持对话历史
- `frontend/src/lib/hooks/useChat.ts` - 使用 useProjectStore

**修改的 Store**:
- `frontend/src/stores/useProjectStore.ts` - 添加 chatMessages 和相关 actions

**删除的文件**:
- `frontend/src/components/upload/UploadPanel.tsx`
- `frontend/src/components/chat/ChatPanel.tsx`
- `frontend/src/components/chat/ChatInput.tsx`
- `frontend/src/components/chat/MessageList.tsx`
- `frontend/src/stores/useChatStore.ts`
- `frontend/src/__tests__/stores/useChatStore.test.ts`

### 测试结果

- 前端编译: ✅ 成功
- 单元测试: ✅ 28 passed
