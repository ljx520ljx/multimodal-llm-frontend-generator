# Design: Phase 2 后端核心服务

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         HTTP Layer                               │
│  ┌─────────────┐  ┌─────────────────┐  ┌─────────────────────┐  │
│  │ UploadHandler│  │ GenerateHandler │  │    ChatHandler      │  │
│  │ POST /upload │  │ POST /generate  │  │   POST /chat        │  │
│  │ (multipart)  │  │     (SSE)       │  │     (SSE)           │  │
│  └──────┬───────┘  └───────┬─────────┘  └──────────┬──────────┘  │
└─────────┼──────────────────┼───────────────────────┼─────────────┘
          │                  │                       │
          ▼                  ▼                       ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Service Layer                             │
│  ┌─────────────┐  ┌─────────────────┐  ┌─────────────────────┐  │
│  │ImageService │  │ PromptService   │  │  GenerateService    │  │
│  │  - Compress │  │ - BuildSystem   │  │  - Generate         │  │
│  │  - ToBase64 │  │ - BuildUser     │  │  - Chat             │  │
│  │  - Store    │  │ - BuildDiff     │  │  - StreamResponse   │  │
│  └─────────────┘  └─────────────────┘  └─────────────────────┘  │
│                            │                       │             │
│                            ▼                       ▼             │
│                    ┌───────────────────────────────────┐        │
│                    │         SessionStore              │        │
│                    │  - Get/Set Session                │        │
│                    │  - Store Images, Code, History    │        │
│                    └───────────────────────────────────┘        │
└─────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Gateway Layer (已完成)                     │
│                         LLMGateway                               │
│                    - ChatStream(ctx, req)                        │
└─────────────────────────────────────────────────────────────────┘
```

## Core Data Structures

### Session

```go
// internal/service/session.go
type Session struct {
    ID        string                 // UUID
    Images    []ImageData            // 上传的图片
    Code      string                 // 最新生成的代码
    History   []HistoryEntry         // 对话历史
    CreatedAt time.Time
    UpdatedAt time.Time
}

type ImageData struct {
    ID        string    // 图片 ID
    Filename  string    // 原始文件名
    MimeType  string    // MIME 类型
    Base64    string    // Base64 编码数据
    Order     int       // 排序顺序
}

type HistoryEntry struct {
    Role    string    // user | assistant
    Content string    // 消息内容
    Type    string    // text | code
}
```

### API Request/Response

```go
// internal/handler/types.go

// Upload
type UploadResponse struct {
    SessionID string       `json:"session_id"`
    Images    []ImageInfo  `json:"images"`
}

type ImageInfo struct {
    ID       string `json:"id"`
    Filename string `json:"filename"`
    Order    int    `json:"order"`
}

// Generate
type GenerateRequest struct {
    SessionID string   `json:"session_id"`
    ImageIDs  []string `json:"image_ids"`
    Framework string   `json:"framework"` // react | vue
}

// Chat
type ChatRequest struct {
    SessionID string `json:"session_id"`
    Message   string `json:"message"`
}

// SSE Event
type SSEEvent struct {
    Type    string `json:"type"`    // thinking | code | error | done
    Content string `json:"content"`
}
```

## Service Design

### ImageService

**职责**: 图片处理和存储

```go
type ImageService interface {
    // Process 处理上传的图片：验证、压缩、转换为 Base64
    Process(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*ImageData, error)

    // ValidateType 验证图片类型
    ValidateType(mimeType string) error

    // Compress 压缩图片到目标质量
    Compress(img image.Image, quality int) ([]byte, error)

    // ToBase64URL 转换为 data URL 格式
    ToBase64URL(data []byte, mimeType string) string
}
```

**配置**:
- 支持类型: PNG, JPEG, WebP
- 最大单张: 10MB
- 压缩质量: 80%
- 最大尺寸: 2048x2048 (超过则缩放)

### PromptService

**职责**: 构建 LLM Prompt

```go
type PromptService interface {
    // BuildSystemPrompt 构建系统 Prompt
    BuildSystemPrompt(framework string) string

    // BuildGeneratePrompt 构建代码生成 Prompt
    BuildGeneratePrompt(images []ImageData) []types.Message

    // BuildChatPrompt 构建对话修改 Prompt
    BuildChatPrompt(code string, message string, history []HistoryEntry) []types.Message

    // BuildDiffPrompt 构建差异分析 Prompt (多图场景)
    BuildDiffPrompt(images []ImageData) []types.Message
}
```

**Prompt 模板**:

```
System Prompt:
你是一位资深前端架构师，擅长将 UI 设计稿转换为高质量的 {framework} 代码。

## 输出要求
- 使用 {framework} 函数式组件
- 使用 Tailwind CSS 进行样式编写
- 确保代码可直接在 Sandpack 中运行
- 不要使用外部图片 URL，使用占位符或 SVG

## 分析步骤
1. 描述图片中的 UI 布局结构
2. 识别组件层级关系
3. 分析图片间的差异（如果有多张）
4. 推断交互逻辑
5. 生成完整可运行代码

