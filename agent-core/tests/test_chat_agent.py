"""Unit tests for ChatAgent with Tool Calling.

Tests cover:
- ChatAgent initialization
- Tool mapping
- Prompt building
- History formatting
- HTML code extraction
- Tool execution with timeout
- Multi-turn tool calling loop
"""

import asyncio
from unittest.mock import AsyncMock, MagicMock, patch

import pytest
from langchain_core.messages import AIMessage, HumanMessage, SystemMessage, ToolMessage

from agents.chat_agent import ChatAgent
from schemas.chat import ChatMessage
from schemas.common import ImageData, SSEEvent, SSEEventType


# Test fixtures
TEST_HTML_CODE = """<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <script src="https://cdn.tailwindcss.com"></script>
    <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body x-data="{ currentState: 'home' }">
    <div x-show="currentState == 'home'">
        <button @click="currentState = 'about'" class="bg-blue-500 text-white px-4 py-2">About</button>
    </div>
    <div x-show="currentState == 'about'">
        <button @click="currentState = 'home'" class="bg-gray-500 text-white px-4 py-2">Home</button>
    </div>
</body>
</html>"""

TEST_IMAGE_BASE64 = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8DwHwAFBQIAX8jx0gAAAABJRU5ErkJggg=="


@pytest.fixture
def mock_llm_gateway():
    """Create a mock LLM gateway."""
    mock = MagicMock()
    mock.client = MagicMock()
    return mock


@pytest.fixture
def chat_agent(mock_llm_gateway):
    """Create a ChatAgent with mocked LLM gateway."""
    return ChatAgent(mock_llm_gateway)


class TestChatAgentInit:
    """Tests for ChatAgent initialization."""

    def test_init_with_llm(self, mock_llm_gateway):
        """Test ChatAgent initializes with LLM gateway."""
        agent = ChatAgent(mock_llm_gateway)
        assert agent.llm == mock_llm_gateway
        assert len(agent.tools) == 2
        assert agent.MAX_TOOL_ITERATIONS == 5
        assert agent.TOOL_TIMEOUT_SECONDS == 5.0

    def test_tool_map_created(self, chat_agent):
        """Test tool map is correctly created."""
        assert "validate_html" in chat_agent._tool_map
        assert "check_interaction" in chat_agent._tool_map

    def test_tools_are_langchain_tools(self, chat_agent):
        """Test tools are LangChain tool objects."""
        for tool in chat_agent.tools:
            assert hasattr(tool, "name")
            assert hasattr(tool, "invoke")


class TestBuildUserPrompt:
    """Tests for _build_user_prompt method."""

    def test_prompt_without_images(self, chat_agent):
        """Test prompt building without images."""
        message = "把按钮改成蓝色"
        current_code = "<html></html>"
        images = []

        result = chat_agent._build_user_prompt(message, current_code, images)

        assert isinstance(result, HumanMessage)
        assert isinstance(result.content, list)
        assert len(result.content) == 1
        assert result.content[0]["type"] == "text"
        assert "把按钮改成蓝色" in result.content[0]["text"]
        assert "<html></html>" in result.content[0]["text"]

    def test_prompt_with_images(self, chat_agent):
        """Test prompt building with images."""
        message = "修改样式"
        current_code = "<html></html>"
        images = [
            ImageData(id="img1", base64=TEST_IMAGE_BASE64, order=0),
            ImageData(id="img2", base64=TEST_IMAGE_BASE64, order=1),
        ]

        result = chat_agent._build_user_prompt(message, current_code, images)

        assert isinstance(result, HumanMessage)
        # Text prompt + "原始设计稿" text + 2 images = 4 content parts
        assert len(result.content) == 4
        assert result.content[0]["type"] == "text"
        assert result.content[1]["type"] == "text"
        assert "原始设计稿" in result.content[1]["text"]
        assert result.content[2]["type"] == "image_url"
        assert result.content[3]["type"] == "image_url"

    def test_prompt_adds_data_url_prefix(self, chat_agent):
        """Test that raw base64 gets data URL prefix."""
        raw_base64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB"
        images = [ImageData(id="img1", base64=raw_base64, order=0)]

        result = chat_agent._build_user_prompt("test", "<html></html>", images)

        image_url = result.content[2]["image_url"]["url"]
        assert image_url.startswith("data:image/png;base64,")

    def test_prompt_keeps_existing_data_url(self, chat_agent):
        """Test that existing data URL is not modified."""
        images = [ImageData(id="img1", base64=TEST_IMAGE_BASE64, order=0)]

        result = chat_agent._build_user_prompt("test", "<html></html>", images)

        image_url = result.content[2]["image_url"]["url"]
        assert image_url == TEST_IMAGE_BASE64


