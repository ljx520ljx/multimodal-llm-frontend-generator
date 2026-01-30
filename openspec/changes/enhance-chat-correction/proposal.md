# Proposal: enhance-chat-correction

## Summary

实现自然语言修改代码功能，用户通过对话方式修改生成的前端代码，代码和预览实时更新。

## Motivation

**当前状态**：
- 后端 `/api/chat` 接口已实现基础功能
- `ChatModifyPromptV2` Prompt 模板已存在
- 前端缺少对话面板 UI
- 上下文管理需要完善

**目标**：用户发送"把按钮改成蓝色"后，代码和预览能实时更新。

## Scope

### In Scope
1. **对话上下文管理**：维护对话历史，支持多轮修改
2. **增量修改 Prompt**：优化 Prompt 以支持精准的增量修改
3. **POST /api/chat 接口**：增强现有接口（如需要）
4. **对话面板 UI**：聊天输入框、消息列表、发送按钮
5. **代码增量更新**：实时更新编辑器和预览

### Out of Scope
- 代码 diff 可视化
- 语音输入

## Deliverables

1. `frontend/src/features/chat/` - 对话功能模块
2. `frontend/src/components/chat/` - 对话面板组件
3. 后端上下文管理增强（如需要）
4. 单元测试

## Acceptance Criteria

- [ ] 发送"把按钮改成蓝色"，代码中按钮颜色变为蓝色
- [ ] 代码和预览实时更新
- [ ] 支持多轮连续修改
