"""Interaction inference schemas (state machine model)."""

from __future__ import annotations

from typing import Literal

from pydantic import BaseModel, Field


class State(BaseModel):
    """A state in the state machine (corresponds to a design image)."""

    id: str = Field(description="State ID: home, search, product, etc.")
    name: str = Field(description="State display name: 首页, 搜索页, 商品页")
    image_index: int = Field(description="Index of the corresponding design image (0-based)")
    description: str = Field(default="", description="Brief description of this state")


class Transition(BaseModel):
    """A transition between states."""

    from_state: str = Field(description="Source state ID")
    to_state: str = Field(description="Target state ID")
    trigger: str = Field(description="Trigger component ID or selector")
    trigger_event: Literal["click", "hover", "focus"] = Field(
        default="click",
        description="Event type that triggers the transition"
    )
    description: str = Field(
        default="",
        description="Description: 点击搜索按钮进入搜索页"
    )


class InteractionSpec(BaseModel):
    """Interaction inference result (state machine specification).

    Key concept: Transitions are NOT linear (1→2→3).
    Users can freely navigate: 1→2→1→3→1→2→3→2→1...
    """

    states: list[State] = Field(description="All states (one per design image)")
    transitions: list[Transition] = Field(
        description="All state transitions (should cover navigation between all states)"
    )
    initial_state: str = Field(description="Initial state ID")
    summary: str = Field(
        default="",
        description="Brief summary of the interaction flow"
    )
