# Change: Add Go ↔ Python Communication

## Why

Phase 1 已创建 Python Agent Core 项目骨架。现在需要实现 Go 后端与 Python Agent 服务之间的通信，为后续 Agent 逻辑迁移做准备。

Go 作为"门面"需要能够：
1. 调用 Python Agent API
2. 转发 SSE 流式响应给前端

## What Changes

### Go 端
- **ADDED** `AgentClient` 接口和实现 (`internal/service/agent_client.go`)
- **ADDED** Agent 服务配置（URL、超时等）

### Python 端
- **ADDED** `/api/v1/echo` 测试接口（SSE 流式响应）
- **ADDED** SSE 流式响应基础设施

## Impact

- Affected specs: agent-core (新增 SSE 通信规范)
- Affected code:
  - Go: `backend/internal/service/agent_client.go` (新增)
  - Go: `backend/internal/config/config.go` (修改，添加 Agent 服务配置)
  - Python: `agent-core/app/main.py` (添加 /api/v1/echo)
  - Python: `agent-core/app/routes/echo.py` (新增)
