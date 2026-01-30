# code-editor Delta

## ADDED Requirements

### Requirement: Editor Toolbar

The system SHALL provide a toolbar with editor utility actions.

#### Scenario: Copy code to clipboard
- **GIVEN** code is displayed in the editor
- **WHEN** the user clicks the copy button
- **THEN** the current code SHALL be copied to the clipboard
- **AND** a success indicator SHALL be shown briefly

#### Scenario: Switch between files
- **GIVEN** multiple files exist (App.tsx, styles.css)
- **WHEN** the user clicks a file tab
- **THEN** the editor SHALL display the selected file's content
- **AND** maintain the correct syntax highlighting

### Requirement: Enhanced JSX Support

The system SHALL provide proper JSX/TSX syntax support in the editor.

#### Scenario: JSX tags highlighted
- **WHEN** JSX code is displayed
- **THEN** JSX tags SHALL be highlighted differently from regular elements
- **AND** JSX attributes SHALL be properly colored

#### Scenario: TypeScript types recognized
- **WHEN** TypeScript type annotations are present
- **THEN** type keywords SHALL be highlighted appropriately
- **AND** type errors SHALL be underlined if detected

### Requirement: Multi-file Support

The system SHALL support editing multiple files for the generated component.

#### Scenario: Edit App.tsx
- **GIVEN** the default file is App.tsx
- **WHEN** the user modifies the component code
- **THEN** changes SHALL be reflected in the preview

#### Scenario: Edit styles.css
- **GIVEN** a styles.css file exists
- **WHEN** the user switches to styles.css tab
- **THEN** CSS content SHALL be displayed
- **AND** CSS syntax highlighting SHALL be active
