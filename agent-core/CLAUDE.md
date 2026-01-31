# L2 - Agent Core 模块文档 | Python Agent Service

> DocOps 层级: **L2 (Module Level)**
> 父文档: [L1 项目级文档](../CLAUDE.md)

## 模块概述

Python Agent 服务层，负责 AI Agent 编排、Prompt 构建和 LLM 调用。

**职责分离**：
| 职责 | Go 后端 | Python Agent |
|------|---------|--------------|
| HTTP 网关 | ✅ | |
| SSE 流式转发 | ✅ | |
| 会话存储 | ✅ | |
| 图片处理 | ✅ | |
| Agent 编排 | | ✅ |
| Prompt 构建 | | ✅ |
| LLM 调用 | | ✅ |

## 架构设计

```
agent-core/
├── app/
│   ├── __init__.py
│   ├── main.py                 # FastAPI 入口
│   ├── config.py               # 配置管理
│   └── routes/
│       ├── __init__.py
│       └── echo.py             # Echo 测试接口 (Phase 2)
│
├── agents/                     # Agent 实现（Phase 3+）
│   ├── base.py                 # Agent 基类
│   ├── layout_analyzer.py      # Pipeline Agent
│   ├── component_detector.py
│   ├── interaction_infer.py
│   ├── code_generator.py
│   └── chat_agent.py           # Chat Agent
│
├── graph/                      # LangGraph 工作流（Phase 4+）
│   ├── state.py
│   └── generate_workflow.py
│
├── tools/                      # Agent 工具（Phase 5+）
│   ├── html_validator.py
│   └── interaction_checker.py
│
├── llm/                        # LLM 网关（Phase 3+）
│   ├── gateway.py
│   └── prompts/
│
├── schemas/                    # Pydantic Schema（Phase 3+）
│
├── pyproject.toml
├── requirements.txt
├── Dockerfile
└── README.md
```

## 技术栈

| 组件 | 选型 | 理由 |
|------|------|------|
| Web 框架 | FastAPI | 原生 async/await，SSE 支持 |
| Agent 框架 | LangGraph | 状态机模式，适合多步骤工作流 |
| LLM 客户端 | LangChain | 统一接口，支持多 Provider |
| 数据验证 | Pydantic | 类型安全，自动文档 |

## API 设计

### 健康检查

```
GET /health
Response: {"status": "ok"}
```

### Echo 测试（Phase 2）

```
POST /api/v1/echo
Content-Type: application/json

Request:
{
    "message": "hello",
    "count": 5,
    "delay": 0.5
}

Response (SSE):
event: message
data: {"index": 0, "message": "hello", "total": 5}
...
event: done
data: {}
```

### 首次生成（Phase 4+）

```
POST /api/v1/generate
Content-Type: application/json

Request:
{
    "session_id": "uuid",
    "images": [...],
    "options": {"max_retries": 3, "stream": true}
}

Response (SSE):
event: agent_start
data: {"agent": "layout_analyzer", "status": "running"}
...
event: code
data: {"content": "<!DOCTYPE html>..."}
event: done
data: {"success": true}
```

### 对话修改（Phase 5+）

```
POST /api/v1/chat
Content-Type: application/json

Request:
{
    "session_id": "uuid",
    "message": "把按钮改成蓝色",
    "current_code": "...",
    "images": [...],
    "history": [...]
}

Response (SSE):
event: thinking
data: {"content": "..."}
event: code
data: {"content": "..."}
event: done
data: {"success": true}
```

## 配置管理

```python
# app/config.py
class Settings(BaseSettings):
    agent_port: int = 8081
    agent_host: str = "0.0.0.0"
    log_level: str = "INFO"
    cors_origins: list[str] = [...]
    openai_api_key: str = ""
    anthropic_api_key: str = ""
```

## 开发指南

### 本地运行

```bash
cd agent-core
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
uvicorn app.main:app --port 8081 --reload
```

### Docker 运行

```bash
docker build -t agent-core .
docker run -p 8081:8081 agent-core
```

## 开发阶段

| 阶段 | 内容 | 状态 |
|------|------|------|
| Phase 1 | 项目骨架、健康检查 | ✅ 完成 |
| Phase 2 | Go ↔ Python 通信 | ✅ 完成 |
| Phase 3 | 单 Agent 验证 | 待开发 |
| Phase 4 | 完整 Pipeline | 待开发 |
| Phase 5 | ChatAgent + 工具 | 待开发 |
| Phase 6 | 集成测试 | 待开发 |

## 相关文档

- [SDD 文档](/Users/ljx/Documents/claudesidian-vault/01_Projects/多模态LLM交互原型验证平台/多Agent架构升级方案-SDD.md)
- [开发路线图](/Users/ljx/Documents/claudesidian-vault/01_Projects/多模态LLM交互原型验证平台/多Agent架构-开发路线图.md)
