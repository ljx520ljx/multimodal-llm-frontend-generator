# chat-interaction Specification

## Purpose

对话交互规范定义了用户通过自然语言修改已生成代码的能力。支持普通对话修改、标记修改（元素选择器）和原始设计稿参考。

**相关规范**: [agent-core](../agent-core/spec.md) - ChatAgent 实现和工具调用详情
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

### Requirement: Tool Calling 支持

系统 **SHALL** 在对话修改过程中支持 Tool Calling，用于代码验证和交互检查。

详细规范参见 [agent-core](../agent-core/spec.md#requirement-chat-tool---validate_html)

#### Scenario: 修改后自动验证

**Given** 用户请求修改代码
**When** ChatAgent 生成修改后的代码
**Then** ChatAgent MAY 调用 validate_html 验证语法
**And** 如果验证失败，自动修复并重新验证

### Requirement: 多轮工具调用循环

系统 **SHALL** 支持多轮工具调用循环，由 LLM 自主判定是否需要继续调用工具。

详细规范参见 [agent-core](../agent-core/spec.md#requirement-multi-turn-tool-calling-loop)

#### Scenario: 验证失败后自动修复

**Given** ChatAgent 调用 validate_html 返回验证失败
**When** LLM 分析错误信息
**Then** LLM 自动修复代码
**And** 再次调用 validate_html 验证
**And** 重复直到通过或达到最大次数 (5)

#### Scenario: 循环终止条件

**Given** ChatAgent 正在进行多轮工具调用
**When** 以下任一条件满足:
  - LLM 不再调用任何工具
  - 工具调用次数达到 MAX_TOOL_ITERATIONS (5)
**Then** ChatAgent 输出最终代码并结束

### Requirement: 原始设计稿参考

系统 **SHALL** 在对话修改时自动附带原始设计稿图片供 LLM 参考。

#### Scenario: 自动附带原图

**Given** 用户会话中存在上传的设计稿图片 (session.Images)
**When** 用户发送对话修改请求
**Then** 系统自动从 session 获取原图
**And** 将图片附加到 LLM 请求中
**And** LLM 可参考原图进行精修

#### Scenario: 无原图时纯文本处理

**Given** 用户会话中没有设计稿图片
**When** 用户发送对话修改请求
**Then** 系统发送纯文本请求给 LLM
**And** LLM 仅基于当前代码和用户指令进行修改

