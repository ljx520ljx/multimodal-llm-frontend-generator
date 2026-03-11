"""Unit tests for DesignState and AgentError."""

import importlib.util
import json
import sys
from datetime import datetime, timezone
from pathlib import Path

import pytest

# Load graph.state directly from file to avoid graph/__init__.py
# which imports generate_workflow and triggers heavy dependency chain (httpx, etc.)
_state_path = Path(__file__).resolve().parent.parent / "graph" / "state.py"
_spec = importlib.util.spec_from_file_location("graph.state", _state_path)
_state_mod = importlib.util.module_from_spec(_spec)
sys.modules["graph.state"] = _state_mod
_spec.loader.exec_module(_state_mod)
AgentError = _state_mod.AgentError
DesignState = _state_mod.DesignState


class TestAgentError:
    """Tests for AgentError model."""

    def test_creation(self):
        err = AgentError(agent_name="code_generator", error="LLM timeout")
        assert err.agent_name == "code_generator"
        assert err.error == "LLM timeout"
        assert err.recoverable is False
        assert isinstance(err.timestamp, datetime)

    def test_recoverable_error(self):
        err = AgentError(
            agent_name="code_generator",
            error="Rate limited",
            recoverable=True,
        )
        assert err.recoverable is True

    def test_serialization(self):
        err = AgentError(agent_name="test", error="fail")
        data = err.model_dump()
        assert data["agent_name"] == "test"
        assert data["error"] == "fail"
        assert "timestamp" in data


class TestDesignState:
    """Tests for DesignState model."""

    def test_minimal_creation(self):
        state = DesignState(session_id="test-session")
        assert state.session_id == "test-session"
        assert state.images == []
        assert state.completed_agents == []
        assert state.checkpoints == {}
        assert state.errors == []
        assert state.current_agent is None
        assert state.success is False

    def test_full_creation(self):
        state = DesignState(
            session_id="s1",
            images=[{"id": "img1", "base64": "data:..."}],
            options={"max_retries": 5},
            max_retries=5,
        )
        assert len(state.images) == 1
        assert state.options["max_retries"] == 5
        assert state.max_retries == 5


class TestMarkAgentCompleted:
    """Tests for mark_agent_completed method."""

    def test_mark_single_agent(self):
        state = DesignState(session_id="s1")
        state.mark_agent_completed("layout_analyzer")
        assert "layout_analyzer" in state.completed_agents
        assert "layout_analyzer" in state.checkpoints
        assert state.current_agent is None

    def test_mark_multiple_agents(self):
        state = DesignState(session_id="s1")
        state.mark_agent_completed("layout_analyzer")
        state.mark_agent_completed("component_detector")
        assert len(state.completed_agents) == 2
        assert len(state.checkpoints) == 2

    def test_mark_duplicate_agent(self):
        state = DesignState(session_id="s1")
        state.mark_agent_completed("layout_analyzer")
        state.mark_agent_completed("layout_analyzer")
        assert state.completed_agents.count("layout_analyzer") == 1

    def test_clears_current_agent(self):
        state = DesignState(session_id="s1", current_agent="layout_analyzer")
        state.mark_agent_completed("layout_analyzer")
        assert state.current_agent is None


class TestGetNextAgent:
    """Tests for get_next_agent method."""

    def test_first_agent(self):
        state = DesignState(session_id="s1")
        pipeline = ["layout_analyzer", "component_detector", "code_generator"]
        assert state.get_next_agent(pipeline) == "layout_analyzer"

    def test_skip_completed(self):
        state = DesignState(session_id="s1", completed_agents=["layout_analyzer"])
        pipeline = ["layout_analyzer", "component_detector", "code_generator"]
        assert state.get_next_agent(pipeline) == "component_detector"

    def test_all_completed(self):
        state = DesignState(
            session_id="s1",
            completed_agents=["layout_analyzer", "component_detector", "code_generator"],
        )
        pipeline = ["layout_analyzer", "component_detector", "code_generator"]
        assert state.get_next_agent(pipeline) is None

    def test_empty_pipeline(self):
        state = DesignState(session_id="s1")
        assert state.get_next_agent([]) is None


class TestSetAgentOutput:
    """Tests for set_agent_output method."""

    def test_set_layout_info(self):
        state = DesignState(session_id="s1")
        mock_output = {"pages": [{"id": "p1"}]}
        state.set_agent_output("layout_analyzer", mock_output)
        assert state.layout_info == mock_output

    def test_set_component_info(self):
        state = DesignState(session_id="s1")
        mock_output = {"components": []}
        state.set_agent_output("component_detector", mock_output)
        assert state.component_info == mock_output

    def test_set_interaction_info(self):
        state = DesignState(session_id="s1")
        mock_output = {"interactions": []}
        state.set_agent_output("interaction_infer", mock_output)
        assert state.interaction_info == mock_output

    def test_set_generated_code(self):
        state = DesignState(session_id="s1")
        mock_output = {"html": "<div>hello</div>"}
        state.set_agent_output("code_generator", mock_output)
        assert state.generated_code == mock_output

    def test_unknown_agent_ignored(self):
        state = DesignState(session_id="s1")
        state.set_agent_output("unknown_agent", {"data": "test"})
        # Should not raise, just silently ignore
        assert state.layout_info is None

    def test_class_name_alias(self):
        """set_agent_output should also work with class-style names."""
        state = DesignState(session_id="s1")
        state.set_agent_output("LayoutAnalyzer", {"pages": []})
        assert state.layout_info == {"pages": []}


class TestCheckpointSerialization:
    """Tests for to_checkpoint_dict and from_checkpoint."""

    def test_to_checkpoint_excludes_images(self):
        state = DesignState(
            session_id="s1",
            images=[{"id": "img1", "base64": "very-large-data"}],
            completed_agents=["layout_analyzer"],
        )
        data = state.to_checkpoint_dict()
        assert "images" not in data
        assert data["session_id"] == "s1"
        assert data["completed_agents"] == ["layout_analyzer"]

    def test_from_checkpoint_restores_state(self):
        original = DesignState(
            session_id="s1",
            completed_agents=["layout_analyzer", "component_detector"],
            checkpoints={"layout_analyzer": "2025-01-01T00:00:00+00:00"},
        )
        checkpoint_data = original.to_checkpoint_dict()
        images = [{"id": "img1", "base64": "data"}]

        restored = DesignState.from_checkpoint(checkpoint_data, images)
        assert restored.session_id == "s1"
        assert restored.completed_agents == ["layout_analyzer", "component_detector"]
        assert len(restored.images) == 1
        assert restored.images[0]["id"] == "img1"

    def test_roundtrip_json_serialization(self):
        state = DesignState(
            session_id="s1",
            completed_agents=["layout_analyzer"],
            errors=[AgentError(agent_name="test", error="fail")],
        )
        checkpoint = state.to_checkpoint_dict()
        json_str = json.dumps(checkpoint, default=str)
        loaded = json.loads(json_str)
        images = [{"id": "img1"}]
        restored = DesignState.from_checkpoint(loaded, images)
        assert restored.session_id == "s1"
        assert restored.completed_agents == ["layout_analyzer"]


class TestBackwardCompatibility:
    """Tests for PipelineState backward compatibility."""

    def test_pipeline_state_alias(self):
        from graph.state import PipelineState
        assert PipelineState is DesignState
