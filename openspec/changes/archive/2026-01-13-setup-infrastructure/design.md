# Design: Phase 0 基础设施搭建

## Context

这是项目的初始阶段，需要建立前后端的基础代码结构。目前项目只有文档骨架（CLAUDE.md），实际代码文件尚未创建。本次变更将搭建开发所需的基础设施。

### 约束条件
- 后端使用 Go 1.21+ 和 Gin 框架
- 前端使用 Next.js 14+ 和 React 18+
- 需要支持环境变量配置（开发/生产环境切换）
- 需要为后续的 LLM API 集成预留配置

## Goals / Non-Goals

### Goals
- 建立可运行的前后端项目骨架
- 统一代码规范（ESLint、Prettier、gofmt）
- 实现基础中间件（CORS、日志、错误恢复）
- 配置开发环境变量管理

### Non-Goals
- 不实现业务逻辑
- 不集成 LLM API（Phase 1 任务）
- 不创建数据库连接（按需添加）
- 不部署到生产环境

## Decisions

### D1: 后端目录结构

**决定**: 采用 Go 标准项目布局

```
backend/
├── cmd/server/main.go    # 应用入口
├── internal/             # 内部包
│   ├── config/          # 配置管理
│   ├── handler/         # HTTP 处理器
│   └── middleware/      # 中间件
├── pkg/                  # 可复用公共包 (后续添加)
├── go.mod
├── go.sum
└── Makefile
```

**理由**:
- 符合 Go 社区最佳实践
- `internal/` 防止外部导入内部包
- `cmd/` 支持多个可执行程序
- 与 CLAUDE.md 文档中定义的结构一致

### D2: 配置管理方案

**决定**: 使用 Viper + 环境变量

```go
// 配置优先级: 环境变量 > .env 文件 > 默认值
viper.AutomaticEnv()
viper.SetConfigFile(".env")
```

**理由**:
- Viper 是 Go 生态最流行的配置库
- 支持多种配置源（环境变量、文件、远程配置）
- 12-Factor App 最佳实践

**配置项**:
```env
# Server
SERVER_PORT=8080
SERVER_MODE=development  # development | production

# LLM (预留)
OPENAI_API_KEY=
GEMINI_API_KEY=
ANTHROPIC_API_KEY=
LLM_PROVIDER=openai
```

### D3: 前端技术选型

**决定**: Next.js 14 App Router + Tailwind CSS

**理由**:
- App Router 是 Next.js 推荐的新架构
- 内置 Server Components 支持
- Tailwind CSS 快速样式开发
- 与 CLAUDE.md 文档中的技术栈一致

### D4: 中间件实现策略

**决定**: 自定义轻量级中间件

```go
// CORS - 允许前端跨域访问
// Logger - 请求日志 (基于 zap 或 gin 内置)
// Recovery - panic 恢复
```

**理由**:
- Gin 内置 Logger 和 Recovery 中间件可直接使用
- CORS 需要根据实际前端地址配置
- 保持简单，不过度封装

### D5: 工具函数 (cn)

**决定**: 使用 clsx + tailwind-merge

```typescript
// lib/utils.ts
import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
```

**理由**:
- clsx 提供条件类名组合
- tailwind-merge 解决 Tailwind 类名冲突
- 这是 React + Tailwind 项目的标准模式

## Risks / Trade-offs

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| Viper 配置复杂度 | 低 | 保持配置项最小化，只添加必要配置 |
| Next.js App Router 学习曲线 | 中 | 使用简单的 Server/Client Components 模式 |
| 中间件过度设计 | 低 | 只实现必要中间件，其他按需添加 |

## Open Questions

1. **是否需要 Docker 开发环境?**
   - 当前决定：可选，创建 docker-compose.yml 但不强制使用
   - 后续如需 Redis/PostgreSQL 可启用

2. **日志库选择?**
   - 当前决定：先用 Gin 内置 Logger，性能要求高时再切换到 zap
