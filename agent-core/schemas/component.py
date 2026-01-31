"""Component detection schemas."""

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
    """Component detection result."""

    components: list[Component] = Field(
        description="List of detected components"
    )
    summary: str = Field(
        default="",
        description="Brief summary of the components"
    )
