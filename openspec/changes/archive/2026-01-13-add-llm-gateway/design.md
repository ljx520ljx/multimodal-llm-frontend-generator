# Design: Phase 1 LLM Gateway 开发

## Context

项目需要与多个 LLM 提供商集成以实现 UI 设计稿到代码的转换。支持的提供商包括：

**国际提供商**：OpenAI、Google Gemini、Anthropic Claude
**国内提供商**：DeepSeek（深度求索）、Doubao（豆包/火山引擎）

不同提供商的 API 格式各异，需要统一抽象层来简化上层业务逻辑。

### 约束条件
- 必须支持流式输出（SSE）以提供实时反馈
- 必须支持多模态输入（文本 + 图片）
- 需要处理 API 限流和临时错误
- 配置驱动，支持运行时切换提供商
- 国内提供商需要支持自定义 API Endpoint

## Goals / Non-Goals

### Goals
- 建立统一的 LLM Gateway 抽象接口
- 实现 OpenAI GPT-4o 的完整集成（流式+多模态）
- 通过 OpenAI 兼容层支持 DeepSeek 和 Doubao
- 实现健壮的重试和错误处理机制
- 预留 Gemini 和 Anthropic 扩展点

### Non-Goals
- 不在本阶段实现 Gemini 和 Anthropic 的完整集成
- 不实现 Prompt 优化和模板管理（Phase 2）
- 不实现对话历史管理（Phase 7）

## Decisions

### D1: Gateway 接口设计

**决定**: 使用 Channel 返回流式数据

```go
type LLMGateway interface {
    // ChatStream 发送请求并返回流式响应通道
    ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, error)

    // Provider 返回提供商标识
    Provider() string
}
```

**理由**:
- Go Channel 是处理流式数据的惯用方式
- 支持 context 取消，便于超时控制
- 调用方可以使用 range 循环简洁消费

**替代方案**: 使用回调函数
- 缺点：回调嵌套，代码可读性差
- 缺点：错误处理复杂

### D2: Prompt 结构设计

**决定**: 使用消息数组 + 内容块模式

```go
type ChatRequest struct {
    Model    string
    Messages []Message
    Options  *ChatOptions
}

type Message struct {
    Role    string        // system, user, assistant
    Content []ContentPart // 支持混合内容
}

type ContentPart struct {
    Type     string // text, image_url
    Text     string
    ImageURL *ImageURL
}

type ImageURL struct {
    URL    string // data:image/jpeg;base64,... 或 https://...
    Detail string // low, high, auto
}
```

**理由**:
- 与 OpenAI API 格式一致，减少转换开销
- 灵活支持文本、图片混合输入
- 可扩展支持其他内容类型

### D3: 流式响应块设计

**决定**: 统一的 StreamChunk 结构

```go
type StreamChunk struct {
    Type    ChunkType // content, error, done
    Content string    // 文本内容（增量）
    Error   error     // 错误信息
}

type ChunkType string
const (
    ChunkTypeContent ChunkType = "content"
    ChunkTypeError   ChunkType = "error"
    ChunkTypeDone    ChunkType = "done"
)
```

**理由**:
- 类型标识便于上层区分处理
- 错误通过 chunk 传递，不中断流
- Done 标识明确流结束

### D4: 重试策略

**决定**: 指数退避 + 抖动

```go
type RetryConfig struct {
    MaxRetries     int           // 最大重试次数，默认 3
    InitialBackoff time.Duration // 初始退避时间，默认 1s
    MaxBackoff     time.Duration // 最大退避时间，默认 30s
    Multiplier     float64       // 退避倍数，默认 2.0
    Jitter         float64       // 抖动比例，默认 0.1
}
```

**可重试错误**:
- HTTP 429 (Rate Limited)
- HTTP 500, 502, 503, 504 (Server Error)
- 网络超时错误

**不可重试错误**:
- HTTP 400 (Bad Request)
- HTTP 401, 403 (Auth Error)
- Context Canceled

**理由**:
- 指数退避避免雪崩效应
- 抖动分散重试请求，减少并发冲突
- 区分可重试/不可重试错误，避免无效重试

### D5: 工厂模式与 OpenAI 兼容层

**决定**: 简单工厂函数 + OpenAI 兼容层复用

