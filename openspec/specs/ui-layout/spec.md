# ui-layout Specification

## Purpose
定义前端应用的布局结构，提供简洁的双栏布局以支持交互原型验证。

## Requirements
### Requirement: Two-Panel Layout

The system SHALL provide a fixed two-panel layout for all users.

#### Scenario: Main layout
- **WHEN** user accesses the main page
- **THEN** the layout SHALL display two panels:
  - Left panel: Interaction panel (fixed 320px width)
  - Right panel: Preview panel (flex-1, fills remaining space)

#### Scenario: Interaction panel content
- **WHEN** user views the interaction panel
- **THEN** it SHALL contain:
  - Image upload area
  - Chat input field
  - Chat history display

#### Scenario: Preview panel content
- **WHEN** code has been generated
- **THEN** the preview panel SHALL display:
  - HTML + Alpine.js preview in an iframe sandbox
  - Refresh button
  - Fullscreen toggle (optional)

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

