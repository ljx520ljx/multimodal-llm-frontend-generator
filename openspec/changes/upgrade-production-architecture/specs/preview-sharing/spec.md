## ADDED Requirements

### Requirement: Preview Sharing via Short Links
The system SHALL allow users to generate a shareable short link for any generated interactive prototype. The link SHALL serve a self-contained HTML page that renders the prototype without requiring the main application.

#### Scenario: Create share link
- **WHEN** a user clicks "Share" on a generated prototype
- **THEN** the system SHALL create an HTML snapshot, generate a nanoid 8-character short code, and return a shareable URL

#### Scenario: Access shared preview
- **WHEN** a visitor opens a share link (GET /p/:code)
- **THEN** the system SHALL return the HTML snapshot as a standalone page with Alpine.js and Tailwind CSS CDN included

#### Scenario: Share link expiration
- **WHEN** a share link has exceeded its configured TTL
- **THEN** the system SHALL return a 404 page indicating the link has expired

### Requirement: Share Link Management
The system SHALL provide APIs and UI for managing shared previews, including viewing active links, updating snapshots, and revoking access.

#### Scenario: Update shared preview
- **WHEN** a user modifies the prototype and clicks "Update Share"
- **THEN** the system SHALL update the HTML snapshot for the existing short code without changing the URL

#### Scenario: Revoke share link
- **WHEN** a user deletes a share link
- **THEN** the system SHALL mark the link as inactive and return 404 for subsequent access

### Requirement: Share Analytics
The system SHALL track view counts for each shared preview link.

#### Scenario: View count increment
- **WHEN** a visitor opens a share link
- **THEN** the system SHALL increment the view_count for that shared preview
