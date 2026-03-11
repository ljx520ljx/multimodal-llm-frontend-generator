"""FastAPI application entry point."""

from __future__ import annotations

import logging
import sys
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.config import get_settings
from app.routes import chat, echo, generate
from app.state import get_checkpoint_manager, set_checkpoint_manager
from checkpoint.manager import CheckpointManager

# Configure logging
settings = get_settings()
logging.basicConfig(
    level=getattr(logging, settings.log_level),
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[logging.StreamHandler(sys.stdout)],
)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan handler."""
    logger.info(f"Starting Agent Core on port {settings.agent_port}")
    logger.info(f"Log level: {settings.log_level}")

    # Create shared CheckpointManager with backend URL and optional auth token
    mgr = CheckpointManager(
        backend_url=settings.backend_url,
        internal_token=settings.internal_api_token,
    )
    set_checkpoint_manager(mgr)
    logger.info("CheckpointManager initialized")

    yield

    # Cleanup: close the httpx client
    checkpoint_mgr = get_checkpoint_manager()
    if checkpoint_mgr is not None:
        await checkpoint_mgr.close()
        logger.info("CheckpointManager closed")
    logger.info("Shutting down Agent Core")


app = FastAPI(
    title="Agent Core",
    description="Python Agent Service for Multi-Agent Architecture",
    version="0.1.0",
    lifespan=lifespan,
)

# Configure CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.cors_origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/health")
async def health_check():
    """Health check endpoint."""
    return {"status": "ok"}


@app.get("/")
async def root():
    """Root endpoint with service info."""
    return {
        "service": "agent-core",
        "version": "0.1.0",
        "status": "running",
    }


# Register routers
app.include_router(echo.router)
app.include_router(generate.router, prefix="/api/v1", tags=["generate"])
app.include_router(chat.router, prefix="/api/v1", tags=["chat"])
