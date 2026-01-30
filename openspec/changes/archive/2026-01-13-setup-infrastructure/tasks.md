# Tasks: Phase 0 基础设施搭建

## 1. 后端项目初始化

- [x] 1.1 创建 `backend/go.mod` 初始化 Go module (`multimodal-llm-frontend-generator`)
- [x] 1.2 创建 `backend/cmd/server/main.go` 应用入口
- [x] 1.3 创建 `backend/internal/config/config.go` 配置管理 (集成 Viper)
- [x] 1.4 创建 `backend/internal/middleware/cors.go` CORS 中间件
- [x] 1.5 创建 `backend/internal/middleware/logger.go` 日志中间件
- [x] 1.6 创建 `backend/internal/middleware/recovery.go` 错误恢复中间件
- [x] 1.7 创建 `backend/internal/handler/health.go` 健康检查端点
- [x] 1.8 创建 `backend/Makefile` 构建脚本
- [x] 1.9 创建 `backend/.env.example` 环境变量模板
- [x] 1.10 运行 `go mod tidy` 安装依赖
- [x] 1.11 验证后端启动成功，`GET /health` 返回 200

## 2. 前端项目初始化

- [x] 2.1 使用 `create-next-app` 初始化 Next.js 14 项目 (TypeScript + Tailwind + ESLint)
- [x] 2.2 配置 `tsconfig.json` 启用严格模式
- [x] 2.3 配置 `tailwind.config.ts` 扩展 Tailwind 设置
- [x] 2.4 创建 `.prettierrc` 配置 Prettier
- [x] 2.5 更新 `.eslintrc.json` 加强 ESLint 规则
- [x] 2.6 创建 `frontend/app/layout.tsx` 根布局
- [x] 2.7 创建 `frontend/app/page.tsx` 首页
- [x] 2.8 创建 `frontend/lib/utils.ts` 工具函数 (cn 函数)
- [x] 2.9 创建 `frontend/.env.local.example` 环境变量模板
- [x] 2.10 运行 `npm install` 安装依赖
- [x] 2.11 验证前端启动成功，首页可访问

## 3. 开发环境配置

- [x] 3.1 创建根目录 `.gitignore` 文件
- [x] 3.2 创建 `docker-compose.yml` (可选，预留 Redis/PostgreSQL)
- [x] 3.3 运行 ESLint 检查，确保无错误
- [x] 3.4 运行 Prettier 检查，确保代码格式正确

## Dependencies

- 任务组 1 和 2 可以并行执行
- 任务组 3 依赖任务组 1 和 2 完成

## Validation

```bash
# 后端验证
cd backend && go run cmd/server/main.go
curl http://localhost:8080/health
# 期望返回: {"status": "ok"}

# 前端验证
cd frontend && npm run dev
# 访问 http://localhost:3000 应显示首页

# 代码检查
cd frontend && npm run lint
cd frontend && npx prettier --check "app/**/*.{ts,tsx}" "lib/**/*.ts"
```

## Completion Notes

- 后端: Go + Gin 服务正常启动，`/health` 端点返回 `{"status":"ok"}`
- 前端: Next.js 14 项目构建成功，ESLint 和 Prettier 检查通过
- 环境: .gitignore 和 docker-compose.yml 已创建
