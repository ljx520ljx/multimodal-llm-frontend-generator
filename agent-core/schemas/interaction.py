"""Interaction inference schemas (state machine model)."""

from __future__ import annotations

from typing import Literal

from pydantic import BaseModel, Field


class State(BaseModel):
    """A state in the state machine (corresponds to a design image).

    scope distinguishes two essentially different kinds of UI state:
    - "page": cross-page / cross-view navigation (Tab, Segmented, route-level
      switches). Implemented via a top-level ``currentState === 'id'`` guard.
    - "component": in-component micro-state (Slider tooltip, Form validation,
      Collapse panel, Popover visibility). Implemented via a local Alpine
      ``x-data`` boolean / numeric variable — NOT a currentState literal.
    """

    id: str = Field(description="State ID: home, search, product, etc.")
    name: str = Field(description="State display name: 首页, 搜索页, 商品页")
    image_index: int = Field(description="Index of the corresponding design image (0-based)")
    description: str = Field(default="", description="Brief description of this state")
    scope: Literal["page", "component"] = Field(
        default="page",
        description=(
            "Interaction scope. 'page' for cross-view navigation; 'component' "
            "for in-component micro-state (tooltip / validation / panel toggle)"
        ),
    )


class Transition(BaseModel):
    """A transition between states.

    scope follows the same semantics as State.scope and determines how the
    transition should be implemented and validated. Prefer marking the
    transition scope to match the scope of its states.
    """

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
    scope: Literal["page", "component"] = Field(
        default="page",
        description=(
            "Scope of this transition; 'component' transitions are implemented "
            "as local variable mutations and not required to use "
            "currentState = 'id' literal"
        ),
    )


class InteractionSpec(BaseModel):
    """Interaction inference result (state machine specification).

    Key concept: Transitions are NOT linear (1→2→3).
    Users can freely navigate: 1→2→1→3→1→2→3→2→1...

    Field order matters: summary and initial_state are placed before
    large array fields (states, transitions) to avoid LLM output
    truncation when token budget is tight.
    """

    summary: str = Field(
        default="",
        description="One-sentence summary of the interaction flow, e.g. '用户可在首页、搜索页和商品页间自由切换'"
    )
    initial_state: str = Field(description="Initial state ID")
    states: list[State] = Field(description="All states (one per design image)")
    transitions: list[Transition] = Field(
        description="All state transitions (should cover navigation between all states)"
    )
