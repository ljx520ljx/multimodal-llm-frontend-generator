"""Generate API route - Handles code generation requests."""

from __future__ import annotations

import logging
from typing import Any, AsyncIterator, Optional

from fastapi import APIRouter, Depends, HTTPException
from fastapi.responses import StreamingResponse
from pydantic import BaseModel, Field

from app.config import get_settings
from app.dependencies import get_llm_gateway
from app.state import get_checkpoint_manager
from graph import create_generate_workflow
from llm.gateway import LLMGateway
from schemas.common import SSEEvent, SSEEventType

logger = logging.getLogger(__name__)
router = APIRouter()


def get_pipeline_llm_gateways() -> dict[str, LLMGateway]:
    """Get LLM gateways for pipeline agents.

    Returns a dict with "default" and per-agent gateways. Each agent gets
    its own LLM gateway with potentially different model and temperature
    configurations (set via LAYOUT_AGENT_MODEL, CODEGEN_AGENT_TEMPERATURE, etc.).

    Returns:
        Dict mapping agent type to LLMGateway instance.
        Always includes "default" key.
    """
    gateways: dict[str, LLMGateway] = {
        "default": get_llm_gateway("default"),
    }
    for agent_type in ("layout", "component", "interaction", "codegen"):
        gateways[agent_type] = get_llm_gateway(agent_type)
    return gateways


class GenerateRequest(BaseModel):
    """Request body for generate endpoint."""

    session_id: str = Field(..., description="Unique session identifier")
    images: list[dict[str, Any]] = Field(
        default_factory=list,
        description="List of image data with base64 content (optional if description provided)",
    )
    description: Optional[str] = Field(
        default=None,
        description="Text description for UI generation (alternative to images)",
    )
    options: dict[str, Any] = Field(
        default_factory=dict,
        description="Generation options (max_retries, stream, etc.)",
    )
    resume: bool = Field(
        default=False,
        description="If true, attempt to resume from last checkpoint",
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
    llm_gateways: dict[str, LLMGateway],
    resume: bool = False,
    description: Optional[str] = None,
) -> AsyncIterator[str]:
    """Generate SSE events for the workflow.

    Args:
        session_id: Session identifier
        images: List of image data
        options: Generation options
        llm_gateways: Dict of agent_type -> LLMGateway instances
        resume: Whether to resume from checkpoint
        description: Optional text description for text-to-UI generation

    Yields:
        SSE formatted strings
    """
    try:
        # Create workflow with per-agent LLM gateways and shared checkpoint manager
        max_retries = options.get("max_retries", get_settings().max_retries)
        default_llm = llm_gateways["default"]
        agent_llms = {k: v for k, v in llm_gateways.items() if k != "default"}
        workflow = create_generate_workflow(
            default_llm, max_retries, agent_llms,
            checkpoint_mgr=get_checkpoint_manager(),
        )

        # If text description is provided, inject it into options
        if description:
            options["description"] = description

        # Run workflow and stream events
        async for event in workflow.run(session_id, images, options, resume=resume):
            yield event.to_sse()

    except Exception as e:
        logger.exception(f"Generate error for session {session_id}")
        error_event = SSEEvent(
            event=SSEEventType.ERROR,
            data={"error": str(e)},
        )
        yield error_event.to_sse()


@router.post("/generate")
async def generate(
    request: GenerateRequest,
    llm_gateways: dict[str, LLMGateway] = Depends(get_pipeline_llm_gateways),
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
    logger.info(
        f"Generate request: session={request.session_id}, images={len(request.images)}, "
        f"description={'yes' if request.description else 'no'}"
    )

    if not request.images and not request.description:
        raise HTTPException(status_code=400, detail="At least one image or a description is required")

    return StreamingResponse(
        event_generator(
            session_id=request.session_id,
            images=request.images,
            options=request.options,
            llm_gateways=llm_gateways,
            resume=request.resume,
            description=request.description,
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
    llm_gateways: dict[str, LLMGateway] = Depends(get_pipeline_llm_gateways),
) -> GenerateResponse:
    """Generate code synchronously (non-streaming).

    Same as /generate but returns a single JSON response instead of SSE stream.
    Useful for testing or when streaming is not needed.
    """
    logger.info(f"Generate sync request: session={request.session_id}")

    if not request.images and not request.description:
        raise HTTPException(status_code=400, detail="At least one image or a description is required")

    try:
        # Create and run workflow with per-agent LLM gateways and shared checkpoint manager
        settings = get_settings()
        default_llm = llm_gateways["default"]
        agent_llms = {k: v for k, v in llm_gateways.items() if k != "default"}
        workflow = create_generate_workflow(
            default_llm,
            request.options.get("max_retries", settings.max_retries),
            agent_llms,
            checkpoint_mgr=get_checkpoint_manager(),
        )

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
