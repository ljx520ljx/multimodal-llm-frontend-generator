# chat-interaction Specification

## Purpose
TBD - created by archiving change add-backend-services. Update Purpose after archive.
## Requirements
### Requirement: 自然语言代码修改

系统 **SHALL** 支持用户通过自然语言描述来修改已生成的代码。

#### Scenario: 修改样式

**Given** 用户已生成代码，会话中存在有效代码
**And** session_id 有效
**When** 用户发送 POST /api/chat 请求，message 为 "把按钮改成蓝色"
**Then** 系统返回 SSE 流式响应
**And** LLM 基于当前代码和用户指令生成修改后的代码
**And** 返回 type="code" 事件包含更新后的完整代码

#### Scenario: 添加功能

**Given** 用户已生成代码
**When** 用户发送 POST /api/chat 请求，message 为 "添加一个搜索框"
**Then** 系统分析当前代码结构
**And** 生成包含搜索框的新代码
**And** 保持原有功能不变

#### Scenario: 修复问题

**Given** 用户已生成代码
**When** 用户发送 POST /api/chat 请求，message 为 "点击按钮时没有反应，帮我修复"
**Then** 系统分析当前代码的事件处理逻辑
**And** 生成修复后的代码
**And** 返回 type="thinking" 事件解释问题原因

#### Scenario: 会话无代码

**Given** 用户的会话中没有生成过代码
**When** 用户发送 POST /api/chat 请求
**Then** 系统返回 400 Bad Request
**And** 响应包含错误信息 "No code generated yet, please generate code first"

### Requirement: 对话历史管理

系统 **MUST** 维护对话历史，以支持连续的多轮修改。

#### Scenario: 累积对话上下文

**Given** 用户已进行过两轮对话修改
**When** 用户发送第三轮修改请求
**Then** Prompt 包含之前两轮的对话历史
**And** LLM 可以理解完整的修改上下文

#### Scenario: 历史长度限制

**Given** 对话历史超过 10 轮
**When** 用户发送新的修改请求
**Then** 系统保留最近 10 轮对话历史
**And** 较早的历史被移除以控制 token 消耗

### Requirement: 代码更新

系统 **SHALL** 在每次成功的对话修改后更新会话中的代码。

#### Scenario: 更新代码并记录历史

**Given** LLM 成功生成修改后的代码
**When** 对话完成
**Then** 新代码替换会话中的旧代码
**And** 用户消息被添加到对话历史 (role="user")
**And** 生成的代码被添加到对话历史 (role="assistant", type="code")

