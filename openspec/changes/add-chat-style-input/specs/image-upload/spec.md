# image-upload Spec Delta

## ADDED Requirements

### Requirement: Clipboard Paste Upload

The system SHALL support pasting images from clipboard using Ctrl+V or Cmd+V.

#### Scenario: Paste image from clipboard
- **GIVEN** user has an image in clipboard (e.g., from screenshot tool)
- **WHEN** user presses Ctrl+V (or Cmd+V on Mac) in the input area
- **THEN** the image SHALL be extracted from clipboard
- **AND** the image SHALL be added to the upload list
- **AND** a preview thumbnail SHALL be displayed

#### Scenario: Paste non-image content
- **GIVEN** user has text content in clipboard
- **WHEN** user pastes in the input area
- **THEN** text SHALL be inserted into the input field
- **AND** no image upload action SHALL occur

### Requirement: Horizontal Image List

The system SHALL display uploaded images in a horizontal scrollable list.

#### Scenario: Images displayed horizontally
- **WHEN** images are added to the upload list
- **THEN** images SHALL be displayed in a horizontal row
- **AND** the list SHALL scroll horizontally when images exceed available width

#### Scenario: Drag to reorder horizontally
- **WHEN** user drags an image card to a new horizontal position
- **THEN** the image order SHALL update accordingly
