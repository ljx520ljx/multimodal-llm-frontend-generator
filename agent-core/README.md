# Agent Core

Python Agent Service for Multi-Agent Architecture.

## Overview

This service handles AI Agent logic including:
- **Multi-Agent Pipeline**: 5 specialized agents for initial code generation
- **ChatAgent with Tool Calling**: Conversational code modification with validation tools
- LangGraph orchestration
- SSE streaming output

## Quick Start

### Prerequisites

- Python 3.9+
- pip

### Installation

```bash
cd agent-core
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install -r requirements.txt
```

### Configuration

Create `.env` file:

```bash
# LLM Configuration
LLM_PROVIDER=openai
OPENAI_API_KEY=sk-your-api-key
OPENAI_MODEL=gpt-4o
OPENAI_BASE_URL=  # Optional: custom API endpoint

# Server
AGENT_PORT=8081
LOG_LEVEL=INFO
```

### Running

```bash
uvicorn app.main:app --port 8081 --reload
```

### Health Check

```bash
curl http://localhost:8081/health
# Returns: {"status": "ok"}
```

## API Endpoints

### POST /api/v1/generate

Initial code generation using Multi-Agent Pipeline.

```bash
curl -N -X POST http://localhost:8081/api/v1/generate \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "uuid",
    "images": [{"id": "img1", "base64": "data:image/png;base64,...", "order": 0}],
    "options": {"max_retries": 3}
  }'
```

**SSE Events:**
- `agent_start`: Agent started processing
- `agent_result`: Agent completed with result
- `code`: Generated HTML code
- `done`: Processing complete

### POST /api/v1/chat

Conversational code modification with Tool Calling.

```bash
curl -N -X POST http://localhost:8081/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "uuid",
    "message": "把按钮改成蓝色",
    "current_code": "<!DOCTYPE html>...",
    "images": [],
    "history": []
  }'
```

**SSE Events:**
- `thinking`: Agent's thought process
- `tool_call`: Tool being called (validate_html, check_interaction)
- `tool_result`: Tool execution result
- `code`: Modified HTML code
- `done`: Processing complete

## Architecture

### Multi-Agent Pipeline (Initial Generation)

```
LayoutAnalyzer → ComponentDetector → InteractionInfer → CodeGenerator → CodeValidator
```

### ChatAgent (Conversational Modification)

```
User Message → ChatAgent → [Tool Calls] → Modified Code
                  ↓
            validate_html()
            check_interaction()
```

**Features:**
- Multi-turn tool calling loop (LLM decides when to stop)
- MAX_TOOL_ITERATIONS = 5 (prevent infinite loops)
- Tool execution timeout: 5 seconds
- Original design images context support

## Project Structure

```
agent-core/
├── app/
│   ├── main.py              # FastAPI entry point
│   ├── config.py            # Configuration management
│   ├── dependencies.py      # Dependency injection
│   └── routes/
│       ├── generate.py      # /api/v1/generate endpoint
│       └── chat.py          # /api/v1/chat endpoint
├── agents/
│   ├── base.py              # Base agent class
│   ├── layout_analyzer.py   # Pipeline agents
│   ├── component_detector.py
│   ├── interaction_infer.py
│   ├── code_generator.py
│   └── chat_agent.py        # ChatAgent with Tool Calling
├── tools/
│   ├── code_validator.py    # Core validation logic
│   ├── html_validator.py    # validate_html tool
│   └── interaction_checker.py # check_interaction tool
├── llm/
│   ├── gateway.py           # LLM unified gateway
│   └── prompts/             # Prompt templates
├── schemas/
│   ├── common.py            # Shared schemas (SSE events)
│   └── chat.py              # Chat-specific schemas
├── graph/
│   ├── state.py             # LangGraph state
│   └── generate_workflow.py # Pipeline workflow
├── tests/
│   ├── test_chat_agent.py   # ChatAgent tests (26 tests)
│   ├── test_code_validator.py
│   └── test_generate_api.py
├── requirements.txt
├── Dockerfile
└── README.md
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `AGENT_PORT` | 8081 | Server port |
| `LOG_LEVEL` | INFO | Logging level |
| `LLM_PROVIDER` | openai | LLM provider |
| `OPENAI_API_KEY` | - | OpenAI API key |
| `OPENAI_MODEL` | gpt-4o | OpenAI model |
| `OPENAI_BASE_URL` | - | Custom API endpoint |
| `MAX_RETRIES` | 3 | Pipeline retry count |

## Docker

```bash
# Build
docker build -t agent-core .

# Run
docker run -p 8081:8081 \
  -e OPENAI_API_KEY=sk-xxx \
  -e LLM_PROVIDER=openai \
  agent-core
```

## Testing

```bash
# Run all tests
pytest tests/ -v

# Run ChatAgent tests only
pytest tests/test_chat_agent.py -v

# Current test count: 52 tests
```

## Development Status

- [x] Phase 1: Project skeleton
- [x] Phase 2: Go ↔ Python communication
- [x] Phase 3: Single Agent validation
- [x] Phase 4: Full Pipeline (5 Agents)
- [x] Phase 5: ChatAgent + Tool Calling
- [x] Phase 6: Integration testing
