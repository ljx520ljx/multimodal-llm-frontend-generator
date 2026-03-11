"""Chat API route - Handles conversational code modification requests."""

import logging
from typing import AsyncIterator

from fastapi import APIRouter, Depends
from fastapi.responses import StreamingResponse

from agents import ChatAgent
from app.dependencies import get_llm_gateway
from llm.gateway import LLMGateway
from schemas.chat import ChatRequest
from schemas.common import ImageData, SSEEvent, SSEEventType

logger = logging.getLogger(__name__)
router = APIRouter()


def get_chat_llm_gateway() -> LLMGateway:
    """Get LLM gateway configured for ChatAgent.

    This may use a different model than the default if CHAT_AGENT_MODEL is configured.
    """
    return get_llm_gateway("chat")


async def chat_event_generator(
    request: ChatRequest,
    llm: LLMGateway,
) -> AsyncIterator[str]:
    """Generate SSE events for chat-based code modification.

    Args:
        request: Chat request with message, current_code, images, history
        llm: LLM gateway instance (injected)

    Yields:
        SSE formatted strings
    """
    try:
        # Create ChatAgent
        agent = ChatAgent(llm)

        # Convert request images to ImageData objects
        images = [
            ImageData(id=img.id, base64=img.base64, order=img.order)
            for img in request.images
        ]

        # Run agent and stream events
        async for event in agent.run(
            message=request.message,
            current_code=request.current_code,
            images=images,
            history=request.history,
        ):
            yield event.to_sse()

    except Exception as e:
        logger.exception(f"Chat error for session {request.session_id}")
        error_event = SSEEvent(
            event=SSEEventType.ERROR,
            data={"error": str(e)},
        )
        yield error_event.to_sse()

        # Send done event even on error
        done_event = SSEEvent(
            event=SSEEventType.DONE,
            data={"success": False},
        )
        yield done_event.to_sse()


@router.post("/chat")
async def chat(
    request: ChatRequest,
    llm: LLMGateway = Depends(get_chat_llm_gateway),
) -> StreamingResponse:
    """Chat-based code modification.

    This endpoint accepts a user message and current code, then returns
    modified code based on the user's request. Supports tool calling
    for code validation.

    Request body:
    - session_id: Session identifier
    - message: User's modification request
    - current_code: Current HTML code to modify
    - images: Original design images (auto-attached by Go backend)
    - history: Conversation history

    Returns SSE stream with events:
    - thinking: Agent's thought process
    - tool_call: Tool being called
    - tool_result: Tool execution result
    - code: Modified code
    - error: Error occurred
    - done: Processing complete
    """
    logger.info(
        f"Chat request: session={request.session_id}, "
        f"message={request.message[:50]}..., "
        f"images={len(request.images)}"
    )

    return StreamingResponse(
        chat_event_generator(request, llm),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "X-Accel-Buffering": "no",  # Disable nginx buffering
        },
    )
