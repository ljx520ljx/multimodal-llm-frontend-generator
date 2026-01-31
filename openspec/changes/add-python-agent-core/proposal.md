# Change: Add Python Agent Core Service

## Why

为支持多 Agent 架构升级，需要引入 Python 服务层来托管 AI Agent 逻辑。Python 生态在 Agent 框架（LangGraph、LangChain）方面更成熟，适合处理复杂的 LLM 编排任务。这是多 Agent 架构的第一步——创建 Python 项目骨架。

## What Changes

- **ADDED** `agent-core/` 目录：Python FastAPI 项目
- **ADDED** FastAPI 入口和健康检查接口
- **ADDED** 配置管理和日志系统
- **ADDED** Docker 支持

## Impact

- Affected specs: agent-core (新增)
- Affected code:
  - 新增 `agent-core/` 目录
  - 新增 `agent-core/app/main.py` - FastAPI 入口
  - 新增 `agent-core/app/config.py` - 配置管理
  - 新增 `agent-core/Dockerfile`
  - 新增 `agent-core/pyproject.toml` / `requirements.txt`
