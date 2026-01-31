# Agent Core Specification - Generate Pipeline

## ADDED Requirements

### Requirement: Generate Endpoint

The agent-core service SHALL provide a generate endpoint at `/api/v1/generate` that runs the multi-agent pipeline and returns SSE events.

#### Scenario: Successful code generation

- **WHEN** a POST request is made to `/api/v1/generate` with valid images
- **THEN** the response content-type SHALL be `text/event-stream`
- **AND** the pipeline SHALL execute: LayoutAnalyzer → ComponentDetector → InteractionInfer → CodeGenerator → CodeValidator
- **AND** the response SHALL include `agent_start` events for each agent
- **AND** the response SHALL include `agent_result` events with structured output
- **AND** the final event SHALL be `event: done\ndata: {"success": true}`

#### Scenario: Validation failure triggers retry

- **WHEN** CodeValidator detects errors in the generated code
- **THEN** the pipeline SHALL retry CodeGenerator with validation feedback
- **AND** the maximum retry count SHALL be configurable (default 3)
- **AND** if max retries exceeded, the response SHALL include error details

### Requirement: LayoutAnalyzer Agent

The LayoutAnalyzer agent SHALL analyze design images and output a LayoutSchema.

#### Scenario: Layout analysis

- **WHEN** LayoutAnalyzer receives design images
- **THEN** it SHALL identify the overall layout structure (e.g., "sidebar-main", "header-body-footer")
- **AND** it SHALL identify regions with names, positions, and estimated sizes
- **AND** it SHALL determine the grid system (flex or grid)

### Requirement: ComponentDetector Agent

The ComponentDetector agent SHALL identify UI components in design images.

#### Scenario: Component detection

- **WHEN** ComponentDetector receives images and LayoutSchema
- **THEN** it SHALL identify components with id, type, name, and region
- **AND** it SHALL extract component properties (text, variant, etc.)
- **AND** component types SHALL include: button, input, card, nav, list, etc.

### Requirement: InteractionInfer Agent

The InteractionInfer agent SHALL infer state machine transitions between design images.

#### Scenario: State machine inference

- **WHEN** InteractionInfer receives images, LayoutSchema, and ComponentList
- **THEN** it SHALL create a State for each design image
- **AND** it SHALL infer Transitions between states (from_state, to_state, trigger)
- **AND** transitions SHALL support free navigation (1→2→1→3→1→2...)
- **AND** it SHALL identify the initial_state

### Requirement: CodeGenerator Agent

The CodeGenerator agent SHALL generate HTML + Tailwind + Alpine.js code.

#### Scenario: Code generation

- **WHEN** CodeGenerator receives analysis results (Layout, Components, Interactions)
- **THEN** it SHALL generate a single HTML file with embedded Alpine.js
- **AND** the code SHALL implement all states as Alpine.js state machine
- **AND** the code SHALL implement all transitions as @click handlers
- **AND** the code SHALL use Tailwind CSS for styling

#### Scenario: Code generation with validation feedback

- **WHEN** CodeGenerator receives validation errors from a previous attempt
- **THEN** it SHALL incorporate the feedback to fix the issues
- **AND** it SHALL regenerate the complete code

### Requirement: CodeValidator Tool

The CodeValidator tool SHALL validate generated HTML + Alpine.js code using rule-based checks.

#### Scenario: HTML syntax validation

- **WHEN** CodeValidator receives generated code
- **THEN** it SHALL check for proper HTML tag closure
- **AND** it SHALL check for valid Alpine.js directives (x-data, @click, x-show)
- **AND** it SHALL return a list of errors with fix suggestions

#### Scenario: State machine completeness check

- **WHEN** CodeValidator receives generated code and InteractionSpec
- **THEN** it SHALL verify all states are defined in x-data
- **AND** it SHALL verify all transitions have corresponding @click handlers
- **AND** missing transitions SHALL be reported as errors

### Requirement: LLM Gateway

The agent-core service SHALL provide a unified LLM gateway supporting multiple providers.

#### Scenario: OpenAI provider

- **WHEN** LLM_PROVIDER is set to "openai"
- **THEN** the gateway SHALL use OpenAI API with the configured model
- **AND** it SHALL support both sync and streaming responses
- **AND** it SHALL support custom base_url for OpenAI-compatible APIs

#### Scenario: Anthropic provider

- **WHEN** LLM_PROVIDER is set to "anthropic"
- **THEN** the gateway SHALL use Anthropic API with the configured model
- **AND** it SHALL support both sync and streaming responses

#### Scenario: Google provider

- **WHEN** LLM_PROVIDER is set to "google"
- **THEN** the gateway SHALL use Google Generative AI API with the configured model
- **AND** it SHALL support both sync and streaming responses

#### Scenario: OpenAI-compatible providers (DeepSeek, Doubao, GLM, Kimi)

- **WHEN** LLM_PROVIDER is set to "deepseek", "doubao", "glm", or "kimi"
- **THEN** the gateway SHALL use OpenAI-compatible API with provider-specific base_url
- **AND** it SHALL support both sync and streaming responses
- **AND** provider base URLs SHALL be:
  - deepseek: `https://api.deepseek.com/v1`
  - doubao: `https://ark.cn-beijing.volces.com/api/v3`
  - glm: `https://open.bigmodel.cn/api/paas/v4`
  - kimi: `https://api.moonshot.cn/v1`

### Requirement: Structured Output

All agents SHALL use structured output to ensure response schema compliance.

#### Scenario: Schema enforcement

- **WHEN** an agent calls the LLM
- **THEN** the response SHALL conform to the agent's output schema (Pydantic model)
- **AND** parsing failures SHALL trigger automatic retry
