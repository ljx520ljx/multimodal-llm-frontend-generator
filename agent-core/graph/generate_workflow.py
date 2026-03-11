"""Generate workflow using DesignState and optional AgentRegistry."""

from __future__ import annotations

import json
import logging
from pathlib import Path
from typing import Any, AsyncIterator, Optional

from agents import (
    CodeGeneratorAgent,
    ComponentDetectorAgent,
    InteractionInferAgent,
    LayoutAnalyzerAgent,
)
from checkpoint.manager import CheckpointManager
from graph.state import DesignState
from llm.gateway import LLMGateway
from registry.agent_registry import AgentRegistry
from schemas.common import SSEEvent, SSEEventType
from tools.code_validator import CodeValidator

logger = logging.getLogger(__name__)

# Default configs directory relative to agent-core root
_DEFAULT_CONFIGS_DIR = Path(__file__).parent.parent / "agents" / "configs"


class GenerateWorkflow:
    """Generation workflow with DesignState, AgentRegistry, and checkpoint support.

    Pipeline:
    1. LayoutAnalyzer -> Analyze layout structure
    2. ComponentDetector -> Identify UI components
    3. InteractionInfer -> Build state machine
    4. CodeGenerator -> Generate HTML code
    5. CodeValidator -> Validate output (retry if needed)
    """

    def __init__(
        self,
        llm: LLMGateway,
        max_retries: int = 3,
        agent_llms: Optional[dict[str, LLMGateway]] = None,
        registry: Optional[AgentRegistry] = None,
        checkpoint_mgr: Optional[CheckpointManager] = None,
    ):
        """Initialize the workflow.

        Args:
            llm: Default LLM gateway for agent calls
            max_retries: Maximum number of code generation retries
            agent_llms: Optional per-agent LLM gateways
            registry: Optional AgentRegistry for YAML-driven configuration
            checkpoint_mgr: Optional CheckpointManager for state persistence
        """
        self.llm = llm
        self.max_retries = max_retries
        self.registry = registry
        self.checkpoint_mgr = checkpoint_mgr
        agent_llms = agent_llms or {}

        # Initialize agents with per-agent LLM gateways
        self.layout_agent = LayoutAnalyzerAgent(agent_llms.get("layout", llm))
        self.component_agent = ComponentDetectorAgent(agent_llms.get("component", llm))
        self.interaction_agent = InteractionInferAgent(agent_llms.get("interaction", llm))
        self.code_agent = CodeGeneratorAgent(agent_llms.get("codegen", llm))

        # Initialize validator
        self.validator = CodeValidator()

        # Agent map: YAML config name → agent instance (analysis agents only)
        self._agent_map: dict[str, Any] = {
            "layout_analyzer": self.layout_agent,
            "component_detector": self.component_agent,
            "interaction_infer": self.interaction_agent,
        }

        # Analysis pipeline order from registry or hardcoded default
        # Excludes code_generator/code_validator which have special retry logic
        _GENERATION_AGENTS = {"code_generator", "code_validator"}
        if registry:
            try:
                full_order = registry.get_pipeline_order()
                self._analysis_order = [
                    name for name in full_order if name not in _GENERATION_AGENTS
                ]
            except ValueError as e:
                logger.warning(f"Registry pipeline order failed: {e}, using default")
                self._analysis_order = list(self._agent_map.keys())
        else:
            self._analysis_order = list(self._agent_map.keys())

    async def run(
        self,
        session_id: str,
        images: list[dict[str, Any]],
        options: Optional[dict[str, Any]] = None,
        resume: bool = False,
    ) -> AsyncIterator[SSEEvent]:
        """Run the generation workflow with SSE streaming.

        Args:
            session_id: Session identifier
            images: List of image data with base64 content
            options: Optional generation options
            resume: If True, attempt to resume from checkpoint

        Yields:
            SSE events for workflow progress
        """
        options = options or {}

        # Initialize or restore state
        state: Optional[DesignState] = None

        if resume and self.checkpoint_mgr:
            try:
                checkpoint_json = await self.checkpoint_mgr.load(session_id)
                if checkpoint_json:
                    checkpoint_data = json.loads(checkpoint_json)
                    state = DesignState.from_checkpoint(checkpoint_data, images)
                    logger.info(
                        f"Resumed from checkpoint: completed_agents={state.completed_agents}"
                    )
                    yield SSEEvent(
                        event=SSEEventType.THINKING,
                        data={
                            "content": f"从断点恢复，已完成: {', '.join(state.completed_agents)}"
                        },
                    )
            except Exception as e:
                logger.warning(f"Failed to load checkpoint, starting fresh: {e}")

        if state is None:
            state = DesignState(
                session_id=session_id,
                images=images,
                options=options,
                max_retries=options.get("max_retries", self.max_retries),
            )

        try:
            async for event in self._stream_workflow(state):
                yield event
        except Exception as e:
            logger.exception("Workflow error")
            yield SSEEvent(
                event=SSEEventType.ERROR,
                data={"error": str(e)},
            )
            # Always emit DONE after ERROR so frontend properly transitions state
            yield SSEEvent(
                event=SSEEventType.DONE,
                data={"success": False},
            )

    def _build_agent_kwargs(self, agent_name: str, state: DesignState) -> dict[str, Any]:
        """Build kwargs for an analysis agent based on registry config or defaults.

        Maps the agent's declared input fields to DesignState values.
        """
        state_fields = {
            "layout_info": state.layout_info,
            "component_info": state.component_info,
            "interaction_info": state.interaction_info,
        }

        if self.registry:
            config = self.registry.get(agent_name)
            if config:
                # Use registry-declared input fields (excluding 'images' which is passed separately)
                return {
                    field: state_fields[field]
                    for field in config.input_fields
                    if field in state_fields
                }

        # Fallback: pass all available (non-None) state fields
        return {k: v for k, v in state_fields.items() if v is not None}

    async def _stream_workflow(
        self,
        state: DesignState,
    ) -> AsyncIterator[SSEEvent]:
        """Stream workflow execution with SSE events.

        Phase 1: Run analysis agents in registry-driven order (or default).
        Phase 2: Run CodeGenerator with retry/validation loop.
        """

        # Phase 1: Analysis agents (dynamic, driven by _analysis_order)
        for agent_name in self._analysis_order:
            agent = self._agent_map.get(agent_name)
            if not agent:
                logger.warning(f"No agent instance for '{agent_name}', skipping")
                continue

            # Check completion (support both PascalCase and snake_case for checkpoint compat)
            if agent.name in state.completed_agents or agent_name in state.completed_agents:
                continue

            state.current_agent = agent.name
            kwargs = self._build_agent_kwargs(agent_name, state)
            async for event in agent.run(images=state.images, **kwargs):
                yield event
            state.set_agent_output(agent.name, agent.last_result)
            state.mark_agent_completed(agent.name)
            await self._save_checkpoint(state)

        # Phase 2: Code generation with retry loop
        max_retries = state.max_retries
        retry_count = state.retry_count

        while retry_count < max_retries:
            if retry_count > 0:
                yield SSEEvent(
                    event=SSEEventType.THINKING,
                    data={"content": f"第 {retry_count}/{max_retries} 次重试代码生成..."},
                )

            state.current_agent = "CodeGenerator"
            async for event in self.code_agent.run(
                images=state.images,
                layout_info=state.layout_info,
                component_info=state.component_info,
                interaction_info=state.interaction_info,
                validation_errors=state.validation_errors,
            ):
                yield event

            state.generated_code = self.code_agent.last_result

            # Validate
            if state.generated_code and state.generated_code.html:
                result = self.validator.validate_full(
                    state.generated_code.html,
                    state.interaction_info,
                )

                if result.valid:
                    state.mark_agent_completed("CodeGenerator")
                    state.success = True
                    state.final_html = state.generated_code.html
                    await self._save_checkpoint(state)
                    yield SSEEvent(
                        event=SSEEventType.CODE,
                        data={"html": state.generated_code.html},
                    )
                    yield SSEEvent(
                        event=SSEEventType.DONE,
                        data={"success": True},
                    )
                    return
                else:
                    error_strings = [e.message for e in result.errors]
                    state.validation_errors = error_strings
                    yield SSEEvent(
                        event=SSEEventType.THINKING,
                        data={
                            "message": f"验证失败，重试中 ({retry_count + 1}/{max_retries})",
                            "errors": [e.model_dump() for e in result.errors],
                        },
                    )
                    retry_count += 1
                    state.retry_count = retry_count
            else:
                state.validation_errors = ["生成的代码为空"]
                retry_count += 1
                state.retry_count = retry_count

        # Max retries reached
        logger.warning(f"Max retries ({max_retries}) reached for code generation")
        last_code = state.generated_code
        if last_code and getattr(last_code, "html", None):
            state.final_html = last_code.html
            yield SSEEvent(
                event=SSEEventType.CODE,
                data={"html": last_code.html},
            )
        yield SSEEvent(
            event=SSEEventType.ERROR,
            data={
                "error": "达到最大重试次数，输出未通过完整验证",
                "validation_errors": state.validation_errors,
            },
        )
        yield SSEEvent(
            event=SSEEventType.DONE,
            data={"success": False},
        )

    async def _save_checkpoint(self, state: DesignState) -> None:
        """Save state checkpoint (non-blocking on failure)."""
        if not self.checkpoint_mgr:
            return
        try:
            state_json = json.dumps(state.to_checkpoint_dict(), ensure_ascii=False, default=str)
            await self.checkpoint_mgr.save(state.session_id, state_json)
        except Exception as e:
            logger.warning(f"Checkpoint save failed (non-blocking): {e}")


def create_generate_workflow(
    llm: LLMGateway,
    max_retries: int = 3,
    agent_llms: Optional[dict[str, LLMGateway]] = None,
    registry: Optional[AgentRegistry] = None,
    checkpoint_mgr: Optional[CheckpointManager] = None,
) -> GenerateWorkflow:
    """Factory function to create a generate workflow.

    Args:
        llm: Default LLM gateway instance
        max_retries: Maximum code generation retries
        agent_llms: Optional per-agent LLM gateways
        registry: Optional AgentRegistry
        checkpoint_mgr: Optional CheckpointManager

    Returns:
        Configured GenerateWorkflow instance
    """
    return GenerateWorkflow(llm, max_retries, agent_llms, registry, checkpoint_mgr)
