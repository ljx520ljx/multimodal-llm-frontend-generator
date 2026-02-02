# Design: ChatAgent with Tool Calling

## Context

对话修改场景需要快速响应用户的局部修改请求。与首次生成的多 Agent Pipeline 不同，ChatAgent 采用单 Agent + Tool Calling 模式。

**核心认知**：ChatAgent = 单个通用 Prompt + 工具调用能力 + 上下文感知

## Goals / Non-Goals

**Goals:**
- 实现 ChatAgent，支持 Tool Calling
- 支持多轮工具调用循环（LLM 自主判定是否继续）
- 支持标记修改（元素选择器 + 用户描述）
- 支持原始设计稿图片上下文
- SSE 输出 thinking → tool_call → tool_result → code 事件流
- 工具：validate_html, check_interaction

**Non-Goals:**
- 不修改前端代码（由前端团队负责）
- 不实现历史记录持久化（Go 后端负责）

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                ChatAgent Flow (Multi-Turn Tool Loop)             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  [Inputs]                                                        │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ • session_id                                                │ │
│  │ • message (用户指令/标记修改)                               │ │
│  │ • current_code (当前代码)                                   │ │
│  │ • images (原始设计稿，后端自动从 session 附带)              │ │
│  │ • history (对话历史)                                        │ │
│  └────────────────────────────────────────────────────────────┘ │
│         │                                                        │
│         ▼                                                        │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                  Multi-Turn Loop                          │   │
│  │  ┌──────────────┐                                         │   │
│  │  │  ChatAgent   │ ← Tools: validate_html, check_interaction│  │
│  │  └──────┬───────┘                                         │   │
│  │         │                                                 │   │
│  │         ▼                                                 │   │
│  │  ┌─────────────┐   Yes   ┌───────────────┐               │   │
│  │  │ Tool Call?  │────────►│ Execute Tool  │               │   │
│  │  └──────┬──────┘         └───────┬───────┘               │   │
│  │         │ No                     │                        │   │
│  │         ▼                        ▼                        │   │
│  │  ┌─────────────┐         ┌───────────────┐               │   │
│  │  │ Final Code  │         │ Append Result │               │   │
│  │  └─────────────┘         │ to Messages   │               │   │
│  │                          └───────┬───────┘               │   │
│  │                                  │                        │   │
│  │                                  └──────► Continue Loop   │   │
│  │                            (LLM decides next action)      │   │
│  └──────────────────────────────────────────────────────────┘   │
│         │                                                        │
│  [SSE Events]                                                    │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ event: thinking                                             │ │
│  │ data: {"content": "我将修改按钮的背景颜色..."}             │ │
│  │                                                             │ │
│  │ event: tool_call        ─┐                                  │ │
│  │ data: {"tool": "validate_html", "args": {...}}              │ │
│  │                          │  可能多轮                        │ │
│  │ event: tool_result       │  (LLM 自主判定)                  │ │
│  │ data: {"tool": "validate_html", "result": {"valid": false}}│ │
│  │                          │                                  │ │
│  │ event: thinking          │                                  │ │
│  │ data: {"content": "验证失败，我来修复..."}                 │ │
│  │                          │                                  │ │
│  │ event: tool_call        ─┘                                  │ │
│  │ data: {"tool": "validate_html", "args": {...}}              │ │
│  │                                                             │ │
│  │ event: tool_result                                          │ │
│  │ data: {"tool": "validate_html", "result": {"valid": true}} │ │
│  │                                                             │ │
│  │ event: code                                                 │ │
│  │ data: {"html": "<!DOCTYPE html>..."}                       │ │
│  │                                                             │ │
│  │ event: done                                                 │ │
│  │ data: {"success": true}                                     │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Decisions

### 1. 单 Agent vs Pipeline

**Decision:** 使用单 Agent + Tool Calling

**Why:**
- 对话修改通常是局部改动，不需要重新分析全局
- 响应速度更快（1 次 LLM 调用 vs 5 次）
- 工具调用由 LLM 自主决定是否需要

### 2. Tool Calling 实现

