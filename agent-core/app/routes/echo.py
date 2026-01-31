"""Echo route for testing Go ↔ Python SSE communication."""

import asyncio
import json
from typing import AsyncGenerator

from fastapi import APIRouter
from fastapi.responses import StreamingResponse
from pydantic import BaseModel, Field

router = APIRouter(prefix="/api/v1", tags=["echo"])


class EchoRequest(BaseModel):
    """Echo request model."""

    message: str = Field(default="hello", description="Message to echo")
    count: int = Field(default=5, ge=1, le=20, description="Number of events to send")
    delay: float = Field(default=0.5, ge=0.1, le=5.0, description="Delay between events in seconds")


async def generate_sse_events(request: EchoRequest) -> AsyncGenerator[str, None]:
    """Generate SSE events for echo endpoint."""
    for i in range(request.count):
        event_data = {
            "index": i,
            "message": request.message,
            "total": request.count,
        }
        yield f"event: message\ndata: {json.dumps(event_data)}\n\n"
        await asyncio.sleep(request.delay)

    # Send done event
    yield f"event: done\ndata: {json.dumps({})}\n\n"


@router.post("/echo")
async def echo_stream(request: EchoRequest) -> StreamingResponse:
    """
    Echo endpoint that returns a Server-Sent Events stream.

    Used for testing Go ↔ Python SSE communication.
    """
    return StreamingResponse(
        generate_sse_events(request),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "X-Accel-Buffering": "no",
        },
    )
