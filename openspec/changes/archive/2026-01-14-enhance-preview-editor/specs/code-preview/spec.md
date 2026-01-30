# code-preview Delta

## ADDED Requirements

### Requirement: Real-time Code Synchronization

The system SHALL automatically update the preview when code changes in the editor.

#### Scenario: Manual code edit triggers preview update
- **GIVEN** the user is in developer mode
- **WHEN** the user edits code in the Monaco Editor
- **THEN** the preview SHALL update after a 300ms debounce delay
- **AND** only the changed file SHALL be updated in Sandpack

#### Scenario: Streaming code updates preview
- **WHEN** code is being streamed from the backend
- **THEN** the preview SHALL update incrementally
- **AND** maintain stable rendering during streaming

### Requirement: Preview Toolbar

The system SHALL provide a toolbar with preview control actions.

#### Scenario: Refresh preview manually
- **GIVEN** the preview is displaying generated code
- **WHEN** the user clicks the refresh button
- **THEN** the Sandpack preview SHALL recompile and re-render

#### Scenario: Enter fullscreen preview
- **WHEN** the user clicks the fullscreen button
- **THEN** the preview SHALL display in a modal overlay
- **AND** cover the entire viewport
- **AND** support ESC key to exit

### Requirement: Enhanced Error Display

The system SHALL provide detailed and actionable error messages.

#### Scenario: Show error location
- **WHEN** a compilation error occurs
- **THEN** the error message SHALL include the file name
- **AND** include the line number if available
- **AND** provide a clear error description

#### Scenario: Recover from error
- **GIVEN** the preview is showing an error
- **WHEN** the user fixes the code error
- **THEN** the preview SHALL automatically recover
- **AND** display the corrected output
