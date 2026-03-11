"""Common schemas used across the pipeline."""

from __future__ import annotations

import json
from enum import Enum
from typing import Any

from pydantic import BaseModel, Field


class ImageData(BaseModel):
    """Image data for pipeline input."""

    id: str = Field(description="Image ID")
    base64: str = Field(description="Base64 encoded image data (data:image/...;base64,...)")
    order: int = Field(default=0, description="Sort order")


class SSEEventType(str, Enum):
    """SSE event types."""

    AGENT_START = "agent_start"
    AGENT_RESULT = "agent_result"
    THINKING = "thinking"
    CODE = "code"
    ERROR = "error"
    DONE = "done"
    # ChatAgent tool calling events
    TOOL_CALL = "tool_call"
    TOOL_RESULT = "tool_result"


class SSEEvent(BaseModel):
    """SSE event for streaming output."""

    event: SSEEventType = Field(description="Event type")
    data: dict[str, Any] = Field(default_factory=dict, description="Event payload")

    def to_sse(self) -> str:
        """Convert to SSE format string."""
        return f"event: {self.event.value}\ndata: {json.dumps(self.data, ensure_ascii=False, default=str)}\n\n"