```python
from langchain_core.tools import tool

@tool
def validate_html(code: str) -> dict:
    """验证 HTML 代码语法是否正确。

    Args:
        code: 要验证的 HTML 代码

    Returns:
        验证结果，包含 valid 和 errors 字段
    """
    validator = CodeValidator()
    result = validator.validate(code)
    return {
        "valid": result.is_valid,
        "errors": result.errors
    }

@tool
def check_interaction(code: str) -> dict:
    """检查状态机转换是否完整（所有状态是否可达）。

    Args:
        code: 要检查的 HTML 代码

    Returns:
        检查结果，包含 complete 和 issues 字段
    """
    validator = CodeValidator()
    # 检查状态定义和转换
    states_result = validator.validate_states_coverage(code)
    transitions_result = validator.validate_transitions(code)
    return {
        "complete": states_result["all_states_defined"] and transitions_result["all_transitions_valid"],
        "missing_states": states_result.get("undefined_states", []),
        "invalid_transitions": transitions_result.get("invalid_transitions", [])
    }
```

**Why:**
- LangChain `@tool` 装饰器自动生成 JSON Schema
- 复用现有 CodeValidator 逻辑
- 工具描述作为 LLM 的使用指南

### 3. ChatAgent 实现（多轮工具调用循环）

**核心设计**：LLM 自主决定是否需要继续调用工具

```python
class ChatAgent:
    MAX_TOOL_ITERATIONS = 5  # 防止无限循环

    def __init__(self, llm: LLMGateway):
        self.llm = llm
        self.tools = [validate_html, check_interaction]

    async def run(
        self,
        message: str,
        current_code: str,
        images: List[ImageData],  # 后端从 session 自动附带
        history: Optional[List[Message]] = None,
    ) -> AsyncIterator[SSEEvent]:
        """执行对话修改，支持多轮工具调用。"""

        # 构建初始 prompt（images 由后端确保非空）
        prompt = self.build_prompt(message, current_code, images)
        messages = self._format_history(history) + [prompt]

        # 多轮工具调用循环
        for iteration in range(self.MAX_TOOL_ITERATIONS):
            # 调用 LLM with tools
            response = await self.llm.chat_with_tools(
                messages=messages,
                tools=self.tools,
                stream=True
            )

            tool_calls_in_turn = []  # 本轮的工具调用
            final_content = None

            # 处理响应流
            async for chunk in response:
                if chunk.type == "thinking":
                    yield SSEEvent(event=SSEEventType.THINKING, data={"content": chunk.content})

                elif chunk.type == "tool_call":
                    yield SSEEvent(event=SSEEventType.TOOL_CALL, data={
                        "tool": chunk.tool_name,
                        "args": chunk.tool_args
                    })
                    # 执行工具
                    result = await self._execute_tool(chunk.tool_name, chunk.tool_args)
                    yield SSEEvent(event=SSEEventType.TOOL_RESULT, data={
                        "tool": chunk.tool_name,
                        "result": result
                    })
                    # 记录工具调用和结果
                    tool_calls_in_turn.append({
                        "tool": chunk.tool_name,
                        "args": chunk.tool_args,
                        "result": result
                    })

                elif chunk.type == "content":
                    final_content = chunk.content

            # 如果本轮有工具调用，将结果追加到 messages，继续循环
            if tool_calls_in_turn:
                # 追加 assistant 消息（包含工具调用）
                messages.append(AssistantMessage(tool_calls=tool_calls_in_turn))
                # 追加工具结果消息
                for tc in tool_calls_in_turn:
                    messages.append(ToolResultMessage(
                        tool_name=tc["tool"],
                        result=tc["result"]
                    ))
                # 继续下一轮，让 LLM 根据工具结果决定下一步
                continue

            # 如果没有工具调用，说明 LLM 完成了处理
            if final_content:
                code = self._extract_code(final_content)
                if code:
                    yield SSEEvent(event=SSEEventType.CODE, data={"html": code})
            break  # 退出循环

        yield SSEEvent(event=SSEEventType.DONE, data={"success": True})
```

### 4. 标记修改支持

标记修改是**前端构造的特殊消息格式**，ChatAgent 无需特殊处理：

```python
# 前端构造的消息格式
message = """请修改这个元素：
- 元素类型: button
- 当前类名: bg-red-500 px-4 py-2 rounded
- 当前文本: 提交

用户要求: 把它改成蓝色"""

# ChatAgent 像处理普通消息一样处理
```

### 5. 原图上下文支持

```python
def build_prompt(self, message: str, current_code: str, images: List[ImageData]) -> List:
    content = [
        {"type": "text", "text": CHAT_PROMPT.format(
            current_code=current_code,
            user_message=message
        )}
    ]

    # 添加设计稿图片
    if images:
        content.append({"type": "text", "text": "\n## 原始设计稿\n请参考以下设计稿进行精修："})
        for img in images:
            content.append({
                "type": "image_url",
                "image_url": {"url": img.base64}
            })

    return [HumanMessage(content=content)]
```

