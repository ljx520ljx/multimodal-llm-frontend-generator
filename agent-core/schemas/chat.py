"""Chat-related schemas for ChatAgent."""

from typing import Any, Optional

from pydantic import BaseModel, Field

from schemas.common import ImageData


class ChatMessage(BaseModel):
    """A message in the chat history."""

    role: str = Field(description="Message role: 'user' or 'assistant'")
    content: str = Field(description="Message content")


class ChatRequest(BaseModel):
    """Request schema for chat endpoint."""

    session_id: str = Field(description="Session ID")
    message: str = Field(description="User message / modification request")
    current_code: str = Field(description="Current HTML code to modify")
    images: list[ImageData] = Field(
        default_factory=list,
        description="Original design images (auto-attached by Go backend from session)",
    )
    history: list[ChatMessage] = Field(
        default_factory=list,
        description="Conversation history",
    )


class ToolCallData(BaseModel):
    """Data for a tool call event."""

    tool: str = Field(description="Tool name")
    args: dict[str, Any] = Field(default_factory=dict, description="Tool arguments")


class ToolResultData(BaseModel):
    """Data for a tool result event."""

    tool: str = Field(description="Tool name")
    result: dict[str, Any] = Field(default_factory=dict, description="Tool execution result")
