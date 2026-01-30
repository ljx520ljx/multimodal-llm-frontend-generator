# Multimodal LLM Frontend Generator

基于多模态大语言模型的交互原型验证平台，将连续 UI 设计稿序列自动转化为可交互的前端原型。

## Features

- **多图上传与排序** - 支持拖拽上传多张 UI 设计稿，可重新排序
- **智能代码生成** - 利用 GPT-4o/Gemini/Claude 等多模态 LLM 分析设计稿并生成代码
- **实时交互预览** - 基于 iframe 的沙箱环境，HTML + Alpine.js 实时交互体验
- **对话式修改** - 通过自然语言描述修改需求，AI 增量更新代码
- **AI 分析展示** - 展示 AI 的思考步骤和分析过程

## Tech Stack

### Frontend
- Next.js 14+ / React 18+
- TypeScript
- Tailwind CSS
- Monaco Editor
- Sandpack (CodeSandbox)
- Zustand (状态管理)

### Backend
- Go 1.21+
- Gin Web Framework
- 多 LLM Provider 支持 (OpenAI, DeepSeek, Doubao, Gemini, Anthropic)

## Quick Start

### Prerequisites

- Node.js 18+
- Go 1.21+
- LLM API Key (OpenAI/DeepSeek/其他)

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd multimodal-llm-frontend-generator

# Install frontend dependencies
cd frontend
npm install

# Install backend dependencies
cd ../backend
go mod download
```

### Configuration

复制环境变量模板并配置：

```bash
# Backend
cd backend
cp .env.example .env
```

编辑 `.env` 文件：

```env
# Server
SERVER_PORT=8080
SERVER_MODE=development
ALLOWED_ORIGINS=http://localhost:3000

# LLM Provider (选择一个)
LLM_PROVIDER=openai  # openai | deepseek | doubao | gemini | anthropic

# OpenAI
OPENAI_API_KEY=your-api-key
OPENAI_MODEL=gpt-4o

# DeepSeek
DEEPSEEK_API_KEY=your-api-key
DEEPSEEK_MODEL=deepseek-chat

# Image Processing
IMAGE_MAX_SIZE=10485760  # 10MB
IMAGE_QUALITY=80
```

### Running

```bash
# Start backend (terminal 1)
cd backend
go run cmd/server/main.go

# Start frontend (terminal 2)
cd frontend
npm run dev
```

访问 http://localhost:3000 开始使用。

## Development

### Frontend Scripts

```bash
npm run dev          # 开发服务器
npm run build        # 生产构建
npm run lint         # ESLint 检查
npm run test         # 单元测试 (Vitest)
npm run test:coverage # 测试覆盖率
npm run e2e          # E2E 测试 (Playwright)
```

### Backend Scripts

```bash
go run cmd/server/main.go  # 运行服务器
go test ./...              # 运行所有测试
go test ./... -cover       # 测试覆盖率
```

## API Reference

### POST /api/upload

上传 UI 设计稿图片。

**Request:**
```
Content-Type: multipart/form-data

images[]: File[]  # 或 files[]: File[]
session_id: string (optional)
```

**Response:**
```json
{
  "session_id": "uuid",
  "images": [
    { "id": "img_1", "filename": "design.png", "order": 0 }
  ]
}
```

### POST /api/generate

生成代码（SSE 流式输出）。

**Request:**
```json
{
  "session_id": "uuid",
  "image_ids": ["img_1", "img_2"]
}
```

**Response (SSE):**
```
data: {"type": "thinking", "content": "分析图片..."}
data: {"type": "code", "content": "import React..."}
data: {"type": "done"}
```

### POST /api/chat

对话式代码修改（SSE 流式输出）。

**Request:**
```json
{
  "session_id": "uuid",
  "message": "把按钮改成蓝色"
}
```

**Response (SSE):**
```
data: {"type": "thinking", "content": "理解修改需求..."}
data: {"type": "code", "content": "...updated code..."}
data: {"type": "done"}
```

### GET /health

健康检查端点。

**Response:**
```json
{
  "status": "ok"
}
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  Frontend (Next.js + React)                  │
│  ┌─────────────────────┬───────────────────────────────────┐ │
│  │  交互面板 (320px)   │         预览面板 (flex-1)         │ │
│  │  - 图片上传         │     HTML + Alpine.js 预览         │ │
│  │  - 对话输入         │       (iframe 沙箱)               │ │
│  │  - 对话历史         │                                   │ │
│  └─────────────────────┴───────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Backend (Go + Gin)                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐   │
│  │ Handler      │→ │ Service      │→ │ Gateway          │   │
│  │ (HTTP/SSE)   │  │ (Business)   │  │ (LLM APIs)       │   │
│  └──────────────┘  └──────────────┘  └──────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   LLM Providers                              │
│     OpenAI GPT-4o │ DeepSeek │ Doubao │ Gemini │ Claude     │
└─────────────────────────────────────────────────────────────┘
```

## Project Structure

```
multimodal-llm-frontend-generator/
├── frontend/                 # Next.js 前端
│   ├── app/                 # Next.js App Router
│   ├── src/
│   │   ├── components/      # UI 组件
│   │   ├── stores/          # Zustand 状态管理
│   │   ├── lib/             # 工具函数和 API 客户端
│   │   └── types/           # TypeScript 类型
│   ├── e2e/                 # Playwright E2E 测试
│   └── vitest/              # Vitest 测试配置
│
├── backend/                  # Go 后端
│   ├── cmd/server/          # 应用入口
│   ├── internal/
│   │   ├── handler/         # HTTP 处理器
│   │   ├── service/         # 业务逻辑
│   │   ├── gateway/         # LLM API 网关
│   │   ├── config/          # 配置管理
│   │   └── middleware/      # 中间件
│   └── pkg/prompt/          # Prompt 模板
│
└── openspec/                 # OpenSpec 规范文档
    ├── specs/               # 功能规范
    └── changes/             # 变更提案
```

## Testing

### Frontend

```bash
# 单元测试
npm run test

# 测试覆盖率
npm run test:coverage

# E2E 测试
npm run e2e
```

### Backend

```bash
# 运行所有测试
go test ./...

# 带覆盖率
go test ./... -cover

# 详细输出
go test ./... -v
```

## License

MIT
