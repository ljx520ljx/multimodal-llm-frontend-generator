"""Integration tests for Generate API endpoint.

Note: API endpoint tests require proper mocking of LLM gateway which
is initialized at module import time. These tests focus on schema
validation and SSE event formatting.
"""

import json
from typing import AsyncIterator

import pytest

from schemas.common import SSEEvent, SSEEventType


class TestSSEEvent:
    """Tests for SSEEvent model."""

    def test_sse_event_creation(self):
        """Test creating an SSE event."""
        event = SSEEvent(
            event=SSEEventType.AGENT_START,
            data={"agent": "test", "status": "running"},
        )

        assert event.event == SSEEventType.AGENT_START
        assert event.data["agent"] == "test"

    def test_sse_event_format(self):
        """Test SSE event to_sse method."""
        event = SSEEvent(
            event=SSEEventType.AGENT_START,
            data={"agent": "test", "status": "running"},
        )

        sse_str = event.to_sse()

        assert sse_str.startswith("event: agent_start\n")
        assert "data:" in sse_str
        assert sse_str.endswith("\n\n")

    def test_sse_event_json_data(self):
        """Test that SSE data is valid JSON."""
        event = SSEEvent(
            event=SSEEventType.THINKING,
            data={"content": "处理中...", "progress": 50},
        )

        sse_str = event.to_sse()

        # Extract data line
        lines = sse_str.strip().split("\n")
        data_line = [l for l in lines if l.startswith("data:")][0]
        json_str = data_line[5:].strip()  # Remove "data:" prefix

        # Should be valid JSON
        parsed = json.loads(json_str)
        assert parsed["content"] == "处理中..."
        assert parsed["progress"] == 50

    def test_all_event_types(self):
        """Test all SSE event types."""
        event_types = [
            SSEEventType.AGENT_START,
            SSEEventType.AGENT_RESULT,
            SSEEventType.THINKING,
            SSEEventType.CODE,
            SSEEventType.ERROR,
            SSEEventType.DONE,
        ]

        for event_type in event_types:
            event = SSEEvent(event=event_type, data={"test": True})
            sse_str = event.to_sse()

            # Should contain the event type value
            assert f"event: {event_type.value}\n" in sse_str


class TestSSEEventType:
    """Tests for SSEEventType enum."""

    def test_event_type_values(self):
        """Test SSE event type string values."""
        assert SSEEventType.AGENT_START.value == "agent_start"
        assert SSEEventType.AGENT_RESULT.value == "agent_result"
        assert SSEEventType.THINKING.value == "thinking"
        assert SSEEventType.CODE.value == "code"
        assert SSEEventType.ERROR.value == "error"
        assert SSEEventType.DONE.value == "done"