class TestFormatHistory:
    """Tests for _format_history method."""

    def test_empty_history(self, chat_agent):
        """Test formatting empty history."""
        result = chat_agent._format_history(None)

        assert len(result) == 1
        assert isinstance(result[0], SystemMessage)

    def test_history_with_messages(self, chat_agent):
        """Test formatting history with messages."""
        history = [
            ChatMessage(role="user", content="Hello"),
            ChatMessage(role="assistant", content="Hi there"),
            ChatMessage(role="user", content="Help me"),
        ]

        result = chat_agent._format_history(history)

        assert len(result) == 4  # 1 system + 3 history
        assert isinstance(result[0], SystemMessage)
        assert isinstance(result[1], HumanMessage)
        assert isinstance(result[2], AIMessage)
        assert isinstance(result[3], HumanMessage)
        assert result[1].content == "Hello"
        assert result[2].content == "Hi there"
        assert result[3].content == "Help me"


class TestExtractHtmlCode:
    """Tests for _extract_html_code method."""

    def test_extract_html_code_block(self, chat_agent):
        """Test extracting code from ```html block."""
        response = """这是一些思考...

```html
<!DOCTYPE html>
<html></html>
```

完成了。"""

        result = chat_agent._extract_html_code(response)

        assert result == "<!DOCTYPE html>\n<html></html>"

    def test_extract_generic_code_block(self, chat_agent):
        """Test extracting code from generic ``` block."""
        response = """```
<!DOCTYPE html>
<html></html>
```"""

        result = chat_agent._extract_html_code(response)

        assert result == "<!DOCTYPE html>\n<html></html>"

    def test_no_code_block(self, chat_agent):
        """Test when no code block is present."""
        response = "这是一段普通文本，没有代码块。"

        result = chat_agent._extract_html_code(response)

        assert result == ""

    def test_multiple_code_blocks(self, chat_agent):
        """Test extracting first code block when multiple exist."""
        response = """```html
<div>First</div>
```

Some text

```html
<div>Second</div>
```"""

        result = chat_agent._extract_html_code(response)

        # Should extract the first match
        assert result == "<div>First</div>"


