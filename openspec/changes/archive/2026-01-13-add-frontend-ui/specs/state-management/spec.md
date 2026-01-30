# State Management Capability

## ADDED Requirements

### Requirement: Global Project State

The system SHALL maintain a global state store for the current project session.

#### Scenario: State structure
- **WHEN** the application initializes
- **THEN** a Zustand store SHALL be created with:
  - `viewMode`: current view mode ('experience' | 'developer')
  - `codeExpanded`: whether code panel is expanded in experience mode
  - `images`: array of uploaded image files
  - `sessionId`: current session identifier
  - `generatedCode`: the generated code string
  - `status`: current operation status (idle, uploading, generating, completed, error)
  - `error`: error message if any

### Requirement: View Mode State

The system SHALL manage the view mode state for layout switching.

#### Scenario: Default view mode
- **WHEN** user first visits the application
- **THEN** viewMode SHALL default to 'experience'

#### Scenario: View mode persistence
- **WHEN** user changes the view mode
- **THEN** the selection SHALL be persisted to localStorage
- **AND** restored on next visit

#### Scenario: Code panel toggle
- **WHEN** user toggles the code panel in experience mode
- **THEN** codeExpanded state SHALL update accordingly

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
