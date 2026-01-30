# Capability: LLM Gateway

统一的大语言模型网关，封装多个 LLM 提供商的 API 调用，支持流式输出和多模态输入。

支持的提供商：
- **国际**：OpenAI、Google Gemini、Anthropic Claude
- **国内**：DeepSeek（深度求索）、Doubao（豆包/火山引擎）

## ADDED Requirements

### Requirement: LLM Gateway Interface

系统 SHALL 提供统一的 LLMGateway 接口，屏蔽不同 LLM 提供商的实现差异。

#### Scenario: Interface definition

- **WHEN** 调用 LLMGateway 接口
- **THEN** 支持 ChatStream 方法发送请求
- **AND** 返回流式响应通道

#### Scenario: Provider identification

- **WHEN** 调用 Provider() 方法
- **THEN** 返回当前使用的提供商标识（如 "openai", "deepseek", "doubao"）

### Requirement: Multi-modal Input Support

LLM Gateway SHALL 支持多模态输入，包括文本和图片。

#### Scenario: Text-only input

- **WHEN** ChatRequest 只包含文本内容
- **THEN** 正确发送到 LLM API
- **AND** 返回流式文本响应

#### Scenario: Image and text input

- **WHEN** ChatRequest 包含图片（Base64）和文本
- **THEN** 正确编码并发送到 LLM API
- **AND** LLM 能够分析图片内容

#### Scenario: Multiple images input

- **WHEN** ChatRequest 包含多张图片
- **THEN** 按顺序发送所有图片
- **AND** LLM 能够分析图片间的差异

### Requirement: Streaming Response

LLM Gateway SHALL 支持流式响应输出，实现实时反馈。

#### Scenario: Stream content chunks

- **WHEN** LLM 开始生成响应
- **THEN** 通过 channel 逐块返回内容
- **AND** 每个 chunk 包含增量文本

#### Scenario: Stream completion

- **WHEN** LLM 完成响应生成
- **THEN** 发送 Done 类型的 chunk
- **AND** 关闭响应 channel

#### Scenario: Stream error handling

- **WHEN** 流式响应过程中发生错误
- **THEN** 发送 Error 类型的 chunk
- **AND** 包含错误详情

### Requirement: Retry Mechanism

LLM Gateway SHALL 实现重试机制，处理临时性错误。

#### Scenario: Rate limit retry

- **WHEN** 收到 HTTP 429 (Rate Limited) 响应
- **THEN** 使用指数退避策略重试
- **AND** 解析 Retry-After 头部（如果存在）

#### Scenario: Server error retry

- **WHEN** 收到 HTTP 5xx 服务器错误
- **THEN** 使用指数退避策略重试
- **AND** 最多重试配置的次数

#### Scenario: Non-retryable error

- **WHEN** 收到 HTTP 400/401/403 错误
- **THEN** 不进行重试
- **AND** 立即返回错误

#### Scenario: Max retries exceeded

- **WHEN** 重试次数超过最大限制
- **THEN** 返回最后一次错误
- **AND** 包含重试次数信息

### Requirement: OpenAI Integration

系统 SHALL 集成 OpenAI GPT-4o API，支持完整的多模态功能。

#### Scenario: OpenAI text generation

- **WHEN** 使用 OpenAI provider 发送文本请求
- **THEN** 调用 OpenAI Chat Completion API
- **AND** 返回流式文本响应

#### Scenario: OpenAI image analysis

- **WHEN** 使用 OpenAI provider 发送图片请求
- **THEN** 使用 GPT-4o 的视觉能力
- **AND** 返回图片分析结果

#### Scenario: OpenAI API key configuration

- **WHEN** 配置 OPENAI_API_KEY 环境变量
- **THEN** OpenAI Gateway 使用该密钥认证

### Requirement: DeepSeek Integration

系统 SHALL 集成 DeepSeek API，通过 OpenAI 兼容层实现。

#### Scenario: DeepSeek text generation

- **WHEN** 使用 DeepSeek provider 发送文本请求
- **THEN** 调用 DeepSeek API（OpenAI 兼容格式）
- **AND** 返回流式文本响应

#### Scenario: DeepSeek configuration

- **WHEN** 配置 DEEPSEEK_API_KEY 和 DEEPSEEK_MODEL 环境变量
- **THEN** DeepSeek Gateway 使用配置的密钥和模型
- **AND** 默认模型为 deepseek-chat

#### Scenario: DeepSeek API endpoint

- **WHEN** 使用 DeepSeek provider
- **THEN** 请求发送到 https://api.deepseek.com
- **AND** 使用 OpenAI 兼容的 API 格式

### Requirement: Doubao Integration

系统 SHALL 集成 Doubao（豆包）API，通过火山引擎 OpenAI 兼容层实现。

#### Scenario: Doubao text generation

- **WHEN** 使用 Doubao provider 发送文本请求
- **THEN** 调用火山引擎 API（OpenAI 兼容格式）
- **AND** 返回流式文本响应

#### Scenario: Doubao configuration

- **WHEN** 配置 DOUBAO_API_KEY 和 DOUBAO_MODEL 环境变量
- **THEN** Doubao Gateway 使用配置的密钥和端点 ID
- **AND** DOUBAO_MODEL 格式为 ep-xxx-xxx

#### Scenario: Doubao API endpoint

- **WHEN** 使用 Doubao provider
- **THEN** 请求发送到 https://ark.cn-beijing.volces.com/api/v3
- **AND** 使用 OpenAI 兼容的 API 格式

### Requirement: Gateway Factory

系统 SHALL 提供 Gateway 工厂函数，根据配置创建对应的 LLM Gateway 实例。

#### Scenario: Create OpenAI gateway

- **WHEN** LLM_PROVIDER 配置为 "openai"
- **THEN** 工厂返回 OpenAI Gateway 实例

#### Scenario: Create DeepSeek gateway

- **WHEN** LLM_PROVIDER 配置为 "deepseek"
- **THEN** 工厂返回 DeepSeek Gateway 实例（使用 OpenAI 兼容层）

#### Scenario: Create Doubao gateway

- **WHEN** LLM_PROVIDER 配置为 "doubao"
- **THEN** 工厂返回 Doubao Gateway 实例（使用 OpenAI 兼容层）

#### Scenario: Create Gemini gateway

- **WHEN** LLM_PROVIDER 配置为 "gemini"
- **THEN** 工厂返回 Gemini Gateway 实例（占位实现）

#### Scenario: Create Anthropic gateway

- **WHEN** LLM_PROVIDER 配置为 "anthropic"
- **THEN** 工厂返回 Anthropic Gateway 实例（占位实现）

#### Scenario: Invalid provider

- **WHEN** LLM_PROVIDER 配置为不支持的值
- **THEN** 工厂返回错误
- **AND** 错误信息包含支持的提供商列表

### Requirement: Request Timeout

LLM Gateway SHALL 支持请求超时控制。

#### Scenario: Context timeout

- **WHEN** 请求的 context 超时
- **THEN** 取消正在进行的 API 调用
- **AND** 返回超时错误

#### Scenario: Configurable timeout

- **WHEN** 配置 LLM_TIMEOUT 环境变量
- **THEN** 使用配置的超时时间
- **AND** 默认值为 5 分钟
