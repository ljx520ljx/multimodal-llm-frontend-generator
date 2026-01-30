# L3 - Service 层文档 | Business Logic

> DocOps 层级: **L3 (Implementation Level)**
> 父文档: [L2 后端模块文档](../../CLAUDE.md)

## 模块职责

实现核心业务逻辑，包括图片处理、Prompt 构建和代码后处理。

## 目录结构

```
internal/service/
├── image_service.go      # 图片处理服务
├── generate_service.go   # 代码生成服务
├── prompt_service.go     # Prompt 构建服务
├── prompt_strategy.go    # Prompt 构建器实现
├── memory_store.go       # 内存会话存储
├── session_store.go      # 会话存储接口
├── types.go              # 服务层类型定义
└── CLAUDE.md             # 本文档
```

## 核心实现

### ImageService

```go
// image.go
package service

import (
    "bytes"
    "context"
    "encoding/base64"
    "image"
    "image/jpeg"
    _ "image/png"

    "github.com/disintegration/imaging"
    "github.com/google/uuid"
)

type ImageService struct {
    maxWidth  int
    maxHeight int
    quality   int
}

func NewImageService(maxWidth, maxHeight, quality int) *ImageService {
    return &ImageService{
        maxWidth:  maxWidth,
        maxHeight: maxHeight,
        quality:   quality,
    }
}

type ProcessedImage struct {
    ID     string
    Base64 string
    Order  int
}

func (s *ImageService) ProcessImage(ctx context.Context, data []byte, order int) (*ProcessedImage, error) {
    // 解码图片
    img, _, err := image.Decode(bytes.NewReader(data))
    if err != nil {
        return nil, err
    }

    // 调整尺寸
    bounds := img.Bounds()
    width, height := bounds.Dx(), bounds.Dy()

    if width > s.maxWidth || height > s.maxHeight {
        img = imaging.Fit(img, s.maxWidth, s.maxHeight, imaging.Lanczos)
    }

    // 编码为 JPEG
    var buf bytes.Buffer
    if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: s.quality}); err != nil {
        return nil, err
    }

    // Base64 编码
    b64 := base64.StdEncoding.EncodeToString(buf.Bytes())

    return &ProcessedImage{
        ID:     uuid.New().String(),
        Base64: b64,
        Order:  order,
    }, nil
}
```

### GenerateService

```go
// generate.go
package service

import (
    "context"

    "multimodal-llm-frontend-generator/internal/gateway"
)

type GenerateService struct {
    promptService *PromptService
    codeService   *CodeService
    llmGateway    gateway.LLMGateway
}

type GenerateRequest struct {
    SessionID string
    ImageIDs  []string
    Framework string
}

type GenerateEvent struct {
    Type    string `json:"type"`    // thinking | code | error
    Content string `json:"content"`
}

func NewGenerateService(
    promptService *PromptService,
    codeService *CodeService,
    llmGateway gateway.LLMGateway,
) *GenerateService {
    return &GenerateService{
        promptService: promptService,
        codeService:   codeService,
        llmGateway:    llmGateway,
    }
}

func (s *GenerateService) GenerateStream(
    ctx context.Context,
    req GenerateRequest,
) <-chan GenerateEvent {
    eventCh := make(chan GenerateEvent)

    go func() {
        defer close(eventCh)

        // 1. 构建 Prompt
        prompt, err := s.promptService.BuildGenerationPrompt(ctx, req.ImageIDs, req.Framework)
        if err != nil {
            eventCh <- GenerateEvent{Type: "error", Content: err.Error()}
            return
        }

        // 2. 调用 LLM
        streamCh := s.llmGateway.ChatStream(ctx, prompt)

        var codeBuffer string
        isThinking := true

        for chunk := range streamCh {
            if chunk.Error != nil {
                eventCh <- GenerateEvent{Type: "error", Content: chunk.Error.Error()}
                return
            }

            // 检测是否从思考转到代码
            if isThinking && containsCodeMarker(chunk.Content) {
                isThinking = false
            }

            if isThinking {
                eventCh <- GenerateEvent{Type: "thinking", Content: chunk.Content}
            } else {
                codeBuffer += chunk.Content
                eventCh <- GenerateEvent{Type: "code", Content: chunk.Content}
            }
        }

        // 3. 后处理代码
        processed, err := s.codeService.PostProcess(codeBuffer)
        if err != nil {
            eventCh <- GenerateEvent{Type: "error", Content: err.Error()}
            return
        }

        // 发送最终处理后的代码
        eventCh <- GenerateEvent{Type: "code_final", Content: processed}
    }()

    return eventCh
}

func containsCodeMarker(s string) bool {
    // 检测代码块开始标记
    return strings.Contains(s, "```") || strings.Contains(s, "import ")
}
```

### PromptService

```go
// prompt_service.go
package service

