# ui-layout Spec Delta

## MODIFIED Requirements

### Requirement: Experience Mode Layout

The experience mode layout SHALL use a two-column layout without a separate chat panel.

#### Scenario: Experience mode without right chat panel
- **GIVEN** user is in experience mode
- **WHEN** viewing the interface
- **THEN** the layout SHALL have two columns: interaction panel (left) and preview (right)
- **AND** no separate chat panel SHALL be displayed on the right

#### Scenario: Left panel width
- **GIVEN** user is in experience mode
- **WHEN** viewing the interface
- **THEN** the left interaction panel SHALL have a width of 300-320px
- **AND** the preview panel SHALL take remaining space

## REMOVED Requirements

### Requirement: Separate Chat Panel

~~The system SHALL display a separate chat panel on the right side for code modifications.~~

This requirement is removed. Chat functionality is now integrated into the left interaction panel.
