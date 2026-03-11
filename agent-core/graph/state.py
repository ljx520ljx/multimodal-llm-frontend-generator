"""Pipeline state definitions for the generation workflow.

DesignState is the unified Pydantic model that accumulates data as agents
run through the pipeline. It replaces the previous PipelineState TypedDict
with full type safety, serialization, and checkpoint support.
"""

from __future__ import annotations

from datetime import datetime, timezone
from typing import Any, Optional

from pydantic import BaseModel, Field

from schemas.code import GeneratedCode
from schemas.component import ComponentList
from schemas.interaction import InteractionSpec
from schemas.layout import LayoutInfo


class AgentError(BaseModel):
    """Records an error from a pipeline agent."""

    agent_name: str
    error: str
    recoverable: bool = False
    timestamp: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))


class DesignState(BaseModel):
    """Unified state model passed through the generation pipeline.

    Replaces the previous PipelineState TypedDict with:
    - Full Pydantic validation and serialization
    - Checkpoint metadata (completed_agents, checkpoints, errors)
    - Helper methods for pipeline orchestration
    """

    # Required fields
    session_id: str
    images: list[dict[str, Any]] = Field(default_factory=list)

    # Input options
    options: dict[str, Any] = Field(default_factory=dict)

    # Agent outputs (populated incrementally)
    layout_info: Optional[LayoutInfo] = None
    component_info: Optional[ComponentList] = None
    interaction_info: Optional[InteractionSpec] = None
    generated_code: Optional[GeneratedCode] = None

    # Validation state
    validation_errors: list[str] = Field(default_factory=list)
    retry_count: int = 0
    max_retries: int = 3

    # Final output
    final_html: Optional[str] = None
    success: bool = False
    error: Optional[str] = None

    # Checkpoint metadata
    current_agent: Optional[str] = None
    completed_agents: list[str] = Field(default_factory=list)
    checkpoints: dict[str, str] = Field(default_factory=dict)
    errors: list[AgentError] = Field(default_factory=list)

    def mark_agent_completed(self, name: str) -> None:
        """Mark an agent as completed and record checkpoint timestamp."""
        if name not in self.completed_agents:
            self.completed_agents.append(name)
        self.checkpoints[name] = datetime.now(timezone.utc).isoformat()
        self.current_agent = None

    def get_next_agent(self, pipeline_order: list[str]) -> Optional[str]:
        """Get the next agent to run, skipping completed ones.

        Args:
            pipeline_order: Ordered list of agent names.

        Returns:
            Name of next agent to run, or None if all completed.
        """
        for name in pipeline_order:
            if name not in self.completed_agents:
                return name
        return None

    def set_agent_output(self, agent_name: str, output: Any) -> None:
        """Store an agent's output in the appropriate field.

        Args:
            agent_name: The agent name (e.g., "layout_analyzer")
            output: The agent's result object
        """
        field_map = {
            "layout_analyzer": "layout_info",
            "LayoutAnalyzer": "layout_info",
            "component_detector": "component_info",
            "ComponentDetector": "component_info",
            "interaction_infer": "interaction_info",
            "InteractionInfer": "interaction_info",
            "code_generator": "generated_code",
            "CodeGenerator": "generated_code",
        }
        field = field_map.get(agent_name)
        if field:
            setattr(self, field, output)

    def to_checkpoint_dict(self) -> dict[str, Any]:
        """Serialize state for checkpoint storage, excluding images."""
        data = self.model_dump(exclude={"images"})
        # Convert Pydantic sub-models to dicts for JSON serialization
        return data

    @classmethod
    def from_checkpoint(cls, checkpoint_data: dict[str, Any], images: list[dict[str, Any]]) -> "DesignState":
        """Restore state from checkpoint data plus images.

        Args:
            checkpoint_data: Deserialized checkpoint dict
            images: Image data to inject (not stored in checkpoint)

        Returns:
            Restored DesignState instance
        """
        checkpoint_data["images"] = images
        return cls.model_validate(checkpoint_data)


# Backward compatibility alias
PipelineState = DesignState