## 输出格式
首先用 <thinking> 标签描述你的分析过程，然后用 ```jsx 或 ```tsx 包裹代码。
```

### GenerateService

**职责**: 协调代码生成流程

```go
type GenerateService interface {
    // Generate 生成代码（流式）
    Generate(ctx context.Context, sessionID string, imageIDs []string, framework string) (<-chan SSEEvent, error)

    // Chat 对话修改代码（流式）
    Chat(ctx context.Context, sessionID string, message string) (<-chan SSEEvent, error)
}
```

**流程**:

1. **Generate 流程**:
   ```
   获取 Session → 获取图片 → 构建 Prompt → 调用 LLM → 解析响应 → 保存代码 → 返回流
   ```

2. **Chat 流程**:
   ```
   获取 Session → 获取历史代码 → 构建对话 Prompt → 调用 LLM → 解析响应 → 更新代码 → 返回流
   ```

### SessionStore

**职责**: 会话存储（内存实现）

```go
type SessionStore interface {
    // Create 创建新会话
    Create(ctx context.Context) (*Session, error)

    // Get 获取会话
    Get(ctx context.Context, id string) (*Session, error)

    // Update 更新会话
    Update(ctx context.Context, session *Session) error

    // Delete 删除会话
    Delete(ctx context.Context, id string) error

    // AddImage 添加图片到会话
    AddImage(ctx context.Context, sessionID string, image *ImageData) error

    // GetImages 获取会话图片
    GetImages(ctx context.Context, sessionID string, imageIDs []string) ([]ImageData, error)

    // UpdateCode 更新生成的代码
    UpdateCode(ctx context.Context, sessionID string, code string) error

    // AddHistory 添加对话历史
    AddHistory(ctx context.Context, sessionID string, entry HistoryEntry) error
}
```

**内存实现特性**:
- 使用 `sync.RWMutex` 保护并发访问
- 会话过期时间: 30 分钟
- 后台 goroutine 定期清理过期会话

## Handler Design

### UploadHandler

```go
// POST /api/upload
// Content-Type: multipart/form-data
// Body: images[] - 多张图片文件

func (h *UploadHandler) Handle(c *gin.Context) {
    // 1. 解析 multipart form
    // 2. 创建或获取 session
    // 3. 遍历处理每张图片
    // 4. 返回 session_id 和图片信息
}
```

### GenerateHandler

```go
// POST /api/generate
// Content-Type: application/json
// Response: text/event-stream (SSE)

func (h *GenerateHandler) Handle(c *gin.Context) {
    // 1. 解析请求参数
    // 2. 设置 SSE headers
    // 3. 调用 GenerateService.Generate
    // 4. 流式返回事件
}
```

### ChatHandler

```go
// POST /api/chat
// Content-Type: application/json
// Response: text/event-stream (SSE)

func (h *ChatHandler) Handle(c *gin.Context) {
    // 1. 解析请求参数
    // 2. 设置 SSE headers
    // 3. 调用 GenerateService.Chat
    // 4. 流式返回事件
}
```

## SSE Event Format

```
event: message
data: {"type": "thinking", "content": "正在分析图片布局..."}

event: message
data: {"type": "thinking", "content": "识别到导航栏、内容区、底部栏..."}

event: message
data: {"type": "code", "content": "import React from 'react';"}

event: message
data: {"type": "code", "content": "\n\nexport default function App() {"}

event: message
data: {"type": "done", "content": ""}
```

## Error Handling

| 错误场景 | HTTP 状态码 | 错误类型 |
|----------|-------------|----------|
| 图片格式不支持 | 400 | InvalidImageFormat |
| 图片过大 | 400 | ImageTooLarge |
| 会话不存在 | 404 | SessionNotFound |
| 图片不存在 | 404 | ImageNotFound |
| LLM 调用失败 | 500 | GenerationFailed |
| LLM 限流 | 429 | RateLimited |

## Configuration

```go
// internal/config/config.go 扩展

type Config struct {
    // ... 现有配置

    // Image 配置
    ImageMaxSize      int64    // 单张图片最大字节数 (默认 10MB)
    ImageMaxTotal     int64    // 总上传最大字节数 (默认 50MB)
    ImageQuality      int      // 压缩质量 (默认 80)
    ImageMaxDimension int      // 最大尺寸 (默认 2048)
    ImageAllowedTypes []string // 允许的 MIME 类型

    // Session 配置
    SessionTTL time.Duration // 会话过期时间 (默认 30 分钟)
}
```

## Dependencies

新增依赖:
- `github.com/disintegration/imaging` - 图片处理
- `github.com/google/uuid` - UUID 生成

## Testing Strategy

1. **单元测试**
   - ImageService: 测试压缩、格式转换
   - PromptService: 测试 Prompt 构建
   - SessionStore: 测试 CRUD 操作

2. **集成测试**
   - Handler 层: 使用 httptest 模拟请求
   - GenerateService: Mock LLM Gateway

3. **端到端测试**
   - 上传图片 → 生成代码 → 修改代码 完整流程
