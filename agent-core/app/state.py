"""Application-level shared state (avoids circular imports with main.py)."""

from __future__ import annotations

from typing import Optional

from checkpoint.manager import CheckpointManager

# Shared CheckpointManager instance (set during lifespan in main.py)
_checkpoint_manager: Optional[CheckpointManager] = None


def set_checkpoint_manager(mgr: CheckpointManager) -> None:
    """Set the shared CheckpointManager (called from lifespan)."""
    global _checkpoint_manager
    _checkpoint_manager = mgr


def get_checkpoint_manager() -> Optional[CheckpointManager]:
    """Get the shared CheckpointManager instance."""
    return _checkpoint_manager
