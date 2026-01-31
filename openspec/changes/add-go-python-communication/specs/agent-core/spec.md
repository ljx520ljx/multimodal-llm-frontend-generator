# Agent Core Specification - SSE Communication

## ADDED Requirements

### Requirement: Echo Endpoint for Communication Testing

The agent-core service SHALL provide an echo endpoint at `/api/v1/echo` that returns a Server-Sent Events stream for testing Go ↔ Python communication.

#### Scenario: Echo stream returns specified count of messages

- **WHEN** a POST request is made to `/api/v1/echo` with body `{"message": "test", "count": 3}`
- **THEN** the response content-type SHALL be `text/event-stream`
- **AND** the response SHALL contain 3 SSE events with the message
- **AND** each event SHALL have format `event: message\ndata: {"index": N, "message": "test"}\n\n`
- **AND** the final event SHALL be `event: done\ndata: {}\n\n`

#### Scenario: Echo stream with default count

- **WHEN** a POST request is made to `/api/v1/echo` with body `{"message": "hello"}`
- **THEN** the response SHALL default to 5 SSE events

### Requirement: SSE Stream Format

The agent-core service SHALL use a consistent SSE format for all streaming responses.

#### Scenario: SSE event format

- **WHEN** the service sends an SSE event
- **THEN** each event SHALL follow the format:
  ```
  event: <event_type>
  data: <json_payload>

  ```
- **AND** event types SHALL be one of: `message`, `agent_start`, `agent_result`, `thinking`, `code`, `tool_call`, `tool_result`, `error`, `done`

#### Scenario: SSE stream completion

- **WHEN** the service completes a streaming response
- **THEN** the final event SHALL be `event: done\ndata: {}\n\n`

### Requirement: Request Timeout Handling

The agent-core service SHALL handle client disconnection gracefully.

#### Scenario: Client disconnects during stream

- **WHEN** the client closes the connection during an active stream
- **THEN** the service SHALL stop generating events
- **AND** the service SHALL release associated resources
