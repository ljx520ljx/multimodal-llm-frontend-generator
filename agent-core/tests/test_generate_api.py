"""Integration tests for Generate API endpoint.

Uses FastAPI's dependency_overrides to properly mock LLM gateway dependency.
"""

import json
from typing import AsyncIterator
from unittest.mock import MagicMock, patch

import pytest
from fastapi.testclient import TestClient

from app.dependencies import get_llm_gateway
from app.main import app
from schemas.common import SSEEvent, SSEEventType


# Create a minimal 1x1 red pixel PNG for testing
TEST_IMAGE_BASE64 = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8DwHwAFBQIAX8jx0gAAAABJRU5ErkJggg=="


def create_mock_llm_gateway():
    """Create a mock LLM gateway for testing."""
    mock = MagicMock()
    mock.chat = MagicMock(return_value="mocked response")
    mock.chat_structured = MagicMock(return_value=MagicMock())
    return mock


@pytest.fixture
def test_client():
    """Create a test client with mocked LLM gateway dependency."""
    # Use FastAPI's dependency_overrides to mock the LLM gateway
    mock_llm = create_mock_llm_gateway()
    app.dependency_overrides[get_llm_gateway] = lambda: mock_llm

    client = TestClient(app)
    yield client

    # Clean up dependency overrides after test
    app.dependency_overrides.clear()


class TestGenerateEndpoint:
    """Tests for /api/v1/generate endpoint."""

    def test_missing_images(self, test_client: TestClient):
        """Test error when no images provided."""
        response = test_client.post(
            "/api/v1/generate",
            json={
                "session_id": "test-session-1",
                "images": [],
                "options": {},
            },
        )

        assert response.status_code == 400
        assert "image" in response.json()["detail"].lower()

    def test_request_validation(self, test_client: TestClient):
        """Test request body validation."""
        # Missing session_id
        response = test_client.post(
            "/api/v1/generate",
            json={
                "images": [{"id": "1", "base64": TEST_IMAGE_BASE64, "order": 0}],
                "options": {},
            },
        )

        assert response.status_code == 422  # Validation error

    def test_generate_returns_stream(self, test_client: TestClient):
        """Test that generate endpoint returns SSE stream."""
        with patch("app.routes.generate.create_generate_workflow") as mock_workflow:
            # Mock the workflow to return a simple event sequence
            async def mock_run(*args, **kwargs) -> AsyncIterator[SSEEvent]:
                yield SSEEvent(
                    event=SSEEventType.AGENT_START,
                    data={"agent": "layout_analyzer", "status": "running"},
                )
                yield SSEEvent(
                    event=SSEEventType.DONE,
                    data={"success": True},
                )

            mock_instance = MagicMock()
            mock_instance.run = mock_run
            mock_workflow.return_value = mock_instance

            response = test_client.post(
                "/api/v1/generate",
                json={
                    "session_id": "test-session-2",
                    "images": [{"id": "1", "base64": TEST_IMAGE_BASE64, "order": 0}],
                    "options": {},
                },
            )

            assert response.status_code == 200
            assert response.headers["content-type"] == "text/event-stream; charset=utf-8"


class TestGenerateSyncEndpoint:
    """Tests for /api/v1/generate/sync endpoint."""

    def test_missing_images_sync(self, test_client: TestClient):
        """Test error when no images provided for sync endpoint."""
        response = test_client.post(
            "/api/v1/generate/sync",
            json={
                "session_id": "test-session-3",
                "images": [],
                "options": {},
            },
        )

        assert response.status_code == 400

    def test_generate_sync_returns_json(self, test_client: TestClient):
        """Test that sync endpoint returns JSON response."""
        with patch("app.routes.generate.create_generate_workflow") as mock_workflow:
            async def mock_run(*args, **kwargs) -> AsyncIterator[SSEEvent]:
                yield SSEEvent(
                    event=SSEEventType.CODE,
                    data={"html": "<html></html>"},
                )
                yield SSEEvent(
                    event=SSEEventType.DONE,
                    data={"success": True},
                )

            mock_instance = MagicMock()
            mock_instance.run = mock_run
            mock_workflow.return_value = mock_instance

            response = test_client.post(
                "/api/v1/generate/sync",
                json={
                    "session_id": "test-session-4",
                    "images": [{"id": "1", "base64": TEST_IMAGE_BASE64, "order": 0}],
                    "options": {},
                },
            )

            assert response.status_code == 200
            data = response.json()
            assert "success" in data
            assert "html" in data


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
