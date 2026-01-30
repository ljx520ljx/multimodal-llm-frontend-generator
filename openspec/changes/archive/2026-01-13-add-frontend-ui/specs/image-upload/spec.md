# Image Upload Capability

## ADDED Requirements

### Requirement: Drag and Drop Upload

The system SHALL provide a drag-and-drop zone for uploading UI design images.

#### Scenario: User drags images into upload zone
- **WHEN** user drags one or more image files into the upload zone
- **THEN** the zone SHALL display a visual highlight indicating drop target
- **AND** upon drop, images SHALL be added to the upload list

#### Scenario: User clicks to select files
- **WHEN** user clicks on the upload zone
- **THEN** a file picker dialog SHALL open
- **AND** selected images SHALL be added to the upload list

#### Scenario: Invalid file type rejected
- **WHEN** user attempts to upload a non-image file
- **THEN** the system SHALL reject the file
- **AND** display an error message indicating valid formats (PNG, JPG, WebP)

### Requirement: Image Preview List

The system SHALL display uploaded images in a preview list with thumbnails.

#### Scenario: Images displayed as cards
- **WHEN** images are added to the upload list
- **THEN** each image SHALL be displayed as a card with thumbnail preview
- **AND** the card SHALL show the image filename

#### Scenario: Remove image from list
- **WHEN** user clicks the delete button on an image card
- **THEN** the image SHALL be removed from the upload list

### Requirement: Image Reordering

The system SHALL allow users to reorder uploaded images via drag-and-drop.

#### Scenario: Drag to reorder
- **WHEN** user drags an image card to a new position
- **THEN** the image order SHALL update accordingly
- **AND** visual feedback SHALL indicate the new position during drag

#### Scenario: Order persisted
- **WHEN** images are reordered
- **THEN** the new order SHALL be reflected when generating code

### Requirement: Upload to Backend

The system SHALL upload images to the backend API and obtain a session ID.

#### Scenario: Successful upload
- **WHEN** user triggers code generation with images in the list
- **THEN** images SHALL be uploaded to `POST /api/upload`
- **AND** the returned session_id SHALL be stored for subsequent API calls

#### Scenario: Upload failure handling
- **WHEN** the upload API returns an error
- **THEN** the system SHALL display an error message
- **AND** allow the user to retry
