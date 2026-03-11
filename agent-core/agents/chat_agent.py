"""Chat Agent - Handles conversational code modifications via single LLM call."""

from __future__ import annotations

import logging
import re
from typing import Any, AsyncIterator, Optional

from langchain_core.messages import AIMessage, HumanMessage, SystemMessage

from llm.gateway import LLMGateway
from llm.prompts import CHAT_MODIFY_PROMPT, CHAT_SYSTEM_PROMPT
from schemas.chat import ChatMessage
from schemas.common import ImageData, SSEEvent, SSEEventType
from utils.code_extractor import extract_html_code as _extract_html_code_shared
from utils.image_utils import normalize_base64_url

logger = logging.getLogger(__name__)


class ChatAgent:
    """Agent for conversational code modifications.

    Features:
    - Single LLM call for maximum speed
    - Original design image context
    - Support for marked element modifications
    """

    def __init__(self, llm: LLMGateway):
        """Initialize ChatAgent with LLM gateway.

        Args:
            llm: LLM gateway instance for making API calls
        """
        self.llm = llm

    async def run(
        self,
        message: str,
        current_code: str,
        images: list[ImageData],
        history: Optional[list[ChatMessage]] = None,
    ) -> AsyncIterator[SSEEvent]:
        """Execute chat-based code modification with a single LLM call.

        Args:
            message: User's modification request
            current_code: Current HTML code to modify
            images: Original design images (from session)
            history: Optional conversation history

        Yields:
            SSE events for thinking, code, and done
        """
        logger.info(f"ChatAgent.run() called with message: {message[:50]}...")

        # Build initial prompt
        user_prompt = self._build_user_prompt(message, current_code, images)
        messages = self._format_history(history) + [user_prompt]

        success = True

        try:
            # Single LLM call — no tool binding, maximum speed
            response = await self.llm.client.ainvoke(messages)
            content = response.content if isinstance(response.content, str) else ""
            logger.debug(f"LLM response length: {len(content)}")

            if content:
                # Extract code from response
                code = self._extract_html_code(content)

                # Emit thinking (text before code block)
                thinking_match = re.search(r"^(.*?)```", content, re.DOTALL)
                if thinking_match and thinking_match.group(1).strip():
                    yield SSEEvent(
                        event=SSEEventType.THINKING,
                        data={"content": thinking_match.group(1).strip()},
                    )

                # Emit code
                if code:
                    yield SSEEvent(
                        event=SSEEventType.CODE,
                        data={"html": code},
                    )
                else:
                    logger.warning("No HTML code block found in LLM response")
            else:
                logger.warning("LLM returned empty content")

        except Exception as e:
            logger.error(f"Error in ChatAgent: {e}")
            yield SSEEvent(
                event=SSEEventType.ERROR,
                data={"error": str(e)},
            )
            success = False

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
                content.append({
                    "type": "image_url",
                    "image_url": {"url": normalize_base64_url(img.base64)},
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

    def _extract_html_code(self, response: str) -> str:
        """Extract HTML code from LLM response."""
        return _extract_html_code_shared(response)
