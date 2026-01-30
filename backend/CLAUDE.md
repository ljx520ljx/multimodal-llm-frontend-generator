# L2 - 后端模块文档 | Backend Module

> DocOps 层级: **L2 (Module Level)**
> 父文档: [L1 项目级文档](../CLAUDE.md)

## 模块概述

基于 Go + Gin 的后端服务，负责图片处理、Prompt 构建、LLM API 调用及代码后处理。

## 架构设计

```
backend/
├── cmd/
│   └── server/
│       └── main.go             # 应用入口
│
├── internal/                    # 内部包（不对外暴露）
│   ├── handler/                 # HTTP 处理器 (L3)
│   │   ├── upload.go           # 图片上传
│   │   ├── generate.go         # 代码生成
│   │   └── chat.go             # 对话交互
│   │
│   ├── service/                 # 业务逻辑层 (L3)
│   │   ├── image.go            # 图片处理服务
│   │   ├── prompt.go           # Prompt 构建服务
│   │   └── code.go             # 代码后处理服务
│   │
│   ├── gateway/                 # 外部服务网关 (L3)
│   │   ├── openai.go           # OpenAI API
│   │   ├── gemini.go           # Google Gemini API
│   │   └── anthropic.go        # Anthropic Claude API
│   │
│   ├── middleware/              # 中间件
│   │   ├── cors.go
│   │   ├── logger.go
│   │   └── ratelimit.go
│   │
│   └── config/                  # 配置管理
│       └── config.go
│
├── pkg/                         # 可复用公共包
│   ├── prompt/                  # Prompt 模板
│   └── validator/               # 代码验证器
│
├── api/                         # API 定义
│   └── openapi.yaml            # OpenAPI 规范
│
├── go.mod
├── go.sum
└── Makefile
```

## 核心模块职责

### 1. Handler 层 (`internal/handler`)
- 接收 HTTP 请求
- 参数验证
- 调用 Service 层
- 返回响应（包括 SSE 流式输出）

### 2. Service 层 (`internal/service`)
- **ImageService**: 图片压缩、格式转换、Base64 编码
- **PromptService**: 构建多模态 Prompt，管理对话上下文
- **CodeService**: 代码格式化、语法校验、后处理

### 3. Gateway 层 (`internal/gateway`)
- 封装 LLM API 调用
- 处理认证、重试、限流
- 流式响应转发

## API 设计

### 核心接口

```yaml
# POST /api/upload
# 上传 UI 设计稿
Request:
  Content-Type: multipart/form-data
  Body:
    images[]: File[]  # 多张图片
Response:
  {
    "session_id": "uuid",
    "images": [
      { "id": "img_1", "url": "/temp/xxx.webp", "order": 0 }
    ]
  }

# POST /api/generate
# 生成代码（SSE 流式输出）
Request:
  Content-Type: application/json
  Body:
    {
      "session_id": "uuid",
      "image_ids": ["img_1", "img_2"],
      "framework": "react" | "vue"
    }
Response:
  Content-Type: text/event-stream
  data: {"type": "thinking", "content": "分析图片差异..."}
  data: {"type": "code", "content": "import React..."}
  data: {"type": "done"}

# POST /api/chat
# 自然语言修改
Request:
  {
    "session_id": "uuid",
    "message": "把按钮改成蓝色",
    "current_code": "..."
  }
Response:
  Content-Type: text/event-stream
  data: {"type": "code", "content": "...updated code..."}
```

## 技术规范

### 依赖清单

| 依赖 | 版本 | 用途 |
|------|------|------|
| github.com/gin-gonic/gin | v1.9+ | Web 框架 |
| github.com/sashabaranov/go-openai | latest | OpenAI SDK |
| github.com/google/generative-ai-go | latest | Gemini SDK |
| github.com/disintegration/imaging | latest | 图片处理 |
| go.uber.org/zap | latest | 日志 |
| github.com/spf13/viper | latest | 配置管理 |

### 代码风格

```go
// 1. 接口定义在使用方
type ImageProcessor interface {
    Compress(img []byte, quality int) ([]byte, error)
    ToBase64(img []byte) string
}

// 2. 结构体使用构造函数
func NewImageService(processor ImageProcessor) *ImageService {
    return &ImageService{processor: processor}
}

// 3. 错误处理使用 errors.Wrap
import "github.com/pkg/errors"

func (s *ImageService) Process(data []byte) (string, error) {
    compressed, err := s.processor.Compress(data, 80)
    if err != nil {
        return "", errors.Wrap(err, "failed to compress image")
    }
    return s.processor.ToBase64(compressed), nil
}

// 4. Context 贯穿整个调用链
func (h *Handler) Generate(c *gin.Context) {
    ctx := c.Request.Context()
    result, err := h.service.Generate(ctx, req)
    // ...
}
```

### 分层依赖规则

```
Handler → Service → Gateway
   ↓         ↓         ↓
  只能      只能      只能
  调用      调用      调用
 Service   Gateway   外部API
```

### 错误处理模式

```go
// pkg/errors/errors.go
type AppError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Err     error  `json:"-"`
}

var (
    ErrInvalidImage    = &AppError{Code: 400, Message: "Invalid image format"}
    ErrGenerationFailed = &AppError{Code: 500, Message: "Code generation failed"}
    ErrRateLimited     = &AppError{Code: 429, Message: "Rate limit exceeded"}
)
```

### SSE 流式输出模式

```go
func (h *Handler) Generate(c *gin.Context) {
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")

    ctx := c.Request.Context()
    ch := h.service.GenerateStream(ctx, req)

    c.Stream(func(w io.Writer) bool {
        select {
        case event, ok := <-ch:
            if !ok {
                return false
            }
            c.SSEvent("message", event)
            return true
        case <-ctx.Done():
            return false
        }
    })
}
```

## Prompt 工程

### 模板结构

```go
// pkg/prompt/templates.go
const SystemPrompt = `你是一位资深前端架构师，擅长将 UI 设计稿转换为高质量的 React 代码。

## 输出要求
- 使用 React 18+ 函数式组件
- 使用 Tailwind CSS 进行样式编写
- 确保代码可直接在 Sandpack 中运行
- 不要使用外部图片 URL，使用占位符

## 分析步骤
1. 首先描述图片中的 UI 布局结构
2. 识别组件层级关系
3. 分析两张图片的差异（如果有多张）
4. 推断交互逻辑
5. 生成完整代码`

const DiffAnalysisPrompt = `请对比以下两张 UI 设计稿，分析：
1. 视觉上发生了什么变化？
2. 这个变化可能是由什么用户操作触发的？
3. 生成对应的 React 状态和事件处理代码`
```

## 配置管理

```yaml
# config/config.yaml
server:
  port: 8080
  mode: development  # development | production

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
    model: claude-3-opus

image:
  max_size: 10485760  # 10MB
  quality: 80
  allowed_types:
    - image/png
    - image/jpeg
    - image/webp
```

## 测试策略

```go
// internal/service/image_test.go
func TestImageService_Compress(t *testing.T) {
    // Arrange
    mockProcessor := &MockImageProcessor{}
    service := NewImageService(mockProcessor)

    // Act
    result, err := service.Process(testImageBytes)

    // Assert
    require.NoError(t, err)
    assert.NotEmpty(t, result)
}
```

## 相关 L3 文档

- [Handler 层文档](./internal/handler/CLAUDE.md)
- [Service 层文档](./internal/service/CLAUDE.md)
- [Gateway 层文档](./internal/gateway/CLAUDE.md)
- [Prompt 模板文档](./pkg/prompt/CLAUDE.md)
