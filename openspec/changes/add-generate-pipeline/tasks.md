# Tasks: Add Multi-Agent Generate Pipeline

## 1. Dependencies Setup

- [x] 1.1 更新 `requirements.txt` 添加:
  - langgraph
  - langchain
  - langchain-openai (OpenAI + 兼容 API)
  - langchain-anthropic
  - langchain-google-genai
  - beautifulsoup4 (HTML 验证)
- [x] 1.2 安装依赖并验证导入

## 2. Pydantic Schemas

- [x] 2.1 创建 `schemas/__init__.py`
- [x] 2.2 创建 `schemas/common.py` - ImageData, SSEEvent
- [x] 2.3 创建 `schemas/layout.py` - LayoutSchema, Region
- [x] 2.4 创建 `schemas/component.py` - ComponentList, Component
- [x] 2.5 创建 `schemas/interaction.py` - InteractionSpec, State, Transition
- [x] 2.6 创建 `schemas/code.py` - GeneratedCode, ValidationResult

## 3. LLM Gateway

- [x] 3.1 创建 `llm/__init__.py`
- [x] 3.2 创建 `llm/gateway.py` - LLMGateway 类
- [x] 3.3 支持原生 Provider: OpenAI, Anthropic, Google
- [x] 3.4 支持 OpenAI 兼容 Provider: DeepSeek, Doubao, GLM, Kimi
- [x] 3.5 实现流式输出支持
- [x] 3.6 更新 `app/config.py` 添加多 Provider 配置

## 4. Prompts

- [x] 4.1 创建 `llm/prompts/__init__.py`
- [x] 4.2 创建 `llm/prompts/layout.py` - LayoutAnalyzer Prompt
- [x] 4.3 创建 `llm/prompts/component.py` - ComponentDetector Prompt
- [x] 4.4 创建 `llm/prompts/interaction.py` - InteractionInfer Prompt
- [x] 4.5 创建 `llm/prompts/generator.py` - CodeGenerator Prompt

## 5. Agents

- [x] 5.1 创建 `agents/__init__.py`
- [x] 5.2 创建 `agents/base.py` - BaseAgent 基类
- [x] 5.3 创建 `agents/layout_analyzer.py`
- [x] 5.4 创建 `agents/component_detector.py`
- [x] 5.5 创建 `agents/interaction_infer.py`
- [x] 5.6 创建 `agents/code_generator.py`

## 6. Tools

- [x] 6.1 创建 `tools/__init__.py`
- [x] 6.2 创建 `tools/code_validator.py` - HTML/Alpine.js 验证

## 7. LangGraph Workflow

- [x] 7.1 创建 `graph/__init__.py`
- [x] 7.2 创建 `graph/state.py` - AgentState TypedDict
- [x] 7.3 创建 `graph/generate_workflow.py` - StateGraph 定义
- [x] 7.4 实现重试逻辑（验证失败 → CodeGenerator）

## 8. API Route

- [x] 8.1 创建 `app/routes/generate.py` - `/api/v1/generate` 接口
- [x] 8.2 实现 SSE 流式输出（每个 Agent 进度）
- [x] 8.3 在 `app/main.py` 注册路由

## 9. Go Integration

- [x] 9.1 扩展 `AgentClient` 添加 `Generate` 方法
- [x] 9.2 创建 `AgentGenerateService` 和 `AgentGenerateHandler`

## 10. Testing

- [x] 10.1 Python 端到端测试（curl 验证）
  - 21 个单元测试通过 (CodeValidator + SSEEvent)
  - E2E 脚本测试 6 项全部通过
- [x] 10.2 Go → Python → Frontend 集成测试
  - Go 测试 7/7 packages 通过
  - Python agent-core 服务正常启动
- [ ] 10.3 验证状态机自由跳转（1→2→1→3→1...）
  - 需要配置 LLM API key 进行完整测试
