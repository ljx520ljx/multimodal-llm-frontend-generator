"""Layout analysis schemas."""

from __future__ import annotations

from typing import Literal

from pydantic import BaseModel, Field


class Region(BaseModel):
    """A region in the layout."""

    name: str = Field(description="Region name: header, sidebar, main, footer, etc.")
    position: Literal["top", "left", "right", "center", "bottom"] = Field(
        description="Position in the layout"
    )
    estimated_size: str = Field(
        description="Estimated Tailwind size: w-64, flex-1, h-16, etc."
    )


class LayoutSchema(BaseModel):
    """Layout analysis result.

    Field order: description before regions to avoid LLM output truncation.
    """

    description: str = Field(
        default="",
        description="One-sentence description of the layout, e.g. '左侧边栏+右侧主内容区的经典后台布局'"
    )
    structure: str = Field(
        description="Layout structure type: sidebar-main, header-body-footer, single-column, etc."
    )
    grid_system: Literal["flex", "grid"] = Field(
        default="flex",
        description="Recommended grid system"
    )
    regions: list[Region] = Field(
        description="List of identified regions"
    )


# Alias for backward compatibility
LayoutInfo = LayoutSchema
