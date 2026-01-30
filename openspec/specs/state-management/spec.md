# state-management Specification

## Purpose
TBD - created by archiving change add-frontend-ui. Update Purpose after archive.
## Requirements
### Requirement: Global Project State

The system SHALL maintain a global state store for the current project session.

#### Scenario: State structure
- **WHEN** the application initializes
- **THEN** a Zustand store SHALL be created with:
  - `sessionId`: current session identifier
  - `images`: array of uploaded image files
  - `imageIds`: array of image IDs from backend
  - `generatedCode`: the generated code object (code + timestamp)
  - `thinkingContent`: AI analysis/thinking content
  - `status`: current operation status (idle, uploading, generating, completed, error)
  - `errorMessage`: error message if any
  - `chatMessages`: array of chat messages (user + assistant)
  - `activeFile`: currently active file in editor ('App.tsx' | 'styles.css')

### Requirement: Image State Management

The system SHALL provide actions to manage uploaded images.

#### Scenario: Add images
- **WHEN** user uploads images
- **THEN** the `addImages` action SHALL append images to the state

#### Scenario: Remove image
- **WHEN** user removes an image
- **THEN** the `removeImage` action SHALL remove the specified image from state

#### Scenario: Reorder images
- **WHEN** user reorders images
- **THEN** the `reorderImages` action SHALL update the image order in state

### Requirement: Code Generation State

The system SHALL track the code generation process status.

#### Scenario: Generation started
- **WHEN** code generation begins
- **THEN** status SHALL change to 'generating'
- **AND** previous error state SHALL be cleared

#### Scenario: Code streaming
- **WHEN** code chunks are received from SSE
- **THEN** `appendCode` action SHALL concatenate chunks to generatedCode

#### Scenario: Generation completed
- **WHEN** generation finishes successfully
- **THEN** status SHALL change to 'completed'

#### Scenario: Generation failed
- **WHEN** generation encounters an error
- **THEN** status SHALL change to 'error'
- **AND** error message SHALL be stored

### Requirement: Session Management

The system SHALL manage the session ID for API communication.

#### Scenario: Session created
- **WHEN** images are uploaded to the backend
- **THEN** the returned session_id SHALL be stored in state

#### Scenario: Reset session
- **WHEN** user starts a new project
- **THEN** the `reset` action SHALL clear all state to initial values

