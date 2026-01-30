# Tasks: Phase 2 后端核心服务

## 1. 基础类型与配置

- [x] 1.1 创建 `backend/internal/service/types.go` - 定义 Session, ImageData, HistoryEntry 结构
- [x] 1.2 创建 `backend/internal/handler/types.go` - 定义 Request/Response 结构
- [x] 1.3 更新 `backend/internal/config/config.go` - 添加 Image 和 Session 配置项
- [x] 1.4 更新 `backend/.env.example` - 添加新配置项示例

## 2. Session 存储

- [x] 2.1 创建 `backend/internal/service/session_store.go` - 定义 SessionStore 接口
- [x] 2.2 创建 `backend/internal/service/memory_store.go` - 实现内存存储
- [x] 2.3 实现会话 CRUD 操作
- [x] 2.4 实现图片存储和检索
- [x] 2.5 实现过期会话清理 (后台 goroutine)
- [x] 2.6 编写 SessionStore 单元测试

## 3. 图片服务

- [x] 3.1 添加 `github.com/disintegration/imaging` 依赖
- [x] 3.2 创建 `backend/internal/service/image_service.go` - 定义接口和实现
- [x] 3.3 实现图片类型验证 (PNG, JPEG, WebP)
- [x] 3.4 实现图片压缩和缩放
- [x] 3.5 实现 Base64 编码转换
- [x] 3.6 编写 ImageService 单元测试

## 4. Prompt 服务

- [x] 4.1 创建 `backend/pkg/prompt/templates.go` - 定义 Prompt 模板常量
- [x] 4.2 创建 `backend/internal/service/prompt_service.go` - 定义接口和实现
- [x] 4.3 实现 BuildSystemPrompt - 支持 React/Vue 框架
- [x] 4.4 实现 BuildGeneratePrompt - 构建多模态消息
- [x] 4.5 实现 BuildChatPrompt - 构建对话上下文
- [x] 4.6 实现 BuildDiffPrompt - 构建差异分析消息
- [x] 4.7 编写 PromptService 单元测试

## 5. 代码生成服务

- [x] 5.1 创建 `backend/internal/service/generate_service.go` - 定义接口和实现
- [x] 5.2 实现 Generate 方法 - 协调图片处理和 LLM 调用
- [x] 5.3 实现 Chat 方法 - 协调对话修改流程
- [x] 5.4 实现 SSE 事件流转换 (LLM chunk → SSE event)
- [x] 5.5 实现 thinking/code 内容解析
- [x] 5.6 编写 GenerateService 单元测试 (Mock Gateway)

## 6. HTTP Handler

- [x] 6.1 创建 `backend/internal/handler/upload.go` - 实现图片上传
- [x] 6.2 创建 `backend/internal/handler/generate.go` - 实现代码生成 (SSE)
- [x] 6.3 创建 `backend/internal/handler/chat.go` - 实现对话修改 (SSE)
- [x] 6.4 创建 `backend/internal/handler/errors.go` - 统一错误处理
- [x] 6.5 编写 Handler 集成测试 (httptest)

## 7. 路由与依赖注入

- [x] 7.1 创建 `backend/internal/app/app.go` - 应用初始化和依赖注入
- [x] 7.2 更新 `backend/cmd/server/main.go` - 注册新路由
- [x] 7.3 注册 /api/upload, /api/generate, /api/chat 端点
- [x] 7.4 添加请求大小限制中间件

## 8. 集成验证

- [x] 8.1 编写端到端测试脚本
- [x] 8.2 单元测试：上传图片接口 (3 tests)
- [x] 8.3 单元测试：代码生成接口 (3 tests)
- [x] 8.4 单元测试：对话修改接口 (2 tests)
- [x] 8.5 验证编译通过和所有测试通过

## Dependencies

```
1.x (基础类型) ─┬─→ 2.x (Session)  ─┬─→ 5.x (Generate) ─┬─→ 6.x (Handler) ─→ 7.x (路由)
               │                   │                   │
               └─→ 3.x (Image)  ───┤                   │
                                   │                   │
               └─→ 4.x (Prompt) ───┘                   │
                                                       │
                                         8.x (验证) ←──┘
```

