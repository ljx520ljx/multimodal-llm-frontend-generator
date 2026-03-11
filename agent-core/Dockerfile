FROM python:3.11-slim

WORKDIR /app

# Install system dependencies for lxml
RUN apt-get update && apt-get install -y --no-install-recommends \
    libxml2-dev \
    libxslt-dev \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY app/ ./app/
COPY agents/ ./agents/
COPY graph/ ./graph/
COPY llm/ ./llm/
COPY schemas/ ./schemas/
COPY tools/ ./tools/
COPY checkpoint/ ./checkpoint/
COPY registry/ ./registry/
COPY utils/ ./utils/

# Set environment variables
ENV PYTHONUNBUFFERED=1
ENV AGENT_PORT=8081
ENV LOG_LEVEL=INFO

EXPOSE 8081

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8081"]
