# L3 - Gateway 层文档 | LLM Gateway

> DocOps 层级: **L3 (Implementation Level)**
> 父文档: [L2 后端模块文档](../../CLAUDE.md)

## 模块职责

封装外部 LLM API 调用，处理认证、重试、限流和流式响应转发。

## 目录结构

```
internal/gateway/
├── interface.go       # Gateway 接口定义
├── openai.go          # OpenAI GPT-4o 实现
├── gemini.go          # Google Gemini 实现
├── anthropic.go       # Anthropic Claude 实现
├── factory.go         # Gateway 工厂
├── retry.go           # 重试逻辑
└── CLAUDE.md          # 本文档
```

## 接口定义

```go
// interface.go
package gateway

import "context"

// LLMGateway 定义 LLM 服务的统一接口
type LLMGateway interface {
    // Chat 同步调用，返回完整响应
    Chat(ctx context.Context, prompt *LLMPrompt) (*LLMResponse, error)

    // ChatStream 流式调用，返回 chunk 通道
    ChatStream(ctx context.Context, prompt *LLMPrompt) <-chan StreamChunk
}

// LLMPrompt 统一的 Prompt 结构
type LLMPrompt struct {
    System   string
    Messages []Message
}

type Message struct {
    Role  string        // "user" | "assistant"
    Parts []ContentPart
}

type ContentPart struct {
    Type      string // "text" | "image"
    Text      string // Type == "text" 时使用
    MediaType string // Type == "image" 时使用，如 "image/jpeg"
    Data      string // Type == "image" 时使用，Base64 编码
}

// LLMResponse 统一的响应结构
type LLMResponse struct {
    Content     string
    FinishReason string
    Usage       TokenUsage
}

type TokenUsage struct {
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
}

// StreamChunk 流式响应块
type StreamChunk struct {
    Content string
    Error   error
    Done    bool
}
```

## OpenAI 实现

```go
// openai.go
package gateway

import (
    "context"
    "errors"
    "io"

    openai "github.com/sashabaranov/go-openai"
)

type OpenAIGateway struct {
    client *openai.Client
    model  string
}

func NewOpenAIGateway(apiKey, model string) *OpenAIGateway {
    client := openai.NewClient(apiKey)
    return &OpenAIGateway{
        client: client,
        model:  model,
    }
}

func (g *OpenAIGateway) ChatStream(ctx context.Context, prompt *LLMPrompt) <-chan StreamChunk {
    ch := make(chan StreamChunk)

    go func() {
        defer close(ch)

        // 转换为 OpenAI 格式
        messages := g.convertMessages(prompt)

        req := openai.ChatCompletionRequest{
            Model:    g.model,
            Messages: messages,
            Stream:   true,
        }

        stream, err := g.client.CreateChatCompletionStream(ctx, req)
        if err != nil {
            ch <- StreamChunk{Error: err}
            return
        }
        defer stream.Close()

        for {
            response, err := stream.Recv()
            if errors.Is(err, io.EOF) {
                ch <- StreamChunk{Done: true}
                return
            }
            if err != nil {
                ch <- StreamChunk{Error: err}
                return
            }

            if len(response.Choices) > 0 {
                ch <- StreamChunk{
                    Content: response.Choices[0].Delta.Content,
                }
            }
        }
    }()

    return ch
}

func (g *OpenAIGateway) convertMessages(prompt *LLMPrompt) []openai.ChatCompletionMessage {
    var messages []openai.ChatCompletionMessage

    // 系统消息
    if prompt.System != "" {
        messages = append(messages, openai.ChatCompletionMessage{
            Role:    openai.ChatMessageRoleSystem,
            Content: prompt.System,
        })
    }

    // 用户消息
    for _, msg := range prompt.Messages {
        var parts []openai.ChatMessagePart

        for _, part := range msg.Parts {
            switch part.Type {
            case "text":
                parts = append(parts, openai.ChatMessagePart{
                    Type: openai.ChatMessagePartTypeText,
                    Text: part.Text,
                })
            case "image":
                parts = append(parts, openai.ChatMessagePart{
                    Type: openai.ChatMessagePartTypeImageURL,
                    ImageURL: &openai.ChatMessageImageURL{
                        URL: "data:" + part.MediaType + ";base64," + part.Data,
                    },
                })
            }
        }

        messages = append(messages, openai.ChatCompletionMessage{
            Role:         msg.Role,
            MultiContent: parts,
        })
    }

    return messages
}
```

## Gemini 实现

