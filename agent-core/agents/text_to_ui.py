"""TextToUI Agent - Generates UI specifications from text descriptions.

Produces layout_info, component_info, and interaction_info in a single pass,
allowing the pipeline to skip the image-based analysis agents when only a
text description is provided.
"""

from typing import Optional, Type

from pydantic import BaseModel, Field

from agents.base import BaseAgent
from schemas.component import ComponentList
from schemas.interaction import InteractionSpec
from schemas.layout import LayoutInfo


class TextToUIOutput(BaseModel):
    """Combined output from the TextToUI agent."""

    layout_info: LayoutInfo = Field(description="Layout structure specification")
    component_info: ComponentList = Field(description="UI components specification")
    interaction_info: InteractionSpec = Field(description="Interaction state machine specification")


class TextToUIAgent(BaseAgent):
    """Agent that generates full UI specifications from a text description.

    Unlike the image-based pipeline (LayoutAnalyzer → ComponentDetector →
    InteractionInfer), this agent produces all three outputs in a single LLM
    call based on a text description.
    """

    @property
    def name(self) -> str:
        return "TextToUI"

    @property
    def description(self) -> str:
        return "根据文字描述生成完整的 UI 规格（布局 + 组件 + 交互）"

    def build_prompt(self, description: Optional[str] = None, **kwargs) -> str:
        """Build the text-to-UI prompt.

        Args:
            description: User's text description of the desired UI.

        Returns:
            Formatted prompt string.
        """
        desc = description or "一个简单的页面"
        return f"""你是一位资深 UI/UX 设计师和前端架构师。根据以下文字描述，生成详细的 UI 规格说明。

## 用户描述
{desc}

## 输出要求

请生成以下三个部分的 JSON 结构化输出：

### 1. layout_info
分析页面的整体布局结构：
- structure: 布局类型（sidebar-main, header-body-footer, single-column 等）
- regions: 各区域定义（name, position, estimated_size）
- grid_system: 推荐的网格系统（flex 或 grid）

### 2. component_info
列出页面中的所有 UI 组件：
- 每个组件包含 id, type, name, region, properties
- 组件类型包括: button, input, card, nav, list, image, text, icon 等

### 3. interaction_info
定义页面的交互状态机：
- states: 所有状态（至少包含一个默认状态）
- transitions: 状态间的转换（trigger, trigger_event, from_state, to_state）
- initial_state: 初始状态 ID

## 设计原则
- 组件设计应现代、美观，参考主流设计系统
- 交互逻辑应合理、完整
- 考虑移动端响应式
- 输出应足够详细，让代码生成器能直接使用"""

    def get_output_schema(self) -> Type[BaseModel]:
        return TextToUIOutput
