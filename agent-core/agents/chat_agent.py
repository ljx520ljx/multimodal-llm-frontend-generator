"""Chat Agent - Handles conversational code modifications with tool calling."""

import asyncio
import logging
import re
from typing import Any, AsyncIterator, Optional

from langchain_core.messages import AIMessage, HumanMessage, SystemMessage, ToolMessage

from llm.gateway import LLMGateway
from llm.prompts import CHAT_MODIFY_PROMPT, CHAT_SYSTEM_PROMPT
from schemas.chat import ChatMessage
from schemas.common import ImageData, SSEEvent, SSEEventType
from tools import check_interaction, validate_html

logger = logging.getLogger(__name__)


class ChatAgent:
    """Agent for conversational code modifications with tool calling support.

    Features:
    - Single Agent + Tool Calling pattern (vs Pipeline)
    - Multi-turn tool calling loop (LLM decides when to stop)
    - Original design image context
    - Support for marked element modifications
    """

    MAX_TOOL_ITERATIONS = 5  # Prevent infinite loops
    TOOL_TIMEOUT_SECONDS = 5.0  # Tool execution timeout

    def __init__(self, llm: LLMGateway):
        """Initialize ChatAgent with LLM gateway.

        Args:
            llm: LLM gateway instance for making API calls
        """
        self.llm = llm
        self.tools = [validate_html, check_interaction]
        self._tool_map = {tool.name: tool for tool in self.tools}

    async def run(
        self,
        message: str,
        current_code: str,
        images: list[ImageData],
        history: Optional[list[ChatMessage]] = None,
    ) -> AsyncIterator[SSEEvent]:
        """Execute chat-based code modification with multi-turn tool calling.

        Args:
            message: User's modification request
            current_code: Current HTML code to modify
            images: Original design images (from session)
            history: Optional conversation history

        Yields:
            SSE events for thinking, tool calls, tool results, code, and done
        """
        logger.info(f"ChatAgent.run() called with message: {message[:50]}...")

        # Build initial prompt
        user_prompt = self._build_user_prompt(message, current_code, images)
        messages = self._format_history(history) + [user_prompt]

        # Create LLM with tools bound
        llm_with_tools = self.llm.client.bind_tools(self.tools)

        # Track success state
        success = True

        # Multi-turn tool calling loop
        for iteration in range(self.MAX_TOOL_ITERATIONS):
            logger.info(f"Tool calling iteration {iteration + 1}/{self.MAX_TOOL_ITERATIONS}")

            try:
                # Call LLM with tools
                response = await llm_with_tools.ainvoke(messages)

                # Check if there are tool calls
                if response.tool_calls:
                    # Append AI response first (only once, outside the loop)
                    messages.append(response)

                    # Process each tool call
                    for tool_call in response.tool_calls:
                        tool_name = tool_call["name"]
                        tool_args = tool_call["args"]
                        tool_id = tool_call.get("id", f"tool_{iteration}")

                        logger.info(f"Tool call: {tool_name}")

                        # Emit tool_call event
                        yield SSEEvent(
                            event=SSEEventType.TOOL_CALL,
                            data={"tool": tool_name, "args": tool_args},
                        )

                        # Execute tool with timeout
                        result = await self._execute_tool(tool_name, tool_args)

                        # Emit tool_result event
                        yield SSEEvent(
                            event=SSEEventType.TOOL_RESULT,
                            data={"tool": tool_name, "result": result},
                        )

                        # Append tool result to messages with hint to output code
                        result_content = str(result)
                        if tool_name in ["validate_html", "check_interaction"]:
                            result_content += "\n\n请根据验证结果，输出修改后的完整 HTML 代码（使用 ```html 代码块包裹）。"
                        messages.append(
                            ToolMessage(content=result_content, tool_call_id=tool_id)
                        )

                    # Continue to next iteration
                    continue

                # No tool calls - LLM completed processing
                content = response.content if isinstance(response.content, str) else ""
                logger.info(f"LLM response (no tool calls), content length: {len(content)}")
                logger.debug(f"LLM response content: {content[:500]}...")

                # Emit thinking if there's content before code
                if content:
                    # Try to extract code
                    code = self._extract_html_code(content)
                    logger.info(f"Extracted code length: {len(code) if code else 0}")

                    # If there's text before/around the code, emit as thinking
                    thinking_match = re.search(r"^(.*?)```", content, re.DOTALL)
                    if thinking_match and thinking_match.group(1).strip():
                        yield SSEEvent(
                            event=SSEEventType.THINKING,
                            data={"content": thinking_match.group(1).strip()},
                        )

                    # Emit code if extracted
                    if code:
                        yield SSEEvent(
                            event=SSEEventType.CODE,
                            data={"html": code},
                        )
                    else:
                        logger.warning(f"No HTML code block found in LLM response")
                else:
                    logger.warning("LLM returned empty content")

                # Exit loop
                break

            except Exception as e:
                logger.error(f"Error in ChatAgent iteration: {e}")
                yield SSEEvent(
                    event=SSEEventType.ERROR,
                    data={"error": str(e)},
                )
                success = False
                break

        # Emit done event
        yield SSEEvent(
            event=SSEEventType.DONE,
            data={"success": success},
        )

    def _build_user_prompt(
        self,
        message: str,
        current_code: str,
        images: list[ImageData],
    ) -> HumanMessage:
        """Build the user prompt with code and optional images.

        Args:
            message: User's modification request
            current_code: Current HTML code
            images: Original design images

        Returns:
            HumanMessage with multimodal content
        """
        # Format the text prompt
        text_prompt = CHAT_MODIFY_PROMPT.format(
            current_code=current_code,
            user_message=message,
        )

        # Build content parts
        content: list[dict[str, Any]] = [
            {"type": "text", "text": text_prompt},
        ]

        # Add design images if available
        if images:
            content.append({
                "type": "text",
                "text": "\n## 原始设计稿\n请参考以下设计稿进行精修：",
            })
            for img in images:
                # Handle both raw base64 and data URL formats
                base64_data = img.base64
                if not base64_data.startswith("data:"):
                    base64_data = f"data:image/png;base64,{base64_data}"

                content.append({
                    "type": "image_url",
                    "image_url": {"url": base64_data},
                })

        return HumanMessage(content=content)

    def _format_history(
        self,
        history: Optional[list[ChatMessage]],
    ) -> list[Any]:
        """Format conversation history as LangChain messages.

        Args:
            history: Conversation history

        Returns:
            List of LangChain message objects
        """
        messages: list[Any] = [SystemMessage(content=CHAT_SYSTEM_PROMPT)]

        if not history:
            return messages

        for msg in history:
            if msg.role == "user":
                messages.append(HumanMessage(content=msg.content))
            elif msg.role == "assistant":
                messages.append(AIMessage(content=msg.content))

        return messages

    async def _execute_tool(self, tool_name: str, tool_args: dict[str, Any]) -> dict:
        """Execute a tool with timeout and return the result.

        Args:
            tool_name: Name of the tool to execute
            tool_args: Arguments for the tool

        Returns:
            Tool execution result
        """
        if tool_name not in self._tool_map:
            return {"error": f"Unknown tool: {tool_name}"}

        try:
            tool = self._tool_map[tool_name]
            # Execute the tool with timeout (tools are sync, run in thread)
            result = await asyncio.wait_for(
                asyncio.to_thread(tool.invoke, tool_args),
                timeout=self.TOOL_TIMEOUT_SECONDS,
            )
            return result if isinstance(result, dict) else {"result": result}
        except asyncio.TimeoutError:
            logger.error(f"Tool {tool_name} timed out after {self.TOOL_TIMEOUT_SECONDS}s")
            return {"error": f"Tool {tool_name} timed out after {self.TOOL_TIMEOUT_SECONDS}s"}
        except Exception as e:
            logger.error(f"Tool execution error: {e}")
            return {"error": str(e)}

    def _extract_html_code(self, response: str) -> str:
        """Extract HTML code from LLM response.

        Handles markdown code blocks with ```html ... ``` format.

        Args:
            response: Raw LLM response text

        Returns:
            Extracted HTML code or empty string
        """
        # Try to extract from markdown code block
        html_pattern = r"```html\s*([\s\S]*?)\s*```"
        match = re.search(html_pattern, response)

        if match:
            return match.group(1).strip()

        # Try generic code block
        code_pattern = r"```\s*([\s\S]*?)\s*```"
        match = re.search(code_pattern, response)

        if match:
            return match.group(1).strip()

        return ""
