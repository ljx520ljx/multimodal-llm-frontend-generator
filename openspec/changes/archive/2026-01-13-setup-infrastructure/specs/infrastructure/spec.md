# Capability: Infrastructure

基础设施能力，包括后端服务框架、前端应用框架、配置管理和开发环境。

## ADDED Requirements

### Requirement: Backend Server Framework

后端服务 SHALL 基于 Go + Gin 框架提供 HTTP API 服务。

#### Scenario: Server startup

- **WHEN** 执行 `go run cmd/server/main.go`
- **THEN** 服务在配置的端口启动（默认 8080）
- **AND** 控制台输出启动日志

#### Scenario: Health check endpoint

- **WHEN** 发送 `GET /health` 请求
- **THEN** 返回 HTTP 200 状态码
- **AND** 响应体为 `{"status": "ok"}`

### Requirement: Backend Middleware

后端服务 SHALL 实现基础中间件保障服务稳定性和可观测性。

#### Scenario: CORS handling

- **WHEN** 前端发送跨域请求
- **THEN** 服务正确设置 CORS 响应头
- **AND** 允许配置的前端域名访问

#### Scenario: Request logging

- **WHEN** 收到任意 HTTP 请求
- **THEN** 记录请求方法、路径、状态码和耗时

#### Scenario: Panic recovery

- **WHEN** Handler 发生 panic
- **THEN** 中间件捕获 panic 并返回 500 错误
- **AND** 服务继续运行不崩溃

### Requirement: Backend Configuration

后端服务 SHALL 通过环境变量和配置文件管理配置项。

#### Scenario: Load environment variables

- **WHEN** 服务启动
- **THEN** 从环境变量读取配置
- **AND** 支持 `.env` 文件作为默认值

#### Scenario: Configuration items

- **WHEN** 查看配置项
- **THEN** 包含以下必要配置：
  - `SERVER_PORT` - 服务端口（默认 8080）
  - `SERVER_MODE` - 运行模式（development/production）
  - `LLM_PROVIDER` - LLM 提供商（预留）
  - `OPENAI_API_KEY` - OpenAI API 密钥（预留）

### Requirement: Frontend Application Framework

前端应用 SHALL 基于 Next.js 14 App Router 提供 Web 界面。

#### Scenario: Development server startup

- **WHEN** 执行 `npm run dev`
- **THEN** 开发服务器在端口 3000 启动
- **AND** 支持热更新

#### Scenario: Home page access

- **WHEN** 访问 `http://localhost:3000`
- **THEN** 显示项目首页
- **AND** 页面正确应用 Tailwind CSS 样式

### Requirement: Frontend TypeScript Configuration

前端项目 SHALL 使用 TypeScript 严格模式确保类型安全。

#### Scenario: Strict type checking

- **WHEN** 运行 TypeScript 编译
- **THEN** 启用 strict 模式
- **AND** 禁止隐式 any 类型

### Requirement: Frontend Code Quality

前端项目 SHALL 使用 ESLint 和 Prettier 保证代码质量和格式一致性。

#### Scenario: ESLint check

- **WHEN** 执行 `npm run lint`
- **THEN** 检查所有 TypeScript/TSX 文件
- **AND** 无错误时返回成功

#### Scenario: Prettier format

- **WHEN** 执行 Prettier 格式化
- **THEN** 按照配置统一代码风格
- **AND** 支持与 ESLint 集成

### Requirement: Development Environment

项目 SHALL 提供标准化的开发环境配置。

#### Scenario: Git ignore configuration

- **WHEN** 查看 `.gitignore`
- **THEN** 排除 `node_modules/`、`vendor/`、`.env` 等敏感文件
- **AND** 排除 IDE 配置和构建产物

#### Scenario: Environment example files

- **WHEN** 新开发者克隆项目
- **THEN** 可从 `.env.example` 或 `.env.local.example` 复制创建本地配置
- **AND** 示例文件包含所有必要配置项说明