class TestExecuteTool:
    """Tests for _execute_tool method."""

    @pytest.mark.asyncio
    async def test_execute_validate_html(self, chat_agent):
        """Test executing validate_html tool."""
        result = await chat_agent._execute_tool(
            "validate_html",
            {"code": TEST_HTML_CODE}
        )

        assert "valid" in result
        assert "errors" in result
        assert result["valid"] is True

    @pytest.mark.asyncio
    async def test_execute_check_interaction(self, chat_agent):
        """Test executing check_interaction tool."""
        result = await chat_agent._execute_tool(
            "check_interaction",
            {"code": TEST_HTML_CODE}
        )

        assert "complete" in result
        assert "defined_states" in result
        assert "home" in result["defined_states"]
        assert "about" in result["defined_states"]

    @pytest.mark.asyncio
    async def test_execute_unknown_tool(self, chat_agent):
        """Test executing unknown tool returns error."""
        result = await chat_agent._execute_tool(
            "unknown_tool",
            {"arg": "value"}
        )

        assert "error" in result
        assert "Unknown tool" in result["error"]

    @pytest.mark.asyncio
    async def test_execute_tool_with_invalid_args(self, chat_agent):
        """Test executing tool with invalid args returns error."""
        result = await chat_agent._execute_tool(
            "validate_html",
            {}  # Missing required 'code' argument
        )

        assert "error" in result

    @pytest.mark.asyncio
    async def test_tool_timeout(self, chat_agent):
        """Test tool execution timeout."""
        # Set very short timeout for test
        chat_agent.TOOL_TIMEOUT_SECONDS = 0.01

        # Create an async function that never completes in time
        async def slow_to_thread(fn, args):
            await asyncio.sleep(1)  # Sleep much longer than timeout
            return {"result": "done"}

        # Mock asyncio.to_thread to use slow function
        with patch("agents.chat_agent.asyncio.to_thread", side_effect=slow_to_thread):
            result = await chat_agent._execute_tool(
                "validate_html",
                {"code": "<html></html>"}
            )

        # Restore timeout
        chat_agent.TOOL_TIMEOUT_SECONDS = 5.0

        assert "error" in result
        assert "timed out" in result["error"]


class TestChatAgentRun:
    """Tests for ChatAgent.run() method."""

    @pytest.mark.asyncio
    async def test_run_without_tool_calls(self, mock_llm_gateway):
        """Test run() when LLM responds without tool calls."""
        agent = ChatAgent(mock_llm_gateway)

        # Mock LLM response without tool calls
        mock_response = MagicMock()
        mock_response.tool_calls = []
        mock_response.content = f"我来修改按钮颜色。\n\n```html\n{TEST_HTML_CODE}\n```"

        mock_llm_with_tools = MagicMock()
        mock_llm_with_tools.ainvoke = AsyncMock(return_value=mock_response)
        mock_llm_gateway.client.bind_tools = MagicMock(return_value=mock_llm_with_tools)

        # Collect events
        events = []
        async for event in agent.run(
            message="把按钮改成蓝色",
            current_code="<html></html>",
            images=[],
            history=None,
        ):
            events.append(event)

        # Verify events
        event_types = [e.event for e in events]
        assert SSEEventType.THINKING in event_types
        assert SSEEventType.CODE in event_types
        assert SSEEventType.DONE in event_types

        # Verify done event has success=True
        done_event = next(e for e in events if e.event == SSEEventType.DONE)
        assert done_event.data["success"] is True

    @pytest.mark.asyncio
    async def test_run_with_tool_calls(self, mock_llm_gateway):
        """Test run() when LLM uses tool calls."""
        agent = ChatAgent(mock_llm_gateway)

        # First response with tool call
        mock_response1 = MagicMock()
        mock_response1.tool_calls = [
            {"name": "validate_html", "args": {"code": TEST_HTML_CODE}, "id": "call_1"}
        ]
        mock_response1.content = ""

        # Second response with final code (no tool calls)
        mock_response2 = MagicMock()
        mock_response2.tool_calls = []
        mock_response2.content = f"验证通过，输出代码：\n\n```html\n{TEST_HTML_CODE}\n```"

        mock_llm_with_tools = MagicMock()
        mock_llm_with_tools.ainvoke = AsyncMock(side_effect=[mock_response1, mock_response2])
        mock_llm_gateway.client.bind_tools = MagicMock(return_value=mock_llm_with_tools)

        # Collect events
        events = []
        async for event in agent.run(
            message="验证一下代码",
            current_code=TEST_HTML_CODE,
            images=[],
            history=None,
        ):
            events.append(event)

        # Verify tool call events
        event_types = [e.event for e in events]
        assert SSEEventType.TOOL_CALL in event_types
        assert SSEEventType.TOOL_RESULT in event_types
        assert SSEEventType.CODE in event_types
        assert SSEEventType.DONE in event_types

        # Verify tool call event data
        tool_call_event = next(e for e in events if e.event == SSEEventType.TOOL_CALL)
        assert tool_call_event.data["tool"] == "validate_html"

        # Verify tool result event data
        tool_result_event = next(e for e in events if e.event == SSEEventType.TOOL_RESULT)
        assert tool_result_event.data["tool"] == "validate_html"
        assert "valid" in tool_result_event.data["result"]

    @pytest.mark.asyncio
    async def test_run_with_error(self, mock_llm_gateway):
        """Test run() handles errors correctly."""
        agent = ChatAgent(mock_llm_gateway)

        # Mock LLM to raise exception
        mock_llm_with_tools = MagicMock()
        mock_llm_with_tools.ainvoke = AsyncMock(side_effect=Exception("LLM error"))
        mock_llm_gateway.client.bind_tools = MagicMock(return_value=mock_llm_with_tools)

        # Collect events
        events = []
        async for event in agent.run(
            message="测试错误",
            current_code="<html></html>",
            images=[],
            history=None,
        ):
            events.append(event)

        # Verify error event
        event_types = [e.event for e in events]
        assert SSEEventType.ERROR in event_types
        assert SSEEventType.DONE in event_types

        # Verify done event has success=False
        done_event = next(e for e in events if e.event == SSEEventType.DONE)
        assert done_event.data["success"] is False

    @pytest.mark.asyncio
    async def test_run_max_iterations(self, mock_llm_gateway):
        """Test run() respects MAX_TOOL_ITERATIONS limit."""
        agent = ChatAgent(mock_llm_gateway)
        agent.MAX_TOOL_ITERATIONS = 2  # Set low limit for test

        # Mock LLM to always return tool calls
        mock_response = MagicMock()
        mock_response.tool_calls = [
            {"name": "validate_html", "args": {"code": "<html></html>"}, "id": "call_1"}
        ]
        mock_response.content = ""

        mock_llm_with_tools = MagicMock()
        mock_llm_with_tools.ainvoke = AsyncMock(return_value=mock_response)
        mock_llm_gateway.client.bind_tools = MagicMock(return_value=mock_llm_with_tools)

        # Collect events
        events = []
        async for event in agent.run(
            message="无限循环测试",
            current_code="<html></html>",
            images=[],
            history=None,
        ):
            events.append(event)

        # Count tool call events - should be limited to MAX_TOOL_ITERATIONS
        tool_call_count = sum(1 for e in events if e.event == SSEEventType.TOOL_CALL)
        assert tool_call_count == 2  # Should stop at MAX_TOOL_ITERATIONS

        # Verify done event exists
        assert SSEEventType.DONE in [e.event for e in events]