type PromptService interface {
    BuildSystemPrompt(framework string) string
    BuildGeneratePrompt(images []ImageData, framework string) []types.Message
    BuildChatPrompt(code string, message string, history []HistoryEntry) []types.Message
    BuildDiffPrompt(images []ImageData, framework string) []types.Message
}

// PromptServiceConfig holds configuration for PromptService
type PromptServiceConfig struct {
    EnableFewShot bool // Whether to include few-shot examples
}

// NewPromptServiceWithConfig creates a new PromptService with the specified configuration
func NewPromptServiceWithConfig(cfg PromptServiceConfig) PromptService {
    builder := NewPromptBuilder(cfg.EnableFewShot)
    return &promptService{builder: builder}
}
```

### CodeService

```go
// code.go
package service

import (
    "os/exec"
    "regexp"
    "strings"
)

type CodeService struct{}

func NewCodeService() *CodeService {
    return &CodeService{}
}

func (s *CodeService) PostProcess(code string) (string, error) {
    // 1. 提取代码块
    extracted := s.extractCodeBlock(code)

    // 2. 格式化代码（使用 Prettier）
    formatted, err := s.formatWithPrettier(extracted)
    if err != nil {
        // 格式化失败时返回原始代码
        return extracted, nil
    }

    return formatted, nil
}

func (s *CodeService) extractCodeBlock(content string) string {
    // 匹配 ```tsx 或 ```jsx 或 ```typescript 代码块
    re := regexp.MustCompile("```(?:tsx|jsx|typescript|javascript)?\n([\\s\\S]*?)```")
    matches := re.FindStringSubmatch(content)
    if len(matches) > 1 {
        return strings.TrimSpace(matches[1])
    }
    return content
}

func (s *CodeService) formatWithPrettier(code string) (string, error) {
    cmd := exec.Command("npx", "prettier", "--parser", "typescript", "--stdin-filepath", "App.tsx")
    cmd.Stdin = strings.NewReader(code)
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return string(output), nil
}
```

## 依赖关系

```
GenerateService
    ├── PromptService
    │   └── PromptBuilder
    ├── SessionStore
    └── LLMGateway (interface)
```

## Prompt 构建

PromptService 使用优化的 V2 Prompt 模板：

### 五步分析框架

```
Step 1: 布局识别 → 识别整体 UI 结构
Step 2: 组件识别 → 命名主要组件
Step 3: 差异检测 → 对比图片变化
Step 4: 交互推理 → 推断交互类型
Step 5: 代码生成 → 输出状态机代码
```

### 配置

```bash
# 启用 Few-Shot 示例（可选）
ENABLE_FEW_SHOT=true
```

### 使用示例

```go
// 使用配置创建 PromptService
promptService := service.NewPromptServiceWithConfig(service.PromptServiceConfig{
    EnableFewShot: true,
})
```

## 错误处理

| 错误场景 | 处理方式 |
|----------|----------|
| 图片处理失败 | 返回错误，不继续生成 |
| LLM 调用失败 | 重试 3 次后返回错误 |
| 代码格式化失败 | 返回未格式化的代码 |
| 会话过期 | 返回会话过期错误 |

## 测试示例

```go
// generate_test.go
package service

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestGenerateService_GenerateStream(t *testing.T) {
    mockPrompt := &MockPromptService{}
    mockCode := &MockCodeService{}
    mockLLM := &MockLLMGateway{}

    service := NewGenerateService(mockPrompt, mockCode, mockLLM)

    // Setup mocks
    mockPrompt.On("BuildGenerationPrompt", mock.Anything, mock.Anything, mock.Anything).
        Return(&LLMPrompt{}, nil)

    mockLLM.On("ChatStream", mock.Anything, mock.Anything).
        Return(mockStreamChannel())

    mockCode.On("PostProcess", mock.Anything).
        Return("formatted code", nil)

    // Execute
    ctx := context.Background()
    ch := service.GenerateStream(ctx, GenerateRequest{
        SessionID: "test",
        ImageIDs:  []string{"img1"},
        Framework: "react",
    })

    // Collect events
    var events []GenerateEvent
    for event := range ch {
        events = append(events, event)
    }

    assert.NotEmpty(t, events)
}
```
