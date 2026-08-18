"""Focused tests for generation workflow terminal semantics."""

from unittest.mock import MagicMock

import pytest

from graph.generate_workflow import GenerateWorkflow
from graph.state import DesignState
from schemas.code import GeneratedCode, ValidationError, ValidationResult
from schemas.common import SSEEvent, SSEEventType


class _InvalidCodeAgent:
    def __init__(self) -> None:
        self.last_result = None

    async def run(self, **kwargs):
        self.last_result = GeneratedCode(html="<html><body>partial</body></html>")
        yield SSEEvent(
            event=SSEEventType.CODE,
            data={"html": self.last_result.html},
        )


@pytest.mark.asyncio
async def test_validation_exhaustion_is_reported_as_failure():
    workflow = GenerateWorkflow(MagicMock(), max_retries=1)
    workflow.code_agent = _InvalidCodeAgent()
    workflow.validator.validate_full = MagicMock(return_value=ValidationResult(
        valid=False,
        errors=[ValidationError(
            type="missing_state",
            message="缺少状态: search",
        )],
    ))

    state = DesignState(
        session_id="test-session",
        images=[],
        max_retries=1,
        completed_agents=["LayoutAnalyzer", "ComponentDetector", "InteractionInfer"],
    )

    events = [event async for event in workflow._stream_workflow(state)]

    assert [event.event for event in events][-2:] == [
        SSEEventType.ERROR,
        SSEEventType.DONE,
    ]
    assert events[-1].data == {"success": False}
    assert state.success is False
    assert state.final_html == "<html><body>partial</body></html>"
    assert state.validation_errors == ["缺少状态: search"]
