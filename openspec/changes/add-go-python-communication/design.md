# Design: Go ↔ Python Communication

## Context

多 Agent 架构中，Go 后端作为"门面"，需要调用 Python Agent 服务并转发 SSE 流式响应给前端。

**通信流程**：
```
Frontend (Browser)
    │
    │ HTTP Request
    ▼
Go Backend (8080)
    │
    │ HTTP Request (SSE)
    ▼
Python Agent (8081)
    │
    │ SSE Stream
    ▼
Go Backend
    │
    │ SSE Forward
    ▼
Frontend
```

## Goals / Non-Goals

**Goals:**
- Go 能调用 Python API
- SSE 流式响应能正确转发
- 通过 Echo 接口验证通信链路

**Non-Goals:**
- 此阶段不实现真正的 Agent 逻辑
- 不修改现有 Generate/Chat 接口行为

## Decisions

### 1. 通信协议：HTTP + SSE

**Decision:** Go → Python 使用 HTTP POST，响应使用 SSE

**Why:**
- 简单，调试方便
- Go 标准库支持 HTTP 客户端
- SSE 是浏览器原生支持的流式协议

### 2. AgentClient 接口设计

```go
// internal/service/agent_client.go
type AgentClient interface {
    // Echo 测试接口
    Echo(ctx context.Context, req *EchoRequest) (<-chan SSEEvent, error)

    // 后续扩展
    // Generate(ctx context.Context, req *GenerateRequest) (<-chan SSEEvent, error)
    // Chat(ctx context.Context, req *ChatRequest) (<-chan SSEEvent, error)
}
```

**Why:**
- 接口隔离，便于测试
- Channel 模式与现有 SSE 处理一致

### 3. SSE 解析策略

**Decision:** 使用 `bufio.Scanner` 逐行读取 + 手动解析

**Why:**
- Go 没有标准的 SSE 客户端库
- SSE 格式简单（`event:` + `data:` + 空行）
- 避免引入第三方依赖

### 4. 错误处理

| 场景 | 处理 |
|------|------|
| Python 服务不可达 | 返回连接错误，关闭 channel |
| 请求超时 | context 取消，关闭 channel |
| SSE 解析错误 | 发送 error 事件，继续处理 |
| Python 返回非 200 | 返回 HTTP 错误 |

### 5. 配置

```bash
# Go 后端环境变量
AGENT_SERVICE_URL=http://localhost:8081
AGENT_TIMEOUT=60s
```

## Python Echo 接口设计

```python
# POST /api/v1/echo
# Request: {"message": "test", "count": 5}
# Response: SSE stream

@app.post("/api/v1/echo")
async def echo_stream(request: EchoRequest):
    async def generate():
        for i in range(request.count):
            event = {"index": i, "message": request.message}
            yield f"event: message\ndata: {json.dumps(event)}\n\n"
            await asyncio.sleep(0.5)
        yield f"event: done\ndata: {{}}\n\n"
    return StreamingResponse(generate(), media_type="text/event-stream")
```

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| SSE 解析边界情况 | 完整的单元测试覆盖 |
| 连接超时配置 | 合理的默认值 + 可配置 |
| Python 服务不稳定 | 健康检查 + 重试策略（后续） |

## Open Questions

- [x] SSE 事件名是用 `message` 还是自定义？→ 使用 `message` 保持一致
- [ ] 是否需要重试机制？→ Phase 2 不需要，后续按需添加