### 6. SSE 事件类型扩展

```python
class SSEEventType(str, Enum):
    # 现有事件
    AGENT_START = "agent_start"
    AGENT_RESULT = "agent_result"
    THINKING = "thinking"
    CODE = "code"
    ERROR = "error"
    DONE = "done"

    # 新增事件
    TOOL_CALL = "tool_call"
    TOOL_RESULT = "tool_result"
```

## API Design

```
POST /api/v1/chat
Content-Type: application/json

Request:
{
    "session_id": "uuid",
    "message": "把按钮改成蓝色",
    "current_code": "<!DOCTYPE html>...",
    "images": [                           // 后端从 session 自动附带
        {"id": "img_1", "base64": "data:image/png;base64,...", "order": 0}
    ],
    "history": [                          // 可选：对话历史
        {"role": "user", "content": "..."},
        {"role": "assistant", "content": "..."}
    ]
}

Response (SSE):
event: thinking
data: {"content": "我将修改按钮的背景颜色从红色改为蓝色..."}

event: tool_call
data: {"tool": "validate_html", "args": {"code": "<!DOCTYPE html>..."}}

event: tool_result
data: {"tool": "validate_html", "result": {"valid": true, "errors": []}}

event: code
data: {"html": "<!DOCTYPE html>..."}

event: done
data: {"success": true}
```

## Data Flow

```
ChatRequest:
  session_id: str
  message: str
  current_code: str
  images: List[ImageData]           # 后端从 session 自动附带，非空
  history: Optional[List[Message]]

ChatResponse (SSE):
  - THINKING: {"content": str}
  - TOOL_CALL: {"tool": str, "args": dict}
  - TOOL_RESULT: {"tool": str, "result": dict}
  - CODE: {"html": str}
  - ERROR: {"error": str}
  - DONE: {"success": bool}
```

## File Structure

```
agent-core/
├── tools/
│   ├── __init__.py           # 导出工具
│   ├── code_validator.py     # 现有：验证器核心逻辑
│   ├── html_validator.py     # 新增：validate_html 工具
│   └── interaction_checker.py # 新增：check_interaction 工具
│
├── agents/
│   ├── ...
│   └── chat_agent.py         # 新增：ChatAgent
│
├── llm/
│   └── prompts/
│       ├── ...
│       └── chat.py           # 新增：Chat Prompt
│
├── app/
│   └── routes/
│       ├── generate.py       # 现有
│       └── chat.py           # 新增：Chat 路由
│
└── schemas/
    └── common.py             # 修改：添加 TOOL_CALL, TOOL_RESULT
```

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| LLM 不调用工具导致代码错误 | Prompt 明确要求修改后验证 |
| 工具执行失败 | 捕获异常，返回错误信息给 LLM |
| 代码提取失败 | 使用多种正则模式匹配代码块 |
| 多轮循环无限执行 | MAX_TOOL_ITERATIONS = 5 硬限制 |
| 多轮调用延迟累积 | 每轮设置超时，总时间限制 30s |

## Go Backend Changes

```go
// internal/handler/chat.go
type ChatRequest struct {
    SessionID   string         `json:"session_id"`
    Message     string         `json:"message"`
    CurrentCode string         `json:"current_code"`
    Images      []ImageData    `json:"images,omitempty"`
    History     []ChatMessage  `json:"history,omitempty"`
}

func (h *ChatHandler) Handle(c *gin.Context) {
    // 1. 验证请求
    // 2. 调用 Python /api/v1/chat
    // 3. 转发 SSE 流
}

// internal/service/agent_client.go
func (c *AgentClient) Chat(ctx context.Context, req ChatRequest) (<-chan SSEEvent, error) {
    // POST /api/v1/chat
    // 返回 SSE 事件流
}
```

## Open Questions

- [x] 是否需要支持多轮工具调用循环？→ 是，LLM 自主判定是否继续
- [x] 工具执行超时如何处理？→ 设置 5s 超时，超时返回错误
- [x] 历史记录格式如何设计？→ 遵循 OpenAI messages 格式
- [x] 如何防止多轮循环无限执行？→ MAX_TOOL_ITERATIONS = 5
