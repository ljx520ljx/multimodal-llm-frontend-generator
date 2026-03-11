"""Component detection schemas."""

from __future__ import annotations

from typing import Any

from pydantic import BaseModel, Field


class Component(BaseModel):
    """A UI component detected in the design."""

    id: str = Field(description="Component ID (unique within the design)")
    type: str = Field(
        description="Component type: button, input, card, nav, list, image, text, icon, etc."
    )
    name: str = Field(
        description="Component name for code reference: SearchButton, UserCard, etc."
    )
    region: str = Field(
        description="Region this component belongs to"
    )
    properties: dict[str, Any] = Field(
        default_factory=dict,
        description="Component properties: text, variant, size, icon, etc."
    )


class ComponentList(BaseModel):
    """Component detection result.

    Field order: summary before components to avoid LLM output truncation.
    """

    summary: str = Field(
        default="",
        description="One-sentence summary of detected components, e.g. '识别出导航栏、搜索框、商品卡片等12个组件'"
    )
    components: list[Component] = Field(
        description="List of detected components"
    )
