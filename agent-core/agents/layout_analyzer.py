"""Layout Analyzer Agent - Analyzes UI layout structure from design images."""

from typing import Type

from pydantic import BaseModel

from agents.base import BaseAgent
from llm.prompts import LAYOUT_ANALYZER_PROMPT
from schemas.layout import LayoutInfo


class LayoutAnalyzerAgent(BaseAgent):
    """Agent that analyzes the overall layout structure of design images."""

    @property
    def name(self) -> str:
        return "LayoutAnalyzer"

    @property
    def description(self) -> str:
        return "分析设计稿的整体布局结构，识别区域划分和网格系统"

    def build_prompt(self, **kwargs) -> str:
        """Build the layout analysis prompt.

        Returns:
            The layout analyzer prompt template
        """
        return LAYOUT_ANALYZER_PROMPT

    def get_output_schema(self) -> Type[BaseModel]:
        """Get the LayoutInfo schema for output.

        Returns:
            LayoutInfo Pydantic model
        """
        return LayoutInfo
