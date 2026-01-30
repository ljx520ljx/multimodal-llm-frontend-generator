# UI Layout Capability

## ADDED Requirements

### Requirement: Dual-Mode Layout

The system SHALL provide two switchable layout modes to serve different user needs.

#### Scenario: Experience mode (default)
- **WHEN** user accesses the main page
- **THEN** the layout SHALL display in experience mode by default
- **AND** preview area SHALL occupy ~80% of the width
- **AND** upload panel SHALL be displayed as a sidebar (~20%)
- **AND** code editor SHALL be collapsed/hidden

#### Scenario: Developer mode
- **WHEN** user switches to developer mode
- **THEN** the layout SHALL display three panels:
  - Left panel: Image upload area (~20% width)
  - Center panel: Code editor (~45% width)
  - Right panel: Live preview (~35% width)

#### Scenario: Mode switching
- **WHEN** user clicks the mode toggle button
- **THEN** the layout SHALL smoothly transition between modes
- **AND** the selected mode SHALL be persisted (localStorage)

### Requirement: Collapsible Code Panel

The system SHALL provide a collapsible code panel in experience mode.

#### Scenario: Code panel collapsed (default in experience mode)
- **WHEN** user is in experience mode
- **THEN** code panel SHALL be collapsed by default
- **AND** a "Show Code" button SHALL be visible

#### Scenario: Code panel expanded
- **WHEN** user clicks "Show Code" button
- **THEN** code panel SHALL expand from the bottom or side
- **AND** user can view and edit the generated code

### Requirement: Header Component

The system SHALL display a header with the application title and branding.

#### Scenario: Header displays
- **WHEN** user accesses any page
- **THEN** a header SHALL be visible at the top
- **AND** contain the application name

### Requirement: Responsive Design

The system SHALL adapt the layout for different screen sizes.

#### Scenario: Desktop view
- **WHEN** viewport width is >= 1024px
- **THEN** all three panels SHALL be visible side by side

#### Scenario: Tablet view
- **WHEN** viewport width is between 768px and 1023px
- **THEN** the layout MAY stack panels vertically or hide some panels

#### Scenario: Mobile view
- **WHEN** viewport width is < 768px
- **THEN** the layout SHALL display panels in a stacked vertical layout
- **AND** user can navigate between panels

### Requirement: Basic UI Components

The system SHALL provide reusable UI components following consistent design patterns.

#### Scenario: Button component
- **WHEN** a button is needed
- **THEN** the Button component SHALL support variants (primary, secondary, ghost)
- **AND** support sizes (sm, md, lg)
- **AND** support loading state

#### Scenario: Card component
- **WHEN** content needs to be contained
- **THEN** the Card component SHALL provide consistent styling
- **AND** support optional header and footer sections

#### Scenario: Skeleton component
- **WHEN** content is loading
- **THEN** a skeleton placeholder SHALL be displayed
- **AND** animate to indicate loading state