class TestToolFunctions:
    """Tests for the actual tool functions."""

    def test_validate_html_valid(self):
        """Test validate_html with valid HTML."""
        from tools import validate_html

        result = validate_html.invoke({"code": TEST_HTML_CODE})

        assert result["valid"] is True
        assert len(result["errors"]) == 0

    def test_validate_html_invalid(self):
        """Test validate_html with invalid HTML."""
        from tools import validate_html

        result = validate_html.invoke({"code": "<html><body>Hello</body></html>"})

        assert result["valid"] is False
        assert len(result["errors"]) > 0

    def test_check_interaction_complete(self):
        """Test check_interaction with complete state machine."""
        from tools import check_interaction

        result = check_interaction.invoke({"code": TEST_HTML_CODE})

        assert result["complete"] is True
        assert "home" in result["defined_states"]
        assert "about" in result["defined_states"]

    def test_check_interaction_dead_end(self):
        """Test check_interaction detects dead end states."""
        from tools import check_interaction

        dead_end_html = """
        <body x-data="{ state: 'home' }">
            <div x-show="state == 'home'">
                <button @click="state = 'about'">Go</button>
            </div>
            <div x-show="state == 'about'">
                No way back
            </div>
        </body>
        """

        result = check_interaction.invoke({"code": dead_end_html})

        assert result["complete"] is False
        assert "about" in result["dead_end_states"]
