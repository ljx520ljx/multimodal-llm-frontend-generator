# Agent Core Specification

## ADDED Requirements

### Requirement: Health Check Endpoint

The agent-core service SHALL provide a health check endpoint at `/health` that returns the service status.

#### Scenario: Health check returns OK

- **WHEN** a GET request is made to `/health`
- **THEN** the response status code SHALL be 200
- **AND** the response body SHALL be `{"status": "ok"}`

### Requirement: Configuration Management

The agent-core service SHALL support configuration through environment variables.

#### Scenario: Read port from environment

- **WHEN** the `AGENT_PORT` environment variable is set to `8081`
- **THEN** the service SHALL listen on port 8081

#### Scenario: Default port when not configured

- **WHEN** the `AGENT_PORT` environment variable is not set
- **THEN** the service SHALL listen on port 8081 by default

### Requirement: Structured Logging

The agent-core service SHALL output structured logs with configurable log level.

#### Scenario: Log level configuration

- **WHEN** the `LOG_LEVEL` environment variable is set to `DEBUG`
- **THEN** the service SHALL output debug-level logs

#### Scenario: Default log level

- **WHEN** the `LOG_LEVEL` environment variable is not set
- **THEN** the service SHALL default to `INFO` level

### Requirement: CORS Support

The agent-core service SHALL support Cross-Origin Resource Sharing (CORS) for frontend integration.

#### Scenario: CORS preflight request

- **WHEN** an OPTIONS request is made with `Origin` header
- **THEN** the response SHALL include appropriate CORS headers
- **AND** the response status code SHALL be 200

### Requirement: Docker Deployment

The agent-core service SHALL be deployable as a Docker container.

#### Scenario: Build Docker image

- **WHEN** running `docker build -t agent-core .` in the agent-core directory
- **THEN** the build SHALL succeed without errors

#### Scenario: Run Docker container

- **WHEN** running `docker run -p 8081:8081 agent-core`
- **THEN** the service SHALL be accessible at `http://localhost:8081/health`
