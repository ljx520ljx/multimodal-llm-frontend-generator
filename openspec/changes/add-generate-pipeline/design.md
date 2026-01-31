# Design: Multi-Agent Generate Pipeline

## Context

将单一 LLM 调用升级为多 Agent Pipeline，提高代码生成质量和可观测性。

**核心认知**：多 Agent = 多个专用 Prompt + 编排逻辑 + 状态传递

## Goals / Non-Goals

**Goals:**
- 实现 5 个专业 Agent 的 Pipeline
- 使用 LangGraph 编排工作流
- SSE 输出每个 Agent 的进度
- 验证失败自动重试

**Non-Goals:**
- 此阶段不实现 Chat Agent（Phase 5）
- 不修改前端代码

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Generate Pipeline                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  [Images]                                                        │
│      │                                                           │
│      ▼                                                           │
│  ┌──────────────┐                                               │
│  │LayoutAnalyzer│ → LayoutSchema                                │
│  └──────┬───────┘                                               │
│         ▼                                                        │
│  ┌──────────────────┐                                           │
│  │ComponentDetector │ → ComponentList                           │
│  └──────┬───────────┘                                           │
│         ▼                                                        │
│  ┌────────────────┐                                             │
│  │InteractionInfer│ → InteractionSpec (State Machine)          │
│  └──────┬─────────┘                                             │
│         ▼                                                        │
│  ┌──────────────┐                                               │
│  │CodeGenerator │ → GeneratedCode                               │
│  └──────┬───────┘                                               │
│         ▼                                                        │
│  ┌─────────────┐                                                │
│  │CodeValidator│ → ValidationResult                             │
│  └──────┬──────┘                                                │
│         │                                                        │
│         ├── Pass → [Final Code]                                 │
│         └── Fail → Retry CodeGenerator (max 3)                  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Decisions

### 1. LangGraph vs 简单串行

**Decision:** 使用 LangGraph StateGraph

**Why:**
- 条件边支持重试逻辑
- 状态自动传递
- 可视化调试
- 便于后续扩展并行节点

### 2. LLM Gateway 设计

```python
class LLMGateway:
    def __init__(self, provider: str, api_key: str, model: str, base_url: str = None):
        # 原生支持
        if provider == "openai":
            self.client = ChatOpenAI(api_key=api_key, model=model, base_url=base_url)
        elif provider == "anthropic":
            self.client = ChatAnthropic(api_key=api_key, model=model)
        elif provider == "google":
            self.client = ChatGoogleGenerativeAI(api_key=api_key, model=model)
        # OpenAI 兼容接口（国产模型）
        elif provider in ["deepseek", "doubao", "glm", "kimi"]:
            self.client = ChatOpenAI(
                api_key=api_key,
                model=model,
                base_url=PROVIDER_BASE_URLS[provider]
            )

    async def chat(self, messages: List[Message]) -> str: ...
    async def chat_stream(self, messages: List[Message]) -> AsyncIterator[str]: ...
```

**支持的 Provider：**

| Provider | 类型 | Base URL | 模型示例 |
|----------|------|----------|----------|
| openai | 原生 | api.openai.com | gpt-4o, gpt-4o-mini |
| anthropic | 原生 | api.anthropic.com | claude-3-5-sonnet |
| google | 原生 | generativelanguage.googleapis.com | gemini-1.5-pro |
| deepseek | OpenAI 兼容 | api.deepseek.com | deepseek-chat |
| doubao | OpenAI 兼容 | ark.cn-beijing.volces.com | doubao-pro-32k |
| glm | OpenAI 兼容 | open.bigmodel.cn | glm-4v |
| kimi | OpenAI 兼容 | api.moonshot.cn | moonshot-v1-32k |

**Why:**
- 统一接口，支持多 Provider
- 国产模型大多兼容 OpenAI API，只需切换 base_url
- LangChain 已处理 API 差异

### 3. Agent 输出结构化

**Decision:** 使用 LangChain `with_structured_output`

