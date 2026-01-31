"""Pipeline state definition for LangGraph workflow."""

from typing import Any, Optional, TypedDict

from schemas.code import GeneratedCode
from schemas.component import ComponentList
from schemas.interaction import InteractionSpec
from schemas.layout import LayoutInfo


class PipelineState(TypedDict, total=False):
    """State passed through the generation pipeline.

    This TypedDict defines all data shared between agents in the workflow.
    Fields are optional (total=False) as they are populated incrementally.
    """

    # Input data
    session_id: str
    images: list[dict[str, Any]]  # List of image data with base64 content
    options: dict[str, Any]  # Generation options

    # Agent outputs
    layout_info: Optional[LayoutInfo]
    component_info: Optional[ComponentList]
    interaction_info: Optional[InteractionSpec]
    generated_code: Optional[GeneratedCode]

    # Validation state
    validation_errors: list[str]
    retry_count: int
    max_retries: int

    # Final output
    final_html: Optional[str]
    success: bool
    error: Optional[str]
