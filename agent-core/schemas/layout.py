"""Layout analysis schemas."""

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
    """Layout analysis result."""

    structure: str = Field(
        description="Layout structure type: sidebar-main, header-body-footer, single-column, etc."
    )
    regions: list[Region] = Field(
        description="List of identified regions"
    )
    grid_system: Literal["flex", "grid"] = Field(
        default="flex",
        description="Recommended grid system"
    )
    description: str = Field(
        default="",
        description="Brief description of the layout"
    )


# Alias for backward compatibility
LayoutInfo = LayoutSchema
