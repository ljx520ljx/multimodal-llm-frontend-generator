# Change: Phase 1 LLM Gateway 开发

## Why

项目需要与多个 LLM 提供商进行集成，以实现多模态图片分析和代码生成功能。支持的提供商包括：

**国际提供商**：
- OpenAI GPT-4o
- Google Gemini
- Anthropic Claude

**国内提供商**：
- DeepSeek（深度求索）
- Doubao（豆包 - 字节跳动/火山引擎）

为了保持代码的可维护性和可扩展性，需要建立统一的 LLM 网关抽象层，屏蔽不同提供商的 API 差异。

## What Changes

### 新增 LLM Gateway 模块
- 定义统一的 `LLMGateway` 接口
- 定义 `LLMPrompt` 结构（支持文本+图片多模态输入）
- 定义 `StreamChunk` 结构（流式响应）

### OpenAI 兼容层实现
- 集成 `go-openai` SDK
- 实现 `ChatStream` 方法支持流式输出
- 支持 GPT-4o 多模态输入（图片 Base64）
- **复用 OpenAI 兼容层支持 DeepSeek 和 Doubao**（它们的 API 兼容 OpenAI 格式）

### 重试与错误处理
- 实现指数退避重试机制
- 处理 429 限流错误
- 处理超时和网络错误

### Gateway 工厂
- 根据配置动态选择 LLM 提供商
- 支持自定义 API Endpoint（适配国内提供商）
- 预留 Gemini 和 Anthropic 接口（后续实现）

## Impact

- **Affected specs**: 新增 `llm-gateway` capability
- **Affected code**:
  - `backend/internal/gateway/` - 新增 interface.go, openai.go, factory.go, retry.go
  - `backend/internal/gateway/types/` - 新增类型定义
  - `backend/go.mod` - 新增 go-openai 依赖
  - `backend/internal/config/` - 新增 LLM 提供商配置

## Acceptance Criteria

- 单元测试覆盖核心逻辑
- 手动测试：发送文本+图片 Prompt，收到流式响应
- 支持通过配置切换不同 LLM 提供商（包括国内提供商）
- 重试机制正确处理临时错误
