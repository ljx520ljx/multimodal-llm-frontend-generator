# Tasks: Phase 1 LLM Gateway 开发

## 1. Gateway 接口定义

- [x] 1.1 创建 `backend/internal/gateway/types/prompt.go` - 定义 ChatRequest, Message, ContentPart 结构
- [x] 1.2 创建 `backend/internal/gateway/types/chunk.go` - 定义 StreamChunk 结构
- [x] 1.3 创建 `backend/internal/gateway/types/errors.go` - 定义错误类型
- [x] 1.4 创建 `backend/internal/gateway/interface.go` - 定义 LLMGateway 接口

## 2. OpenAI 兼容层实现

- [x] 2.1 添加 `go-openai` 依赖到 go.mod
- [x] 2.2 创建 `backend/internal/gateway/openai_compatible.go` - 实现 OpenAI 兼容 Gateway
- [x] 2.3 实现 `ChatStream` 方法 - 支持流式输出
- [x] 2.4 实现多模态输入处理 - 支持图片 Base64 编码
- [x] 2.5 支持自定义 BaseURL（适配 DeepSeek/Doubao）
- [x] 2.6 编写 OpenAI 兼容层单元测试

## 3. 重试与错误处理

- [x] 3.1 创建 `backend/internal/gateway/retry.go` - 实现重试逻辑
- [x] 3.2 实现指数退避算法
- [x] 3.3 处理 429 限流错误 - 解析 Retry-After 头
- [x] 3.4 处理超时错误 - 可配置超时时间
- [x] 3.5 编写重试逻辑单元测试

## 4. Gateway 工厂

- [x] 4.1 创建 `backend/internal/gateway/factory.go` - 实现工厂函数
- [x] 4.2 支持 OpenAI provider (openai)
- [x] 4.3 支持 DeepSeek provider (deepseek) - 复用 OpenAI 兼容层
- [x] 4.4 支持 Doubao provider (doubao) - 复用 OpenAI 兼容层
- [x] 4.5 预留 Gemini Gateway 接口（占位实现）
- [x] 4.6 预留 Anthropic Gateway 接口（占位实现）

## 5. 配置更新

- [x] 5.1 更新 `backend/internal/config/config.go` - 添加所有 LLM 提供商配置
- [x] 5.2 更新 `backend/.env.example` - 添加 DeepSeek 和 Doubao 配置示例
- [x] 5.3 添加 LLM_TIMEOUT 配置支持

## 6. 集成与验证

- [x] 6.1 创建测试脚本验证流式输出
- [x] 6.2 单元测试：Gateway 工厂 - 验证所有 provider 创建
- [x] 6.3 单元测试：重试机制 - 验证指数退避和错误处理
- [ ] 6.4 手动测试：DeepSeek - 验证 API 连通性（如有 API Key）
- [ ] 6.5 手动测试：Doubao - 验证 API 连通性（如有 API Key）

## Dependencies

- 任务组 1 必须先完成（其他任务依赖接口定义）
- 任务组 2 和 3 可以并行
- 任务组 4 依赖任务组 2
- 任务组 5 可以与任务组 2-4 并行
- 任务组 6 依赖所有其他任务

## Validation

```bash
# 运行单元测试
cd backend && go test -v ./internal/gateway/...

# 手动测试 OpenAI（需要设置环境变量）
export LLM_PROVIDER=openai
export OPENAI_API_KEY=sk-xxx
cd backend && go run cmd/server/main.go

# 手动测试 DeepSeek（需要设置环境变量）
export LLM_PROVIDER=deepseek
export DEEPSEEK_API_KEY=sk-xxx
cd backend && go run cmd/server/main.go

# 手动测试 Doubao（需要设置环境变量）
export LLM_PROVIDER=doubao
export DOUBAO_API_KEY=xxx
export DOUBAO_MODEL=ep-xxx-xxx
cd backend && go run cmd/server/main.go
```

## Completion Notes

### 已完成的工作

1. **Gateway 类型定义** (`backend/internal/gateway/types/`)
   - `prompt.go`: ChatRequest, Message, ContentPart 结构，支持多模态输入
   - `chunk.go`: StreamChunk 结构，支持流式响应 (content/error/done)
   - `errors.go`: GatewayError 错误类型，支持可重试错误检测

2. **LLMGateway 接口** (`backend/internal/gateway/interface.go`)
   - `ChatStream(ctx, req) (<-chan StreamChunk, error)` - 流式聊天
   - `Provider() string` - 获取提供商名称

3. **OpenAI 兼容层** (`backend/internal/gateway/openai_compatible.go`)
   - 统一实现支持 OpenAI、DeepSeek、Doubao 三个提供商
   - 支持自定义 BaseURL 配置
   - 支持多模态输入（文本 + 图片 Base64）
   - 流式响应通过 Go channel 返回

4. **重试机制** (`backend/internal/gateway/retry.go`)
   - 指数退避算法 (可配置 InitialBackoff, MaxBackoff, Multiplier)
   - Jitter 随机抖动避免惊群效应
   - 支持 Retry-After 头解析
   - 可重试错误检测 (429, 500, 502, 503, 504)

5. **Gateway 工厂** (`backend/internal/gateway/factory.go`)
   - 支持 5 个 provider: openai, deepseek, doubao, gemini, anthropic
   - OpenAI/DeepSeek/Doubao 使用 OpenAI 兼容层
   - Gemini/Anthropic 使用占位实现 (待后续开发)
   - 可选重试包装器

6. **配置更新** (`backend/internal/config/config.go`, `backend/.env.example`)
   - 支持所有 5 个 LLM 提供商的 API Key 和 Model 配置
   - LLM_TIMEOUT 超时配置

### 测试结果

```
=== RUN   TestNewGateway_OpenAI        PASS
=== RUN   TestNewGateway_DeepSeek      PASS
=== RUN   TestNewGateway_Doubao        PASS
=== RUN   TestNewGateway_DoubaoMissingModel PASS
=== RUN   TestNewGateway_InvalidProvider PASS
=== RUN   TestNewGateway_MissingAPIKey/OpenAI PASS
=== RUN   TestNewGateway_MissingAPIKey/DeepSeek PASS
=== RUN   TestNewGateway_MissingAPIKey/Doubao PASS
=== RUN   TestNewGateway_WithRetry     PASS
=== RUN   TestNewGateway_Gemini_Placeholder PASS
=== RUN   TestNewGateway_Anthropic_Placeholder PASS
=== RUN   TestRetryer_CalculateBackoff PASS
=== RUN   TestRetryer_Do_Success       PASS
=== RUN   TestRetryer_Do_NonRetryableError PASS
=== RUN   TestRetryer_Do_RetryableError PASS
=== RUN   TestRetryer_Do_EventualSuccess PASS
=== RUN   TestRetryer_Do_ContextCanceled PASS
=== RUN   TestIsRetryableStatusCode    PASS

ok  multimodal-llm-frontend-generator/internal/gateway  0.777s
```

### 待手动验证（需要真实 API Key）

- [ ] OpenAI API 连通性测试
- [ ] DeepSeek API 连通性测试
- [ ] Doubao API 连通性测试