```go
// gemini.go
package gateway

import (
    "context"

    "github.com/google/generative-ai-go/genai"
    "google.golang.org/api/option"
)

type GeminiGateway struct {
    client *genai.Client
    model  string
}

func NewGeminiGateway(ctx context.Context, apiKey, model string) (*GeminiGateway, error) {
    client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
    if err != nil {
        return nil, err
    }

    return &GeminiGateway{
        client: client,
        model:  model,
    }, nil
}

func (g *GeminiGateway) ChatStream(ctx context.Context, prompt *LLMPrompt) <-chan StreamChunk {
    ch := make(chan StreamChunk)

    go func() {
        defer close(ch)

        model := g.client.GenerativeModel(g.model)

        // 设置系统指令
        if prompt.System != "" {
            model.SystemInstruction = &genai.Content{
                Parts: []genai.Part{genai.Text(prompt.System)},
            }
        }

        // 构建内容
        var parts []genai.Part
        for _, msg := range prompt.Messages {
            for _, part := range msg.Parts {
                switch part.Type {
                case "text":
                    parts = append(parts, genai.Text(part.Text))
                case "image":
                    parts = append(parts, genai.ImageData(part.MediaType, []byte(part.Data)))
                }
            }
        }

        // 流式生成
        iter := model.GenerateContentStream(ctx, parts...)
        for {
            resp, err := iter.Next()
            if err != nil {
                if err.Error() == "iterator done" {
                    ch <- StreamChunk{Done: true}
                    return
                }
                ch <- StreamChunk{Error: err}
                return
            }

            for _, cand := range resp.Candidates {
                if cand.Content != nil {
                    for _, part := range cand.Content.Parts {
                        if text, ok := part.(genai.Text); ok {
                            ch <- StreamChunk{Content: string(text)}
                        }
                    }
                }
            }
        }
    }()

    return ch
}
```

## Gateway 工厂

```go
// factory.go
package gateway

import (
    "context"
    "fmt"

    "multimodal-llm-frontend-generator/internal/config"
)

func NewLLMGateway(ctx context.Context, cfg *config.LLMConfig) (LLMGateway, error) {
    switch cfg.Provider {
    case "openai":
        return NewOpenAIGateway(cfg.OpenAI.APIKey, cfg.OpenAI.Model), nil

    case "gemini":
        return NewGeminiGateway(ctx, cfg.Gemini.APIKey, cfg.Gemini.Model)

    case "anthropic":
        return NewAnthropicGateway(cfg.Anthropic.APIKey, cfg.Anthropic.Model), nil

    default:
        return nil, fmt.Errorf("unknown LLM provider: %s", cfg.Provider)
    }
}
```

## 重试机制

```go
// retry.go
package gateway

import (
    "context"
    "time"
)

type RetryConfig struct {
    MaxAttempts int
    InitialWait time.Duration
    MaxWait     time.Duration
}

var DefaultRetryConfig = RetryConfig{
    MaxAttempts: 3,
    InitialWait: 1 * time.Second,
    MaxWait:     10 * time.Second,
}

func WithRetry[T any](
    ctx context.Context,
    cfg RetryConfig,
    fn func() (T, error),
) (T, error) {
    var result T
    var lastErr error

    wait := cfg.InitialWait

    for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
        result, lastErr = fn()
        if lastErr == nil {
            return result, nil
        }

        // 检查是否可重试
        if !isRetryable(lastErr) {
            return result, lastErr
        }

        // 等待后重试
        select {
        case <-ctx.Done():
            return result, ctx.Err()
        case <-time.After(wait):
            wait = min(wait*2, cfg.MaxWait)
        }
    }

    return result, lastErr
}

func isRetryable(err error) bool {
    // 判断错误是否可重试（如网络错误、限流等）
    // 实现具体的判断逻辑
    return true
}
```

## 错误处理

| 错误类型 | 处理策略 |
|----------|----------|
| 认证失败 (401) | 不重试，返回错误 |
| 限流 (429) | 指数退避重试 |
| 服务不可用 (503) | 指数退避重试 |
| 超时 | 重试 1 次 |
| 内容过滤 | 返回特定错误码 |

## 配置示例

```yaml
# config.yaml
llm:
  provider: openai  # openai | gemini | anthropic

  openai:
    api_key: ${OPENAI_API_KEY}
    model: gpt-4o
    max_tokens: 4096

  gemini:
    api_key: ${GEMINI_API_KEY}
    model: gemini-1.5-pro

  anthropic:
    api_key: ${ANTHROPIC_API_KEY}
    model: claude-3-opus-20240229
```

## 测试示例

```go
// openai_test.go
package gateway

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestOpenAIGateway_ChatStream(t *testing.T) {
    // 使用 mock 或实际 API key 测试
    gateway := NewOpenAIGateway("test-api-key", "gpt-4o")

    prompt := &LLMPrompt{
        System: "You are a helpful assistant.",
        Messages: []Message{
            {
                Role: "user",
                Parts: []ContentPart{
                    {Type: "text", Text: "Hello"},
                },
            },
        },
    }

    ctx := context.Background()
    ch := gateway.ChatStream(ctx, prompt)

    var content string
    for chunk := range ch {
        if chunk.Error != nil {
            t.Skipf("API error: %v", chunk.Error)
            return
        }
        content += chunk.Content
    }

    assert.NotEmpty(t, content)
}
```
