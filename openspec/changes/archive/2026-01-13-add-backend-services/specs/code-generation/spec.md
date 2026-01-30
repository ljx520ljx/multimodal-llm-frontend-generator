# Capability: 代码生成

## ADDED Requirements

### Requirement: 流式代码生成

系统 **SHALL** 根据上传的 UI 设计稿图片，调用 LLM 生成前端代码，并以 SSE 流式方式返回。

#### Scenario: 单图生成 React 代码

**Given** 用户已上传一张 UI 设计稿图片
**And** session_id 有效
**When** 用户发送 POST /api/generate 请求，指定 framework 为 "react"
**Then** 系统返回 SSE 流式响应
**And** 首先返回 type="thinking" 的事件，描述分析过程
**And** 然后返回 type="code" 的事件，包含生成的 React 代码
**And** 最后返回 type="done" 的事件

#### Scenario: 多图差异分析生成

**Given** 用户已上传两张 UI 设计稿图片（表示同一页面的两个状态）
**And** session_id 有效
**When** 用户发送 POST /api/generate 请求
**Then** 系统分析两张图片的视觉差异
**And** 推断状态变化和交互逻辑
**And** 生成包含状态管理和事件处理的代码

#### Scenario: 会话不存在

**Given** 用户提供一个不存在的 session_id
**When** 用户发送 POST /api/generate 请求
**Then** 系统返回 404 Not Found
**And** 响应包含错误信息 "Session not found"

#### Scenario: 图片不存在

**Given** 用户提供有效的 session_id
**But** image_ids 中包含不存在的图片 ID
**When** 用户发送 POST /api/generate 请求
**Then** 系统返回 404 Not Found
**And** 响应包含错误信息 "Image not found"

### Requirement: Prompt 构建

系统 **MUST** 根据用户输入构建符合 LLM 输入格式的多模态 Prompt。

#### Scenario: 构建单图 Prompt

**Given** 用户选择了一张图片
**When** PromptService 构建生成 Prompt
**Then** Prompt 包含系统消息（角色设定）
**And** Prompt 包含用户消息（图片 + 生成指令）
**And** 图片以 Base64 data URL 格式嵌入

#### Scenario: 构建多图差异分析 Prompt

**Given** 用户选择了两张图片
**When** PromptService 构建差异分析 Prompt
**Then** Prompt 包含两张图片
**And** Prompt 包含差异分析指令
**And** 指令要求模型描述视觉变化和推断交互逻辑

### Requirement: 代码保存

系统 **SHALL** 保存最新生成的代码到会话中，以支持后续的迭代修改。

#### Scenario: 保存生成的代码

**Given** LLM 成功生成代码
**When** 代码生成完成
**Then** 代码被保存到会话的 code 字段
**And** 会话的 UpdatedAt 时间戳被更新

### Requirement: LLM 错误处理

系统 **MUST** 优雅处理 LLM 调用过程中的各类错误。

#### Scenario: LLM 超时

**Given** LLM 调用超过配置的超时时间（默认 5 分钟）
**When** 超时发生
**Then** 系统返回 SSE 事件 type="error"
**And** 错误消息为 "Generation timeout"

#### Scenario: LLM 限流

**Given** LLM API 返回 429 限流错误
**When** 重试机制耗尽
**Then** 系统返回 SSE 事件 type="error"
**And** 错误消息为 "Rate limited, please try again later"
