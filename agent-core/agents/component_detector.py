"""Component Detector Agent - Identifies UI components from design images."""

import json
from typing import Optional, Type

from pydantic import BaseModel

from agents.base import BaseAgent
from llm.prompts import COMPONENT_DETECTOR_PROMPT
from schemas.component import ComponentList
from schemas.layout import LayoutInfo


class ComponentDetectorAgent(BaseAgent):
    """Agent that identifies UI components in design images."""

    @property
    def name(self) -> str:
        return "ComponentDetector"

    @property
    def description(self) -> str:
        return "识别设计稿中的所有 UI 组件，包括按钮、输入框、导航等"

    def build_prompt(self, layout_info: Optional[LayoutInfo] = None, **kwargs) -> str:
        """Build the component detection prompt.

        Args:
            layout_info: Layout information from LayoutAnalyzer

        Returns:
            Formatted component detector prompt
        """
        layout_str = ""
        if layout_info:
            layout_str = json.dumps(layout_info.model_dump(), ensure_ascii=False, indent=2)

        return COMPONENT_DETECTOR_PROMPT.format(layout_info=layout_str)

    def get_output_schema(self) -> Type[BaseModel]:
        """Get the ComponentList schema for output.

        Returns:
            ComponentList Pydantic model
        """
        return ComponentList