## Validation

```bash
# 运行所有单元测试
cd backend && go test -v ./internal/service/... ./internal/handler/...

# 启动服务
cd backend && go run cmd/server/main.go

# 测试上传接口
curl -X POST http://localhost:8080/api/upload \
  -F "images[]=@test1.png" \
  -F "images[]=@test2.png"

# 测试生成接口 (SSE)
curl -N -X POST http://localhost:8080/api/generate \
  -H "Content-Type: application/json" \
  -d '{"session_id": "xxx", "image_ids": ["img1", "img2"], "framework": "react"}'

# 测试对话接口 (SSE)
curl -N -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"session_id": "xxx", "message": "把按钮改成蓝色"}'
```

## Completion Notes

### 已完成的工作

1. **基础类型** (`backend/internal/service/types.go`, `backend/internal/handler/types.go`)
   - Session, ImageData, HistoryEntry 数据结构
   - UploadResponse, GenerateRequest, ChatRequest API 类型
   - SSEEvent 流式事件结构
   - 错误码常量

2. **Session 存储** (`backend/internal/service/session_store.go`, `memory_store.go`)
   - SessionStore 接口定义
   - MemoryStore 内存实现
   - 会话 CRUD、图片存储、历史记录管理
   - 后台清理过期会话

3. **图片服务** (`backend/internal/service/image_service.go`)
   - 图片类型验证 (PNG, JPEG, WebP)
   - 图片压缩和缩放 (使用 imaging 库)
   - Base64 data URL 转换
   - MIME 类型检测

4. **Prompt 服务** (`backend/pkg/prompt/templates.go`, `internal/service/prompt_service.go`)
   - 系统 Prompt 模板 (支持 React/Vue)
   - 单图/多图生成 Prompt
   - 对话修改 Prompt
   - 差异分析 Prompt

5. **代码生成服务** (`backend/internal/service/generate_service.go`)
   - Generate 方法：图片 → Prompt → LLM → SSE 流
   - Chat 方法：对话修改 → LLM → SSE 流
   - thinking/code 内容检测
   - 代码提取和保存

6. **HTTP Handler** (`backend/internal/handler/`)
   - UploadHandler: POST /api/upload
   - GenerateHandler: POST /api/generate (SSE)
   - ChatHandler: POST /api/chat (SSE)
   - 统一错误处理

7. **应用初始化** (`backend/internal/app/app.go`)
   - 依赖注入
   - 路由配置
   - 优雅关闭

### 测试结果

```
ok  multimodal-llm-frontend-generator/internal/gateway   0.797s
ok  multimodal-llm-frontend-generator/internal/handler   1.250s
ok  multimodal-llm-frontend-generator/internal/service   1.815s
```

### 新增文件列表

```
backend/
├── internal/
│   ├── app/
│   │   └── app.go                    # 应用初始化
│   ├── handler/
│   │   ├── types.go                  # API 类型定义
│   │   ├── errors.go                 # 错误处理
│   │   ├── upload.go                 # 上传 Handler
│   │   ├── generate.go               # 生成 Handler
│   │   ├── chat.go                   # 对话 Handler
│   │   └── handler_test.go           # Handler 测试
│   └── service/
│       ├── types.go                  # 服务类型定义
│       ├── session_store.go          # SessionStore 接口
│       ├── memory_store.go           # 内存存储实现
│       ├── memory_store_test.go      # 存储测试
│       ├── image_service.go          # 图片服务
│       ├── image_service_test.go     # 图片服务测试
│       ├── prompt_service.go         # Prompt 服务
│       ├── prompt_service_test.go    # Prompt 服务测试
│       ├── generate_service.go       # 生成服务
│       └── generate_service_test.go  # 生成服务测试
└── pkg/
    └── prompt/
        └── templates.go              # Prompt 模板
```
