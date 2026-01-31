"""Generate workflow using LangGraph StateGraph."""

import logging
from typing import Any, AsyncIterator, Callable

from langgraph.graph import END, StateGraph

from agents import (
    CodeGeneratorAgent,
    ComponentDetectorAgent,
    InteractionInferAgent,
    LayoutAnalyzerAgent,
)
from graph.state import PipelineState
from llm.gateway import LLMGateway
from schemas.common import SSEEvent, SSEEventType
from tools.code_validator import CodeValidator

logger = logging.getLogger(__name__)


class GenerateWorkflow:
    """LangGraph-based generation workflow.

    Pipeline:
    1. LayoutAnalyzer -> Analyze layout structure
    2. ComponentDetector -> Identify UI components
    3. InteractionInfer -> Build state machine
    4. CodeGenerator -> Generate HTML code
    5. CodeValidator -> Validate output (retry if needed)
    """

    def __init__(self, llm: LLMGateway, max_retries: int = 3):
        """Initialize the workflow.

        Args:
            llm: LLM gateway for agent calls
            max_retries: Maximum number of code generation retries
        """
        self.llm = llm
        self.max_retries = max_retries

        # Initialize agents
        self.layout_agent = LayoutAnalyzerAgent(llm)
        self.component_agent = ComponentDetectorAgent(llm)
        self.interaction_agent = InteractionInferAgent(llm)
        self.code_agent = CodeGeneratorAgent(llm)

        # Initialize validator
        self.validator = CodeValidator()

        # Build the graph
        self.graph = self._build_graph()

    def _build_graph(self) -> StateGraph:
        """Build the LangGraph StateGraph."""
        graph = StateGraph(PipelineState)

        # Add nodes
        graph.add_node("layout_analyzer", self._run_layout_analyzer)
        graph.add_node("component_detector", self._run_component_detector)
        graph.add_node("interaction_infer", self._run_interaction_infer)
        graph.add_node("code_generator", self._run_code_generator)
        graph.add_node("code_validator", self._run_code_validator)

        # Add edges - linear pipeline
        graph.set_entry_point("layout_analyzer")
        graph.add_edge("layout_analyzer", "component_detector")
        graph.add_edge("component_detector", "interaction_infer")
        graph.add_edge("interaction_infer", "code_generator")
        graph.add_edge("code_generator", "code_validator")

        # Conditional edge for validation result
        graph.add_conditional_edges(
            "code_validator",
            self._should_retry,
            {
                "retry": "code_generator",
                "success": END,
                "fail": END,
            },
        )

        return graph.compile()

    async def _run_layout_analyzer(self, state: PipelineState) -> dict[str, Any]:
        """Run layout analyzer agent."""
        logger.info("Running LayoutAnalyzer")

        # Consume events from agent
        async for _ in self.layout_agent.run(
            images=state.get("images"),
            stream_events=False,
        ):
            pass

        return {"layout_info": self.layout_agent.last_result}

    async def _run_component_detector(self, state: PipelineState) -> dict[str, Any]:
        """Run component detector agent."""
        logger.info("Running ComponentDetector")

        async for _ in self.component_agent.run(
            images=state.get("images"),
            layout_info=state.get("layout_info"),
            stream_events=False,
        ):
            pass

        return {"component_info": self.component_agent.last_result}

    async def _run_interaction_infer(self, state: PipelineState) -> dict[str, Any]:
        """Run interaction inference agent."""
        logger.info("Running InteractionInfer")

        async for _ in self.interaction_agent.run(
            images=state.get("images"),
            layout_info=state.get("layout_info"),
            component_info=state.get("component_info"),
            stream_events=False,
        ):
            pass

        return {"interaction_info": self.interaction_agent.last_result}

    async def _run_code_generator(self, state: PipelineState) -> dict[str, Any]:
        """Run code generator agent."""
        logger.info(f"Running CodeGenerator (retry: {state.get('retry_count', 0)})")

        async for _ in self.code_agent.run(
            images=state.get("images"),
            layout_info=state.get("layout_info"),
            component_info=state.get("component_info"),
            interaction_info=state.get("interaction_info"),
            validation_errors=state.get("validation_errors"),
            stream_events=False,
        ):
            pass

        return {"generated_code": self.code_agent.last_result}

    async def _run_code_validator(self, state: PipelineState) -> dict[str, Any]:
        """Run code validator."""
        logger.info("Running CodeValidator")

        generated_code = state.get("generated_code")
        if not generated_code or not generated_code.html:
            return {
                "validation_errors": ["生成的代码为空"],
                "success": False,
            }

        # Validate HTML
        result = self.validator.validate(generated_code.html)

        if result.is_valid:
            logger.info("Validation passed")
            return {
                "final_html": generated_code.html,
                "success": True,
                "validation_errors": [],
            }
        else:
            logger.warning(f"Validation failed: {result.errors}")
            return {
                "validation_errors": result.errors,
                "retry_count": state.get("retry_count", 0) + 1,
            }

    def _should_retry(self, state: PipelineState) -> str:
        """Determine whether to retry code generation."""
        if state.get("success"):
            return "success"

        retry_count = state.get("retry_count", 0)
        max_retries = state.get("max_retries", self.max_retries)

        if retry_count < max_retries:
            logger.info(f"Retrying code generation ({retry_count}/{max_retries})")
            return "retry"
        else:
            logger.error(f"Max retries reached ({max_retries})")
            return "fail"

    async def run(
        self,
        session_id: str,
        images: list[dict[str, Any]],
        options: dict[str, Any] | None = None,
    ) -> AsyncIterator[SSEEvent]:
        """Run the generation workflow with SSE streaming.

        Args:
            session_id: Session identifier
            images: List of image data with base64 content
            options: Optional generation options

        Yields:
            SSE events for workflow progress
        """
        options = options or {}

        # Initialize state
        initial_state: PipelineState = {
            "session_id": session_id,
            "images": images,
            "options": options,
            "layout_info": None,
            "component_info": None,
            "interaction_info": None,
            "generated_code": None,
            "validation_errors": [],
            "retry_count": 0,
            "max_retries": options.get("max_retries", self.max_retries),
            "final_html": None,
            "success": False,
            "error": None,
        }

        try:
            # Run the graph with streaming
            current_state = initial_state

            # Stream through each node
            async for event in self._stream_workflow(current_state):
                yield event

        except Exception as e:
            logger.exception("Workflow error")
            yield SSEEvent(
                event=SSEEventType.ERROR,
                data={"error": str(e)},
            )

    async def _stream_workflow(
        self,
        state: PipelineState,
    ) -> AsyncIterator[SSEEvent]:
        """Stream workflow execution with SSE events.

        This runs each agent and yields events for UI updates.
        """
        # Run LayoutAnalyzer
        yield SSEEvent(
            event=SSEEventType.AGENT_START,
            data={"agent": "LayoutAnalyzer", "description": self.layout_agent.description},
        )

        async for event in self.layout_agent.run(images=state.get("images")):
            yield event

        state["layout_info"] = self.layout_agent.last_result

        # Run ComponentDetector
        yield SSEEvent(
            event=SSEEventType.AGENT_START,
            data={"agent": "ComponentDetector", "description": self.component_agent.description},
        )

        async for event in self.component_agent.run(
            images=state.get("images"),
            layout_info=state.get("layout_info"),
        ):
            yield event

        state["component_info"] = self.component_agent.last_result

        # Run InteractionInfer
        yield SSEEvent(
            event=SSEEventType.AGENT_START,
            data={"agent": "InteractionInfer", "description": self.interaction_agent.description},
        )

        async for event in self.interaction_agent.run(
            images=state.get("images"),
            layout_info=state.get("layout_info"),
            component_info=state.get("component_info"),
        ):
            yield event

        state["interaction_info"] = self.interaction_agent.last_result

        # Run CodeGenerator with retry loop
        max_retries = state.get("max_retries", self.max_retries)
        retry_count = 0

        while retry_count <= max_retries:
            yield SSEEvent(
                event=SSEEventType.AGENT_START,
                data={
                    "agent": "CodeGenerator",
                    "description": self.code_agent.description,
                    "retry": retry_count,
                },
            )

            async for event in self.code_agent.run(
                images=state.get("images"),
                layout_info=state.get("layout_info"),
                component_info=state.get("component_info"),
                interaction_info=state.get("interaction_info"),
                validation_errors=state.get("validation_errors"),
            ):
                yield event

            state["generated_code"] = self.code_agent.last_result

            # Validate (including state machine checks if interaction_info available)
            if state["generated_code"] and state["generated_code"].html:
                result = self.validator.validate_full(
                    state["generated_code"].html,
                    state.get("interaction_info"),
                )

                if result.is_valid:
                    # Success!
                    yield SSEEvent(
                        event=SSEEventType.CODE,
                        data={"html": state["generated_code"].html},
                    )
                    yield SSEEvent(
                        event=SSEEventType.DONE,
                        data={"success": True},
                    )
                    return
                else:
                    # Validation failed
                    state["validation_errors"] = result.errors
                    yield SSEEvent(
                        event=SSEEventType.THINKING,
                        data={
                            "message": f"验证失败，重试中 ({retry_count + 1}/{max_retries})",
                            "errors": result.errors,
                        },
                    )
                    retry_count += 1
            else:
                state["validation_errors"] = ["生成的代码为空"]
                retry_count += 1

        # Max retries reached
        yield SSEEvent(
            event=SSEEventType.ERROR,
            data={
                "error": "达到最大重试次数",
                "validation_errors": state.get("validation_errors", []),
            },
        )
        yield SSEEvent(
            event=SSEEventType.DONE,
            data={"success": False},
        )


def create_generate_workflow(llm: LLMGateway, max_retries: int = 3) -> GenerateWorkflow:
    """Factory function to create a generate workflow.

    Args:
        llm: LLM gateway instance
        max_retries: Maximum code generation retries

    Returns:
        Configured GenerateWorkflow instance
    """
    return GenerateWorkflow(llm, max_retries)
