# Agent Core

Python Agent Service for Multi-Agent Architecture.

## Overview

This service handles AI Agent logic including:
- Agent orchestration (LangGraph)
- Prompt building
- LLM calls

## Quick Start

### Prerequisites

- Python 3.11+
- pip or poetry

### Installation

```bash
cd agent-core
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install -r requirements.txt
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

## Configuration

Environment variables (can be set in `.env` file):

| Variable | Default | Description |
|----------|---------|-------------|
| `AGENT_PORT` | 8081 | Server port |
| `AGENT_HOST` | 0.0.0.0 | Server host |
| `LOG_LEVEL` | INFO | Logging level (DEBUG, INFO, WARNING, ERROR) |
| `CORS_ORIGINS` | ["http://localhost:3000", "http://localhost:8080"] | Allowed CORS origins |
| `OPENAI_API_KEY` | - | OpenAI API key (Phase 2+) |
| `ANTHROPIC_API_KEY` | - | Anthropic API key (Phase 2+) |

## Docker

```bash
# Build
docker build -t agent-core .

# Run
docker run -p 8081:8081 agent-core
```

## Project Structure

```
agent-core/
├── app/
│   ├── __init__.py
│   ├── main.py          # FastAPI entry point
│   ├── config.py        # Configuration management
│   └── routes/          # API routes (Phase 2+)
├── pyproject.toml
├── requirements.txt
├── Dockerfile
└── README.md
```

## Development Roadmap

- **Phase 1** (Current): Project skeleton, health check, configuration
- **Phase 2**: Go ↔ Python communication, SSE streaming
- **Phase 3**: Single Agent validation (CodeGenerator)
- **Phase 4**: Full Pipeline (5 Agents + LangGraph)
- **Phase 5**: ChatAgent + Tools
- **Phase 6**: Integration testing
