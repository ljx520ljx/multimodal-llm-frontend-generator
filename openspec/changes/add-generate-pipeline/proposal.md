# Change: Add Multi-Agent Generate Pipeline

## Why

当前代码生成使用单一 LLM 调用，输出不稳定且难以调试。需要实现多 Agent Pipeline，将任务拆分为专业化步骤：

1. **LayoutAnalyzer** - 分析布局结构
2. **ComponentDetector** - 识别 UI 组件
3. **InteractionInfer** - 推断状态机转换
4. **CodeGenerator** - 生成代码
5. **CodeValidator** - 验证代码质量

这种架构提供：可观测性、单步重试、专业化 Prompt。

## What Changes

### Python Agent Core
- **ADDED** Pydantic Schemas（LayoutSchema, ComponentList, InteractionSpec, GeneratedCode）
- **ADDED** LLM Gateway（统一 OpenAI/Anthropic 调用）
- **ADDED** 5 个专业 Agent + Prompts
- **ADDED** LangGraph 工作流编排
- **ADDED** CodeValidator 工具（规则验证）
- **ADDED** `/api/v1/generate` 接口

### Go Backend
- **MODIFIED** GenerateHandler 调用 Python `/api/v1/generate`

## Impact

- Affected specs: agent-core
- Affected code:
  - Python: `agent-core/schemas/`, `agents/`, `graph/`, `tools/`, `llm/`
  - Go: `backend/internal/handler/generate.go`, `backend/internal/service/agent_client.go`
