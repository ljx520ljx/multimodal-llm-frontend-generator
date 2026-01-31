"""Prompt templates for agents."""

from llm.prompts.layout import LAYOUT_ANALYZER_PROMPT
from llm.prompts.component import COMPONENT_DETECTOR_PROMPT
from llm.prompts.interaction import INTERACTION_INFER_PROMPT
from llm.prompts.generator import CODE_GENERATOR_PROMPT, CODE_GENERATOR_PROMPT_WITH_FEEDBACK

__all__ = [
    "LAYOUT_ANALYZER_PROMPT",
    "COMPONENT_DETECTOR_PROMPT",
    "INTERACTION_INFER_PROMPT",
    "CODE_GENERATOR_PROMPT",
    "CODE_GENERATOR_PROMPT_WITH_FEEDBACK",
]
