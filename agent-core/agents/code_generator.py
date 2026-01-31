"""Code Generator Agent - Generates HTML/Tailwind/Alpine.js code."""

import json
import re
from typing import AsyncIterator, Optional, Type

from pydantic import BaseModel

from agents.base import BaseAgent
from llm.prompts import CODE_GENERATOR_PROMPT, CODE_GENERATOR_PROMPT_WITH_FEEDBACK
from schemas.code import GeneratedCode
from schemas.common import SSEEvent, SSEEventType
from schemas.component import ComponentList
from schemas.interaction import InteractionSpec
from schemas.layout import LayoutInfo


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
        """Build the code generation prompt.

        Args:
            layout_info: Layout information from LayoutAnalyzer
            component_info: Component information from ComponentDetector
            interaction_info: Interaction spec from InteractionInfer
            validation_errors: Optional validation errors for retry

        Returns:
            Formatted code generator prompt
        """
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
        """Get the GeneratedCode schema for output.

        Returns:
            GeneratedCode Pydantic model
        """
        return GeneratedCode

    async def run(
        self,
        images: Optional[list[dict]] = None,
        stream_events: bool = True,
        **kwargs,
    ) -> AsyncIterator[SSEEvent]:
        """Run the code generator agent.

        This overrides the base run method to handle raw text output
        and extract HTML code from markdown code blocks.

        Args:
            images: Optional list of image data
            stream_events: Whether to yield intermediate events
            **kwargs: Agent-specific parameters

        Yields:
            SSE events for agent progress and results
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

            # Get raw text response (not structured) for code generation
            raw_response = await self.llm.chat(messages)

            # Extract HTML code from response
            html_code = self._extract_html_code(raw_response)

            # Create result object
            result = GeneratedCode(
                html=html_code,
                css=None,  # All styles in Tailwind classes
                js=None,   # All logic in Alpine.js
            )

            # Emit code event with the HTML
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
        """Extract HTML code from LLM response.

        Handles markdown code blocks with ```html ... ``` format.

        Args:
            response: Raw LLM response text

        Returns:
            Extracted HTML code
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

        # Return as-is if no code block found
        return response.strip()
