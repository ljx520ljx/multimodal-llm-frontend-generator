# Change: Add ChatAgent with Tool Calling

**Status: ✅ IMPLEMENTED** (2026-01-31)

## Why

首次生成 Pipeline (Phase 3+4) 已完成，但用户需要通过对话修改生成的代码。对话修改场景与首次生成不同：
- 不需要重新分析布局/组件/交互（已在首次生成时完成）
- 需要快速响应局部修改请求
- 需要支持标记修改（用户点击元素）和原始设计稿图片上下文

ChatAgent 采用单 Agent + Tool Calling 模式，比 Pipeline 更轻量高效。支持多轮工具调用循环，LLM 自主判定是否需要继续调用工具（如验证失败后自动修复）。

## What Changes

### Python Agent Core

- **ADDED** `tools/html_validator.py` - validate_html() 工具
- **ADDED** `tools/interaction_checker.py` - check_interaction() 工具
- **ADDED** `agents/chat_agent.py` - ChatAgent 实现，支持 Tool Calling
- **ADDED** `llm/prompts/chat.py` - Chat Prompt 模板
- **ADDED** `app/routes/chat.py` - `/api/v1/chat` 接口
- **MODIFIED** `schemas/common.py` - 添加 TOOL_CALL, TOOL_RESULT 事件类型

### Go Backend

- **ADDED** `internal/handler/chat.go` - Chat Handler
- **MODIFIED** `internal/service/agent_client.go` - 添加 Chat 方法

## Impact

- Affected specs: agent-core
- Affected code:
  - Python: `agent-core/tools/`, `agents/chat_agent.py`, `llm/prompts/chat.py`, `app/routes/chat.py`
  - Go: `backend/internal/handler/chat.go`, `backend/internal/service/agent_client.go`
