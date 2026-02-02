# agent-core Specification

## Purpose

Python Agent Core 提供多 Agent 协作框架，支持首次生成 Pipeline 和对话修改 ChatAgent。ChatAgent 采用单 Agent + Tool Calling + 多轮循环模式，实现快速响应的对话修改能力。

## Requirements

### Requirement: ChatAgent

系统 **SHALL** 提供 ChatAgent，支持通过对话修改生成的代码。

ChatAgent 输入:
- session_id: 会话标识
- message: 用户指令（普通对话或标记修改）
- current_code: 当前代码
- images: 原始设计稿图片（后端从 session 自动附带）
- history: 对话历史（可选）

ChatAgent 能力:
- 支持 Tool Calling（validate_html, check_interaction）
- 支持多轮工具调用循环（LLM 自主判定）
- 支持原始设计稿图片上下文
- 支持标记修改消息格式

#### Scenario: Normal chat modification

- **WHEN** 用户发送 "把按钮改成蓝色"
- **AND** current_code 包含红色按钮
- **THEN** ChatAgent 输出修改后的代码
- **AND** 按钮颜色变为蓝色

#### Scenario: Marked element modification

- **WHEN** 用户发送标记修改消息 "请修改这个元素（元素: button, 类名: bg-red-500）：改成蓝色"
- **THEN** ChatAgent 理解元素定位
- **AND** 精确修改指定元素

#### Scenario: With original design images

- **WHEN** 请求包含原始设计稿图片
- **AND** 用户要求 "参照设计稿调整间距"
- **THEN** ChatAgent 参考图片进行修改

---

### Requirement: Chat Tool - validate_html

系统 **SHALL** 提供 `validate_html` 工具，用于验证 HTML 代码语法正确性。

工具定义:
- 名称: validate_html
- 参数: code (string) - 要验证的 HTML 代码
- 返回: { valid: boolean, errors: string[] }

#### Scenario: Valid HTML

- **WHEN** 调用 validate_html 传入有效 HTML 代码
- **THEN** 返回 { valid: true, errors: [] }

#### Scenario: Invalid HTML

- **WHEN** 调用 validate_html 传入包含未闭合标签的 HTML
- **THEN** 返回 { valid: false, errors: ["未闭合的标签: div"] }

---

### Requirement: Chat Tool - check_interaction

系统 **SHALL** 提供 `check_interaction` 工具，用于检查状态机交互完整性。

工具定义:
- 名称: check_interaction
- 参数: code (string) - 要检查的 HTML 代码
- 返回: { complete: boolean, missing_states: string[], invalid_transitions: string[] }

#### Scenario: Complete interactions

- **WHEN** 调用 check_interaction 传入状态机完整的代码
- **THEN** 返回 { complete: true, missing_states: [], invalid_transitions: [] }

#### Scenario: Missing state definition

- **WHEN** 调用 check_interaction 传入缺少状态定义的代码
- **THEN** 返回 { complete: false, missing_states: ["checkout"], invalid_transitions: [] }

---

### Requirement: Multi-Turn Tool Calling Loop

系统 **SHALL** 支持多轮工具调用循环，由 LLM 自主判定是否继续调用工具。

配置:
- MAX_TOOL_ITERATIONS: 5（防止无限循环）
- 工具执行超时: 5s

#### Scenario: Tool validation fails then retry

- **WHEN** validate_html 返回 { valid: false, errors: [...] }
- **THEN** ChatAgent **SHALL** 根据错误信息自动修复代码
- **AND** 再次调用 validate_html 验证修复结果
- **AND** 重复直到验证通过或达到 MAX_TOOL_ITERATIONS

#### Scenario: LLM decides no more tools needed

- **WHEN** LLM 在处理过程中不再调用任何工具
- **THEN** ChatAgent **SHALL** 输出最终代码
- **AND** 发送 done 事件

#### Scenario: Max iterations reached

- **WHEN** 工具调用次数达到 MAX_TOOL_ITERATIONS (5)
- **THEN** ChatAgent **SHALL** 停止循环
- **AND** 输出当前最佳结果
- **AND** 发送 done 事件（可能包含警告）

---

### Requirement: Chat SSE Event Types

系统 **SHALL** 支持以下 Chat 相关 SSE 事件类型:

| Event | Data |
|-------|------|
| thinking | {"content": "思考过程..."} |
| tool_call | {"tool": "validate_html", "args": {...}} |
| tool_result | {"tool": "validate_html", "result": {...}} |
| code | {"html": "<!DOCTYPE html>..."} |
| error | {"error": "错误信息"} |
| done | {"success": true} |

#### Scenario: Tool call event flow

- **WHEN** ChatAgent 调用工具
- **THEN** 先发送 tool_call 事件
- **AND** 执行工具
- **AND** 发送 tool_result 事件

#### Scenario: Multi-turn event flow

- **WHEN** ChatAgent 进行多轮工具调用
- **THEN** 每轮依次发送 tool_call → tool_result
- **AND** 可能穿插 thinking 事件
- **AND** 最终发送 code 事件和 done 事件

---

### Requirement: Chat API Endpoint

系统 **SHALL** 提供 `/api/v1/chat` 端点，支持对话修改功能。

请求格式:
```json
{
    "session_id": "string",
    "message": "string",
    "current_code": "string",
    "images": [{"id": "string", "base64": "string", "order": number}],
    "history": [{"role": "string", "content": "string"}]
}
```

响应格式: SSE 流

#### Scenario: Successful chat with tool loop

- **WHEN** POST /api/v1/chat 包含有效请求
- **THEN** 返回 SSE 流
- **AND** 包含 thinking 事件
- **AND** 可能包含多轮 tool_call/tool_result 事件
- **AND** 包含 code 事件
- **AND** 以 done 事件结束

#### Scenario: Missing current_code

- **WHEN** POST /api/v1/chat 缺少 current_code
- **THEN** 返回 422 Validation Error

#### Scenario: Missing message

- **WHEN** POST /api/v1/chat 缺少 message
- **THEN** 返回 422 Validation Error