```python
class LayoutAnalyzerAgent:
    async def run(self, images: List[ImageData]) -> LayoutSchema:
        chain = self.llm.with_structured_output(LayoutSchema)
        return await chain.ainvoke(self.build_prompt(images))
```

**Why:**
- 强制 LLM 输出符合 Schema
- 自动重试解析失败

### 4. SSE 事件类型

| Event | Data |
|-------|------|
| `agent_start` | `{"agent": "layout_analyzer", "status": "running"}` |
| `agent_result` | `{"agent": "layout_analyzer", "result": {...}}` |
| `thinking` | `{"content": "分析布局结构..."}` |
| `code` | `{"content": "<html>..."}` |
| `error` | `{"message": "..."}` |
| `done` | `{"success": true, "retries": 0}` |

### 5. 状态机交互模型

```python
class InteractionSpec(BaseModel):
    states: List[State]          # 每张设计稿 = 一个状态
    transitions: List[Transition] # 状态间可自由跳转
    initial_state: str

# 生成的代码使用 Alpine.js 状态机
# x-data="{ currentState: 'home' }"
# @click="currentState = 'search'"
```

**核心**：不是线性 1→2→3，而是任意跳转 1→2→1→3→1→2...

### 6. CodeValidator 验证规则

1. HTML 语法检查（标签闭合）
2. Alpine.js 语法检查（x-data, @click）
3. 状态定义检查（所有状态都定义）
4. 状态转换完整性（所有 transition 都实现）

## Data Flow

```
AgentState = {
    images: List[ImageData],
    layout: LayoutSchema | None,
    components: ComponentList | None,
    interactions: InteractionSpec | None,
    code: GeneratedCode | None,
    validation: ValidationResult | None,
    retry_count: int,
    error: str | None,
}
```

## API Design

```
POST /api/v1/generate
Content-Type: application/json

Request:
{
    "session_id": "uuid",
    "images": [
        {"id": "img_1", "base64": "data:image/png;base64,...", "order": 0}
    ],
    "options": {"max_retries": 3}
}

Response (SSE):
event: agent_start
data: {"agent": "layout_analyzer", "status": "running"}

event: thinking
data: {"content": "正在分析布局结构..."}

event: agent_result
data: {"agent": "layout_analyzer", "result": {"structure": "sidebar-main", ...}}

event: agent_start
data: {"agent": "component_detector", "status": "running"}
...
event: code
data: {"content": "<!DOCTYPE html>..."}

event: done
data: {"success": true, "retries": 0}
```

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| 多次 LLM 调用成本高 | 可配置跳过某些 Agent |
| 结构化输出解析失败 | 自动重试 + fallback |
| 状态传递丢失 | TypedDict 强类型 + 日志 |

## Configuration

```bash
# agent-core/.env
LLM_PROVIDER=openai  # openai | anthropic | google | deepseek | doubao | glm | kimi
MAX_RETRIES=3

# OpenAI
OPENAI_API_KEY=sk-xxx
OPENAI_MODEL=gpt-4o
OPENAI_BASE_URL=  # 可选，自定义 base URL

# Anthropic
ANTHROPIC_API_KEY=sk-ant-xxx
ANTHROPIC_MODEL=claude-3-5-sonnet-20241022

# Google
GOOGLE_API_KEY=xxx
GOOGLE_MODEL=gemini-1.5-pro

# DeepSeek
DEEPSEEK_API_KEY=sk-xxx
DEEPSEEK_MODEL=deepseek-chat

# Doubao (豆包/火山引擎)
DOUBAO_API_KEY=xxx
DOUBAO_MODEL=doubao-pro-32k

# GLM (智谱)
GLM_API_KEY=xxx
GLM_MODEL=glm-4v

# Kimi (Moonshot)
KIMI_API_KEY=sk-xxx
KIMI_MODEL=moonshot-v1-32k
```

## Open Questions

- [x] 是否需要并行执行某些 Agent？→ 暂不需要，串行更简单
- [x] 验证失败重试哪个 Agent？→ 只重试 CodeGenerator
