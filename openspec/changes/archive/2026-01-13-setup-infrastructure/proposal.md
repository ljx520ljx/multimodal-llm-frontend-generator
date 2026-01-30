# Change: Phase 0 基础设施搭建

## Why

项目目前只有文档骨架，缺少实际的代码基础设施。在开发任何业务功能之前，需要先搭建前后端项目的基础框架，确保开发环境可用、代码规范统一、基础中间件就绪。

## What Changes

### 后端基础设施 (Go + Gin)
- 初始化 Go module
- 集成 Gin Web 框架
- 配置 Viper 环境变量管理
- 实现 CORS、Logger、Recovery 中间件
- 创建健康检查端点 `GET /health`
- 创建 Makefile 和 .env.example

### 前端基础设施 (Next.js + React)
- 使用 create-next-app 初始化 Next.js 14 项目
- 配置 TypeScript 严格模式
- 配置 Tailwind CSS
- 配置 ESLint + Prettier
- 创建基础布局组件
- 创建 .env.local.example

### 开发环境
- 配置 .gitignore
- 创建 docker-compose.yml (可选)

## Impact

- **Affected specs**: 新增 `infrastructure` capability
- **Affected code**:
  - `backend/` - 新增 cmd/, internal/config/, internal/middleware/, go.mod, Makefile
  - `frontend/` - 新增 Next.js 项目结构、配置文件
  - 根目录 - .gitignore, docker-compose.yml

## Acceptance Criteria

- 后端启动无报错，`GET /health` 返回 200
- 前端启动无报错，首页可访问
- 代码可提交到 Git
- ESLint/Prettier 检查通过
