# L3 - Handler 层文档 | HTTP Handlers

> DocOps 层级: **L3 (Implementation Level)**
> 父文档: [L2 后端模块文档](../../CLAUDE.md)

## 模块职责

接收 HTTP 请求，验证参数，调用 Service 层，返回响应。

## 目录结构

```
internal/handler/
├── upload.go          # 图片上传处理
├── generate.go        # 代码生成处理
├── chat.go            # 对话交互处理
├── response.go        # 统一响应格式
├── handler.go         # Handler 聚合
└── CLAUDE.md          # 本文档
```

## 核心实现

### Handler 结构

```go
// handler.go
package handler

import (
    "multimodal-llm-frontend-generator/internal/service"
    "github.com/gin-gonic/gin"
)

type Handler struct {
    imageService    *service.ImageService
    generateService *service.GenerateService
    chatService     *service.ChatService
}

func NewHandler(
    imageService *service.ImageService,
    generateService *service.GenerateService,
    chatService *service.ChatService,
) *Handler {
    return &Handler{
        imageService:    imageService,
        generateService: generateService,
        chatService:     chatService,
    }
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
    api := r.Group("/api")
    {
        api.POST("/upload", h.Upload)
        api.POST("/generate", h.Generate)
        api.POST("/chat", h.Chat)
    }
}
```

### Upload Handler

```go
// upload.go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

type UploadResponse struct {
    SessionID string       `json:"session_id"`
    Images    []ImageInfo  `json:"images"`
}

type ImageInfo struct {
    ID    string `json:"id"`
    URL   string `json:"url"`
    Order int    `json:"order"`
}

func (h *Handler) Upload(c *gin.Context) {
    ctx := c.Request.Context()

    // 解析 multipart form
    form, err := c.MultipartForm()
    if err != nil {
        h.error(c, http.StatusBadRequest, "Invalid form data")
        return
    }

    files := form.File["images[]"]
    if len(files) == 0 {
        h.error(c, http.StatusBadRequest, "No images provided")
        return
    }

    // 处理上传
    result, err := h.imageService.ProcessUpload(ctx, files)
    if err != nil {
        h.error(c, http.StatusInternalServerError, err.Error())
        return
    }

    c.JSON(http.StatusOK, UploadResponse{
        SessionID: result.SessionID,
        Images:    result.Images,
    })
}
```

### Generate Handler (SSE)

```go
// generate.go
package handler

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "github.com/gin-gonic/gin"
)

type GenerateRequest struct {
    SessionID string   `json:"session_id" binding:"required"`
    ImageIDs  []string `json:"image_ids" binding:"required,min=1"`
    Framework string   `json:"framework" binding:"required,oneof=react vue"`
}

type SSEEvent struct {
    Type    string `json:"type"`
    Content string `json:"content"`
}

func (h *Handler) Generate(c *gin.Context) {
    var req GenerateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.error(c, http.StatusBadRequest, err.Error())
        return
    }

    ctx := c.Request.Context()

    // 设置 SSE 头
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")
    c.Header("X-Accel-Buffering", "no")

    // 创建事件通道
    eventCh := h.generateService.GenerateStream(ctx, service.GenerateRequest{
        SessionID: req.SessionID,
        ImageIDs:  req.ImageIDs,
        Framework: req.Framework,
    })

    // 流式输出
    c.Stream(func(w io.Writer) bool {
        select {
        case event, ok := <-eventCh:
            if !ok {
                // 通道关闭，发送完成信号
                fmt.Fprintf(w, "data: %s\n\n", mustJSON(SSEEvent{
                    Type:    "done",
                    Content: "",
                }))
                return false
            }

            fmt.Fprintf(w, "data: %s\n\n", mustJSON(event))
            c.Writer.Flush()
            return true

        case <-ctx.Done():
            return false
        }
    })
}

func mustJSON(v interface{}) string {
    b, _ := json.Marshal(v)
    return string(b)
}
```

### Chat Handler

```go
// chat.go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

type ChatRequest struct {
    SessionID   string `json:"session_id" binding:"required"`
    Message     string `json:"message" binding:"required"`
    CurrentCode string `json:"current_code" binding:"required"`
}

func (h *Handler) Chat(c *gin.Context) {
    var req ChatRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.error(c, http.StatusBadRequest, err.Error())
        return
    }

    ctx := c.Request.Context()

    // 设置 SSE 头
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")

    eventCh := h.chatService.ProcessMessage(ctx, service.ChatRequest{
        SessionID:   req.SessionID,
        Message:     req.Message,
        CurrentCode: req.CurrentCode,
    })

    c.Stream(func(w io.Writer) bool {
        select {
        case event, ok := <-eventCh:
            if !ok {
                return false
            }
            c.SSEvent("message", event)
            c.Writer.Flush()
            return true
        case <-ctx.Done():
            return false
        }
    })
}
```

### 统一响应格式

```go
// response.go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

type ErrorResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

func (h *Handler) error(c *gin.Context, code int, message string) {
    c.JSON(code, ErrorResponse{
        Code:    code,
        Message: message,
    })
}

func (h *Handler) success(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, data)
}
```

## 参数验证规则

| 字段 | 规则 | 错误消息 |
|------|------|----------|
| session_id | required | Session ID is required |
| image_ids | required, min=1 | At least one image is required |
| framework | oneof=react,vue | Framework must be react or vue |
| message | required | Message is required |

## 错误码定义

| HTTP 状态码 | 含义 |
|-------------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 429 | 请求频率限制 |
| 500 | 服务器内部错误 |

## 测试示例

```go
// upload_test.go
package handler

import (
    "bytes"
    "mime/multipart"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func TestUploadHandler(t *testing.T) {
    gin.SetMode(gin.TestMode)

    mockService := &MockImageService{}
    h := NewHandler(mockService, nil, nil)

    router := gin.New()
    h.RegisterRoutes(router)

    // 创建 multipart form
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, _ := writer.CreateFormFile("images[]", "test.png")
    part.Write([]byte("fake image data"))
    writer.Close()

    req := httptest.NewRequest(http.MethodPost, "/api/upload", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())

    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
}
```
