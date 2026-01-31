# Tasks: Add Go ↔ Python Communication

## 1. Python SSE Infrastructure

- [x] 1.1 创建 `app/routes/echo.py` - Echo 路由模块
- [x] 1.2 实现 `/api/v1/echo` 接口（SSE 流式响应）
- [x] 1.3 在 `app/main.py` 注册路由
- [x] 1.4 测试 Python SSE 输出（curl 验证）

## 2. Go Agent Client

- [x] 2.1 创建 `internal/service/agent_client.go` - AgentClient 接口
- [x] 2.2 实现 HTTP 客户端 + SSE 解析
- [x] 2.3 添加超时和错误处理
- [x] 2.4 添加 Agent 服务配置（AGENT_SERVICE_URL）

## 3. Integration

- [x] 3.1 创建测试 Handler 验证 Go → Python → Frontend 流程
- [x] 3.2 端到端测试（前端收到流式消息）

## 4. Documentation

- [x] 4.1 更新 Go 后端 CLAUDE.md
- [x] 4.2 更新 Python agent-core CLAUDE.md