```go
func NewGateway(cfg *config.Config) (LLMGateway, error) {
    switch cfg.LLMProvider {
    case "openai":
        return NewOpenAICompatibleGateway(OpenAICompatibleConfig{
            APIKey:   cfg.OpenAIAPIKey,
            Model:    cfg.OpenAIModel,
            BaseURL:  "https://api.openai.com/v1",
            Provider: "openai",
        })
    case "deepseek":
        return NewOpenAICompatibleGateway(OpenAICompatibleConfig{
            APIKey:   cfg.DeepSeekAPIKey,
            Model:    cfg.DeepSeekModel, // 默认 deepseek-chat 或 deepseek-coder
            BaseURL:  "https://api.deepseek.com",
            Provider: "deepseek",
        })
    case "doubao":
        return NewOpenAICompatibleGateway(OpenAICompatibleConfig{
            APIKey:   cfg.DoubaoAPIKey,
            Model:    cfg.DoubaoModel, // ep-xxx-xxx 格式的端点 ID
            BaseURL:  "https://ark.cn-beijing.volces.com/api/v3",
            Provider: "doubao",
        })
    case "gemini":
        return NewGeminiGateway(cfg.GeminiAPIKey)
    case "anthropic":
        return NewAnthropicGateway(cfg.AnthropicAPIKey)
    default:
        return nil, fmt.Errorf("unsupported provider: %s", cfg.LLMProvider)
    }
}
```

**理由**:
- DeepSeek 和 Doubao 都提供 OpenAI 兼容的 API
- 通过配置 BaseURL 和 Model 即可复用 OpenAI 实现
- 减少代码重复，降低维护成本

### D6: 提供商配置结构

**决定**: 统一的配置结构，按提供商分组

```go
type LLMConfig struct {
    Provider string // openai, deepseek, doubao, gemini, anthropic
    Timeout  time.Duration

    // OpenAI
    OpenAIAPIKey string
    OpenAIModel  string // gpt-4o, gpt-4-turbo, etc.

    // DeepSeek
    DeepSeekAPIKey string
    DeepSeekModel  string // deepseek-chat, deepseek-coder

    // Doubao (火山引擎)
    DoubaoAPIKey string
    DoubaoModel  string // ep-xxx-xxx (端点 ID)

    // Gemini
    GeminiAPIKey string

    // Anthropic
    AnthropicAPIKey string
}
```

**环境变量**:
```env
LLM_PROVIDER=openai

# OpenAI
OPENAI_API_KEY=sk-xxx
OPENAI_MODEL=gpt-4o

# DeepSeek
DEEPSEEK_API_KEY=sk-xxx
DEEPSEEK_MODEL=deepseek-chat

# Doubao (火山引擎)
DOUBAO_API_KEY=xxx
DOUBAO_MODEL=ep-xxx-xxx

# Gemini
GEMINI_API_KEY=xxx

# Anthropic
ANTHROPIC_API_KEY=xxx

# 通用
LLM_TIMEOUT=300s
```

## Directory Structure

```
backend/internal/gateway/
├── types/
│   ├── prompt.go      # ChatRequest, Message, ContentPart
│   ├── chunk.go       # StreamChunk, ChunkType
│   └── errors.go      # GatewayError, 错误类型定义
├── interface.go       # LLMGateway 接口定义
├── openai_compatible.go # OpenAI 兼容层（支持 OpenAI/DeepSeek/Doubao）
├── gemini.go          # Gemini 占位实现
├── anthropic.go       # Anthropic 占位实现
├── factory.go         # Gateway 工厂
├── retry.go           # 重试逻辑
└── openai_compatible_test.go # 单元测试
```

## Provider Comparison

| 提供商 | API 格式 | 多模态支持 | 流式支持 | 备注 |
|--------|----------|------------|----------|------|
| OpenAI | OpenAI | ✓ GPT-4o | ✓ | 原生支持 |
| DeepSeek | OpenAI 兼容 | ✓ DeepSeek-VL | ✓ | 需指定 BaseURL |
| Doubao | OpenAI 兼容 | ✓ | ✓ | 火山引擎，需端点 ID |
| Gemini | Google | ✓ | ✓ | 需单独实现 |
| Anthropic | Anthropic | ✓ | ✓ | 需单独实现 |

## Risks / Trade-offs

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| OpenAI API 变更 | 中 | 使用官方 SDK，跟踪版本更新 |
| 流式响应解析错误 | 中 | 完善错误处理，添加日志 |
| 重试导致请求倍增 | 低 | 设置最大重试次数，监控请求量 |
| 不同提供商语义差异 | 中 | 抽象层统一语义，文档明确差异 |
| 国内 API 网络稳定性 | 中 | 重试机制，超时控制 |
| Doubao 端点 ID 管理 | 低 | 配置文件明确说明格式 |

## Open Questions

1. **是否需要请求队列?**
   - 当前决定：暂不需要，由调用方控制并发
   - 后续如遇限流问题可添加

2. **是否缓存响应?**
   - 当前决定：不缓存，每次请求都发送到 API
   - 代码生成场景不适合缓存

3. **如何处理长时间运行的请求?**
   - 当前决定：通过 context timeout 控制
   - 默认 5 分钟超时，可配置

4. **DeepSeek/Doubao 多模态能力差异?**
   - DeepSeek-VL 支持图片理解
   - Doubao 视觉能力取决于具体模型版本
   - 需要在文档中明确说明各提供商的能力边界
