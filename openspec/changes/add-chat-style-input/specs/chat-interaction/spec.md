# chat-interaction Spec Delta

## MODIFIED Requirements

### Requirement: Unified Chat Interface

The chat interaction SHALL be integrated into the left panel instead of a separate right panel.

#### Scenario: Chat history displayed in left panel
- **GIVEN** user has generated code
- **WHEN** viewing the interface
- **THEN** chat history SHALL be displayed in the left panel below the image list
- **AND** no separate right chat panel SHALL exist

#### Scenario: Send modification request
- **GIVEN** code has been generated
- **WHEN** user types a modification request and clicks send
- **THEN** the request SHALL be added to chat history as user message
- **AND** the AI response SHALL be streamed and displayed as assistant message

## ADDED Requirements

### Requirement: Unified Input Component

The system SHALL provide a unified input component for both initial generation and chat modifications.

#### Scenario: Initial generation with prompt
- **GIVEN** images are uploaded but no code generated
- **WHEN** user types in the input and clicks send
- **THEN** the system SHALL generate code using the images and prompt text
- **AND** a user message with images SHALL be added to chat history

#### Scenario: Chat modification after generation
- **GIVEN** code has been generated
- **WHEN** user types a modification request and clicks send
- **THEN** the system SHALL call the chat API
- **AND** messages SHALL be added to chat history

#### Scenario: Enter to send
- **WHEN** user presses Enter in the input field
- **THEN** the message SHALL be sent
- **AND** Shift+Enter SHALL create a new line instead
