# Tasks: Add Go ↔ Python Communication

## 1. Python SSE Infrastructure

- [ ] 1.1 创建 `app/routes/echo.py` - Echo 路由模块
- [ ] 1.2 实现 `/api/v1/echo` 接口（SSE 流式响应）
- [ ] 1.3 在 `app/main.py` 注册路由
- [ ] 1.4 测试 Python SSE 输出（curl 验证）

## 2. Go Agent Client

- [ ] 2.1 创建 `internal/service/agent_client.go` - AgentClient 接口
- [ ] 2.2 实现 HTTP 客户端 + SSE 解析
- [ ] 2.3 添加超时和错误处理
- [ ] 2.4 添加 Agent 服务配置（AGENT_SERVICE_URL）

## 3. Integration

- [ ] 3.1 创建测试 Handler 验证 Go → Python → Frontend 流程
- [ ] 3.2 端到端测试（前端收到 5 条流式消息）

## 4. Documentation

- [ ] 4.1 更新 Go 后端 CLAUDE.md
- [ ] 4.2 更新 Python agent-core CLAUDE.md
