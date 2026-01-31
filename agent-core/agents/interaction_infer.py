"""Interaction Infer Agent - Infers interaction logic and builds state machine."""

import json
from typing import Type

from pydantic import BaseModel

from agents.base import BaseAgent
from llm.prompts import INTERACTION_INFER_PROMPT
from schemas.component import ComponentList
from schemas.interaction import InteractionSpec
from schemas.layout import LayoutInfo


class InteractionInferAgent(BaseAgent):
    """Agent that infers interaction logic and builds state machine model."""

    @property
    def name(self) -> str:
        return "InteractionInfer"

    @property
    def description(self) -> str:
        return "分析设计稿之间的交互逻辑，构建状态机模型"

    def build_prompt(
        self,
        layout_info: LayoutInfo | None = None,
        component_info: ComponentList | None = None,
        **kwargs,
    ) -> str:
        """Build the interaction inference prompt.

        Args:
            layout_info: Layout information from LayoutAnalyzer
            component_info: Component information from ComponentDetector

        Returns:
            Formatted interaction inference prompt
        """
        layout_str = ""
        if layout_info:
            layout_str = json.dumps(layout_info.model_dump(), ensure_ascii=False, indent=2)

        component_str = ""
        if component_info:
            component_str = json.dumps(component_info.model_dump(), ensure_ascii=False, indent=2)

        return INTERACTION_INFER_PROMPT.format(
            layout_info=layout_str,
            component_info=component_str,
        )

    def get_output_schema(self) -> Type[BaseModel]:
        """Get the InteractionSpec schema for output.

        Returns:
            InteractionSpec Pydantic model
        """
        return InteractionSpec
