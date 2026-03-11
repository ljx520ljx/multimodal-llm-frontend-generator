"""CheckpointManager - Saves/restores DesignState via Go backend API."""

from __future__ import annotations

import logging
from typing import Optional

import httpx

logger = logging.getLogger(__name__)

# Default Go backend URL for internal checkpoint API
_DEFAULT_BACKEND_URL = "http://localhost:8080"


class CheckpointManager:
    """Manages pipeline state checkpoints through the Go backend.

    Communicates with the Go backend's internal checkpoint API to persist
    DesignState to PostgreSQL via the sessions.design_state JSONB column.

    Supports async context manager for proper resource cleanup:
        async with CheckpointManager() as mgr:
            await mgr.save(session_id, state_json)
    """

    def __init__(
        self,
        backend_url: str = _DEFAULT_BACKEND_URL,
        timeout: float = 10.0,
        internal_token: str = "",
    ):
        self._backend_url = backend_url.rstrip("/")
        self._timeout = timeout
        self._internal_token = internal_token
        self._client = httpx.AsyncClient(timeout=timeout)

    async def __aenter__(self) -> CheckpointManager:
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb) -> None:
        await self.close()

    async def save(self, session_id: str, state_json: str) -> bool:
        """Save design state checkpoint.

        Args:
            session_id: Session ID
            state_json: Serialized DesignState JSON string

        Returns:
            True if saved successfully, False on error (non-blocking)
        """
        try:
            url = f"{self._backend_url}/api/internal/checkpoint/{session_id}"
            headers: dict[str, str] = {"Content-Type": "application/json"}
            if self._internal_token:
                headers["X-Internal-Token"] = self._internal_token
            resp = await self._client.put(
                url,
                content=state_json,
                headers=headers,
            )
            if resp.status_code == 200:
                logger.debug(f"Checkpoint saved for session {session_id}")
                return True
            else:
                logger.warning(
                    f"Checkpoint save failed: status={resp.status_code}, body={resp.text}"
                )
                return False
        except Exception as e:
            logger.warning(f"Checkpoint save error for session {session_id}: {e}")
            return False

    async def load(self, session_id: str) -> Optional[str]:
        """Load design state checkpoint.

        Args:
            session_id: Session ID

        Returns:
            JSON string of DesignState, or None if not found/error
        """
        try:
            url = f"{self._backend_url}/api/internal/checkpoint/{session_id}"
            headers: dict[str, str] = {}
            if self._internal_token:
                headers["X-Internal-Token"] = self._internal_token
            resp = await self._client.get(url, headers=headers)
            if resp.status_code == 200:
                data = resp.json()
                state_json = data.get("design_state")
                if state_json:
                    logger.debug(f"Checkpoint loaded for session {session_id}")
                    return state_json if isinstance(state_json, str) else None
                return None
            else:
                return None
        except Exception as e:
            logger.warning(f"Checkpoint load error for session {session_id}: {e}")
            return None

    async def close(self) -> None:
        """Close the HTTP client."""
        await self._client.aclose()
