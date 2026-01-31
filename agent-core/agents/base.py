"""Base agent class for all pipeline agents."""

from abc import ABC, abstractmethod
from typing import Any, AsyncIterator, Type

from pydantic import BaseModel

from llm.gateway import LLMGateway
from schemas.common import SSEEvent, SSEEventType


class BaseAgent(ABC):
    """Base class for all agents in the pipeline."""

    def __init__(self, llm: LLMGateway):
        """Initialize agent with LLM gateway.

        Args:
            llm: LLM gateway instance for making API calls
        """
        self.llm = llm

    @property
    @abstractmethod
    def name(self) -> str:
        """Agent name for logging and SSE events."""
        pass

    @property
    @abstractmethod
    def description(self) -> str:
        """Agent description."""
        pass

    @abstractmethod
    def build_prompt(self, **kwargs) -> str:
        """Build the prompt for this agent.

        Args:
            **kwargs: Agent-specific parameters

        Returns:
            Formatted prompt string
        """
        pass

    @abstractmethod
    def get_output_schema(self) -> Type[BaseModel]:
        """Get the Pydantic schema for structured output.

        Returns:
            Pydantic model class for output validation
        """
        pass

    async def run(
        self,
        images: list[dict] | None = None,
        stream_events: bool = True,
        **kwargs,
    ) -> AsyncIterator[SSEEvent]:
        """Run the agent and yield SSE events.

        Args:
            images: Optional list of image data dicts with base64 content
            stream_events: Whether to yield intermediate SSE events
            **kwargs: Agent-specific parameters passed to build_prompt

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

            # Get structured output from LLM
            output_schema = self.get_output_schema()
            result = await self.llm.chat_structured(messages, output_schema)

            # Emit result event
            if stream_events:
                yield SSEEvent(
                    event=SSEEventType.AGENT_RESULT,
                    data={
                        "agent": self.name,
                        "result": result.model_dump(),
                    },
                )

            # Store result for pipeline access
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

    def _build_messages(
        self,
        prompt: str,
        images: list[dict] | None = None,
    ) -> list[dict]:
        """Build message list for LLM call.

        Args:
            prompt: The text prompt
            images: Optional list of image data with base64 content

        Returns:
            List of message dicts for LLM API
        """
        if not images:
            return [{"role": "user", "content": prompt}]

        # Build multimodal message with images
        content = []

        # Add images first
        for img in images:
            content.append({
                "type": "image_url",
                "image_url": {
                    "url": f"data:image/png;base64,{img.get('base64', img.get('data', ''))}"
                },
            })

        # Add text prompt
        content.append({
            "type": "text",
            "text": prompt,
        })

        return [{"role": "user", "content": content}]

    @property
    def last_result(self) -> BaseModel | None:
        """Get the last result from this agent."""
        return getattr(self, "_last_result", None)
