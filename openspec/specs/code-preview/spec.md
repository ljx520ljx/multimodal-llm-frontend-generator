# code-preview Specification

## Purpose
TBD - created by archiving change add-frontend-ui. Update Purpose after archive.
## Requirements
### Requirement: Sandpack Preview Integration

The system SHALL provide a live preview of generated code using Sandpack sandbox.

#### Scenario: Code renders in preview
- **WHEN** valid React code is present in the editor
- **THEN** the Sandpack preview SHALL render the code
- **AND** display the visual output in real-time

#### Scenario: Preview updates on code change
- **WHEN** code in the editor changes (streaming or manual edit)
- **THEN** the preview SHALL automatically re-render
- **AND** reflect the latest code state

### Requirement: React and Tailwind Support

The system SHALL configure Sandpack to support React 18 and Tailwind CSS.

#### Scenario: Tailwind classes work
- **WHEN** generated code uses Tailwind CSS classes
- **THEN** the preview SHALL correctly render the styles
- **AND** all Tailwind utilities SHALL be available

#### Scenario: React 18 features supported
- **WHEN** generated code uses React 18 features
- **THEN** the preview SHALL correctly execute the code

### Requirement: Compilation Error Display

The system SHALL display user-friendly error messages when preview compilation fails.

#### Scenario: Syntax error in code
- **WHEN** the code contains a syntax error
- **THEN** the preview SHALL display an error message
- **AND** indicate the error location if possible

#### Scenario: Runtime error in code
- **WHEN** the code throws a runtime error
- **THEN** the preview SHALL display the error message
- **AND** NOT crash the entire application

### Requirement: Preview Loading State

The system SHALL display a loading indicator while the preview is compiling.

#### Scenario: Initial compilation
- **WHEN** code is first loaded into the preview
- **THEN** a loading indicator SHALL be displayed
- **AND** replaced by the rendered output once ready

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

