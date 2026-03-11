"""Code Generator Agent - Generates HTML/Tailwind/Alpine.js code."""

from __future__ import annotations

import json
import logging
from typing import AsyncIterator, Optional, Type

from pydantic import BaseModel

from agents.base import BaseAgent
from llm.prompts import CODE_GENERATOR_PROMPT, CODE_GENERATOR_PROMPT_WITH_FEEDBACK
from schemas.code import GeneratedCode
from schemas.common import SSEEvent, SSEEventType
from schemas.component import ComponentList
from schemas.interaction import InteractionSpec
from schemas.layout import LayoutInfo
from utils.code_extractor import extract_html_code

logger = logging.getLogger(__name__)


class CodeGeneratorAgent(BaseAgent):
    """Agent that generates HTML/Tailwind CSS/Alpine.js code."""

    @property
    def name(self) -> str:
        return "CodeGenerator"

    @property
    def description(self) -> str:
        return "根据分析结果生成 HTML/Tailwind CSS/Alpine.js 代码"

    def build_prompt(
        self,
        layout_info: Optional[LayoutInfo] = None,
        component_info: Optional[ComponentList] = None,
        interaction_info: Optional[InteractionSpec] = None,
        validation_errors: Optional[list[str]] = None,
        **kwargs,
    ) -> str:
        """Build the code generation prompt."""
        layout_str = ""
        if layout_info:
            layout_str = json.dumps(layout_info.model_dump(), ensure_ascii=False, indent=2)

        component_str = ""
        if component_info:
            component_str = json.dumps(component_info.model_dump(), ensure_ascii=False, indent=2)

        interaction_str = ""
        if interaction_info:
            interaction_str = json.dumps(interaction_info.model_dump(), ensure_ascii=False, indent=2)

        # Build validation feedback section
        validation_feedback = ""
        if validation_errors:
            errors_str = "\n".join(f"- {err}" for err in validation_errors)
            validation_feedback = CODE_GENERATOR_PROMPT_WITH_FEEDBACK.format(errors=errors_str)

        return CODE_GENERATOR_PROMPT.format(
            layout_info=layout_str,
            component_info=component_str,
            interaction_info=interaction_str,
            validation_feedback=validation_feedback,
        )

    def get_output_schema(self) -> Type[BaseModel]:
        """Get the GeneratedCode schema for output."""
        return GeneratedCode

    async def run(
        self,
        images: Optional[list[dict]] = None,
        stream_events: bool = True,
        **kwargs,
    ) -> AsyncIterator[SSEEvent]:
        """Run the code generator agent with streaming.

        Uses chat_stream() instead of blocking chat() to provide real-time
        feedback during code generation. The user sees the LLM output flowing
        as THINKING events, and the final extracted HTML as a CODE event.
        """
        # Emit agent start event
        if stream_events:
            yield SSEEvent(
                event=SSEEventType.AGENT_START,
                data={"agent": self.name, "description": self.description},
            )

        try:
            # Build prompt
            prompt = self.build_prompt(**kwargs)

            # Build messages
            messages = self._build_messages(prompt, images)

            # Stream LLM response for real-time feedback
            raw_response = ""
            chunk_count = 0
            async for chunk in self.llm.chat_stream(messages):
                raw_response += chunk
                chunk_count += 1

                # Emit thinking events so user sees progress in real-time
                if stream_events:
                    yield SSEEvent(
                        event=SSEEventType.THINKING,
                        data={"content": chunk},
                    )

            logger.info(
                f"CodeGenerator received {chunk_count} chunks, "
                f"total {len(raw_response)} chars"
            )

            # Extract HTML code from full response
            html_code = self._extract_html_code(raw_response)

            # Create result object
            result = GeneratedCode(
                html=html_code,
                css=None,  # All styles in Tailwind classes
                js=None,   # All logic in Alpine.js
            )

            # Emit code event with the extracted HTML
            if stream_events:
                yield SSEEvent(
                    event=SSEEventType.CODE,
                    data={
                        "agent": self.name,
                        "html": html_code,
                    },
                )

            # Emit result event
            if stream_events:
                yield SSEEvent(
                    event=SSEEventType.AGENT_RESULT,
                    data={
                        "agent": self.name,
                        "result": result.model_dump(),
                    },
                )

            # Store result
            self._last_result = result

        except Exception as e:
            logger.error(f"CodeGenerator error: {e}")
            if stream_events:
                yield SSEEvent(
                    event=SSEEventType.ERROR,
                    data={
                        "agent": self.name,
                        "error": str(e),
                    },
                )
            raise

    def _extract_html_code(self, response: str) -> str:
        """Extract HTML code from LLM response."""
        code = extract_html_code(response)
        # Return as-is if no code block found
        return code if code else response.strip()
