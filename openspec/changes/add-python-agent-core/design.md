# Design: Python Agent Core Service

## Context

**当前状态**：Go + Gin 处理所有后端逻辑，包括 LLM 调用和 Prompt 构建。

**升级目标**：引入 Python 服务层，实现职责分离：
- **Go 后端**：HTTP 网关、SSE 流式转发、会话管理、图片处理、调用 Python Agent
- **Python Agent**：Agent 编排、Prompt 构建、LLM 调用

**选择 Python 的理由**：
1. Python 生态有成熟的 Agent 框架（LangGraph、LangChain）
2. 复杂的 LLM 编排逻辑在 Python 中更易实现和维护
3. 支持状态机工作流（LangGraph StateGraph）

## Goals / Non-Goals

**Goals:**
- 创建可运行的 FastAPI 项目骨架
- 建立标准的 Python 项目结构
- 配置管理和日志系统
- Docker 支持

**Non-Goals:**
- 此阶段不实现 LLM 调用
- 此阶段不实现 Agent 逻辑
- 此阶段不修改现有 Go 代码

## Decisions

### 1. 框架选择：FastAPI

**Decision:** 使用 FastAPI 作为 Web 框架

**Why:**
- 原生支持异步（async/await）
- 内置 OpenAPI 文档
- 原生 SSE 支持（StreamingResponse）
- 类型提示和 Pydantic 集成

**Alternatives:**
- Flask：不支持原生异步
- Django：过于重量级
- Starlette：FastAPI 已基于它，但缺少便利功能

### 2. 配置管理：Pydantic Settings

**Decision:** 使用 pydantic-settings 管理配置

**Why:**
- 类型安全
- 环境变量自动绑定
- 支持 `.env` 文件

### 3. 项目结构

```
agent-core/
├── app/
│   ├── __init__.py
│   ├── main.py          # FastAPI 入口
│   ├── config.py        # 配置管理
│   └── routes/          # 路由模块（后续扩展）
│       └── __init__.py
├── pyproject.toml       # 项目配置
├── requirements.txt     # 依赖列表
├── Dockerfile
├── .dockerignore
├── README.md
└── CLAUDE.md            # L2 模块文档
```

### 4. 端口分配

**Decision:** Python Agent 服务运行在 8081 端口

**Why:**
- 避免与 Go 后端（8080）冲突
- 便于 Docker Compose 编排

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| 增加运维复杂度 | Docker Compose 统一编排 |
| Go ↔ Python 通信延迟 | 内部网络通信，延迟可忽略 |
| 两种语言维护成本 | 职责分明：Go 网关，Python Agent |

## Migration Plan

1. **Phase 1（本阶段）**: Python 项目骨架，独立运行
2. **Phase 2（下一步）**: Go ↔ Python 通信
3. **Phase 3+**: 逐步迁移 Agent 逻辑到 Python

## Open Questions

- [x] 端口分配：8081 ✓
- [ ] 是否需要共享认证机制？（后续决定）
