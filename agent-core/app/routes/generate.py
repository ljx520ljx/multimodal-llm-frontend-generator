"""Generate API route - Handles code generation requests."""

import json
import logging
from typing import Any, AsyncIterator, Optional

from fastapi import APIRouter, HTTPException
from fastapi.responses import StreamingResponse
from pydantic import BaseModel, Field

from app.config import get_settings
from app.dependencies import get_llm_gateway
from graph import create_generate_workflow
from llm.gateway import LLMGateway
from schemas.common import SSEEvent, SSEEventType

logger = logging.getLogger(__name__)
router = APIRouter()


class GenerateRequest(BaseModel):
    """Request body for generate endpoint."""

    session_id: str = Field(..., description="Unique session identifier")
    images: list[dict[str, Any]] = Field(..., description="List of image data with base64 content")
    options: dict[str, Any] = Field(
        default_factory=dict,
        description="Generation options (max_retries, stream, etc.)",
    )


class GenerateResponse(BaseModel):
    """Non-streaming response for generate endpoint."""

    success: bool
    html: Optional[str] = None
    error: Optional[str] = None


async def event_generator(
    session_id: str,
    images: list[dict[str, Any]],
    options: dict[str, Any],
    llm: LLMGateway,
) -> AsyncIterator[str]:
    """Generate SSE events for the workflow.

    Args:
        session_id: Session identifier
        images: List of image data
        options: Generation options
        llm: LLM gateway instance (injected)

    Yields:
        SSE formatted strings
    """
    try:
        # Create workflow with injected LLM gateway
        max_retries = options.get("max_retries", get_settings().max_retries)
        workflow = create_generate_workflow(llm, max_retries)

        # Run workflow and stream events
        async for event in workflow.run(session_id, images, options):
            yield format_sse(event)

    except Exception as e:
        logger.exception(f"Generate error for session {session_id}")
        error_event = SSEEvent(
            event=SSEEventType.ERROR,
            data={"error": str(e)},
        )
        yield format_sse(error_event)


def format_sse(event: SSEEvent) -> str:
    """Format an SSE event as a string.

    Args:
        event: SSE event object

    Returns:
        SSE formatted string
    """
    event_type = event.event.value if hasattr(event.event, "value") else str(event.event)
    data = json.dumps(event.data, ensure_ascii=False)
    return f"event: {event_type}\ndata: {data}\n\n"


@router.post("/generate")
async def generate(
    request: GenerateRequest,
) -> StreamingResponse:
    """Generate code from design images.

    This endpoint accepts design images and generates interactive HTML code
    using a multi-agent pipeline. Progress is streamed via SSE.

    Request body:
    - session_id: Unique session identifier
    - images: List of image data with base64 content
    - options: Generation options (max_retries, etc.)

    Returns SSE stream with events:
    - agent_start: Agent begins processing
    - agent_result: Agent completed with result
    - thinking: Intermediate thinking/progress
    - code: Generated code chunk
    - error: Error occurred
    - done: Generation complete
    """
    logger.info(f"Generate request: session={request.session_id}, images={len(request.images)}")

    # Validate images first before initializing LLM
    if not request.images:
        raise HTTPException(status_code=400, detail="At least one image is required")

    # Get LLM gateway after validation
    try:
        llm = get_llm_gateway()
    except Exception as e:
        logger.error(f"Failed to initialize LLM gateway: {e}")
        raise HTTPException(
            status_code=503,
            detail=f"LLM service unavailable: {str(e)}. Please configure API key."
        )

    return StreamingResponse(
        event_generator(
            session_id=request.session_id,
            images=request.images,
            options=request.options,
            llm=llm,
        ),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "X-Accel-Buffering": "no",  # Disable nginx buffering
        },
    )


@router.post("/generate/sync", response_model=GenerateResponse)
async def generate_sync(
    request: GenerateRequest,
) -> GenerateResponse:
    """Generate code synchronously (non-streaming).

    Same as /generate but returns a single JSON response instead of SSE stream.
    Useful for testing or when streaming is not needed.
    """
    logger.info(f"Generate sync request: session={request.session_id}")

    # Validate images first before initializing LLM
    if not request.images:
        raise HTTPException(status_code=400, detail="At least one image is required")

    # Get LLM gateway after validation
    try:
        llm = get_llm_gateway()
    except Exception as e:
        logger.error(f"Failed to initialize LLM gateway: {e}")
        raise HTTPException(
            status_code=503,
            detail=f"LLM service unavailable: {str(e)}. Please configure API key."
        )

    try:
        # Create and run workflow with injected LLM gateway
        settings = get_settings()
        workflow = create_generate_workflow(llm, request.options.get("max_retries", settings.max_retries))

        final_html = None
        success = False
        error = None

        async for event in workflow.run(
            request.session_id,
            request.images,
            request.options,
        ):
            if event.event == SSEEventType.CODE:
                final_html = event.data.get("html")
            elif event.event == SSEEventType.DONE:
                success = event.data.get("success", False)
            elif event.event == SSEEventType.ERROR:
                error = event.data.get("error")

        return GenerateResponse(
            success=success,
            html=final_html,
            error=error,
        )

    except Exception as e:
        logger.exception(f"Generate sync error for session {request.session_id}")
        return GenerateResponse(
            success=False,
            html=None,
            error=str(e),
        )
