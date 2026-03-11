## ADDED Requirements

### Requirement: Pluggable Agent Configuration
The system SHALL support declarative Agent configuration via YAML files. Each Agent MUST be defined by a YAML file specifying its name, LLM parameters, input/output schema, prompt templates, retry policy, and dependencies.

#### Scenario: New agent registration
- **WHEN** a new YAML file is added to the `agents/configs/` directory
- **THEN** the AgentRegistry SHALL load it on startup and make it available for pipeline execution

#### Scenario: Agent execution from config
- **WHEN** the WorkflowOrchestrator runs a pipeline
- **THEN** it SHALL create Agent instances from YAML configuration via AgentRegistry

### Requirement: DesignState Unified State Model
The system SHALL use a DesignState model as the shared knowledge base across all pipeline agents. Each Agent SHALL read from and append to the DesignState during execution.

#### Scenario: State accumulation across agents
- **WHEN** LayoutAnalyzer completes
- **THEN** the DesignState SHALL contain the layout analysis result and the agent SHALL be recorded in `completed_agents`

#### Scenario: State checkpoint
- **WHEN** any pipeline agent completes successfully
- **THEN** the DesignState SHALL be checkpointed with a timestamp

### Requirement: Pipeline Checkpoint and Resume
The system SHALL support saving pipeline state after each agent completion and resuming from the last checkpoint on failure.

#### Scenario: Resume after failure
- **WHEN** a pipeline fails at the CodeGenerator stage
- **THEN** the system SHALL allow resuming from the last completed agent (InteractionInfer) without re-running previous agents

#### Scenario: Checkpoint expiration
- **WHEN** a checkpoint is older than the configured TTL (default 1 hour)
- **THEN** the system SHALL discard it and require a fresh pipeline run

### Requirement: Dependency Graph Orchestration
The system SHALL build an execution dependency graph from Agent YAML `dependencies` fields and execute agents in topological order. Agents at the same dependency level MAY be executed in parallel.

#### Scenario: Parallel agent execution
- **WHEN** two agents have no mutual dependencies and share the same dependency level
- **THEN** the orchestrator SHALL execute them concurrently

## MODIFIED Requirements

### Requirement: Pipeline Agent Execution
The system SHALL execute pipeline agents through the WorkflowOrchestrator using YAML-configured agents and DesignState, replacing the previous hardcoded sequential execution in GenerateWorkflow. The pipeline SHALL support checkpoint/resume and dependency-graph-based ordering.

#### Scenario: Full pipeline success
- **WHEN** all pipeline agents complete successfully
- **THEN** the system SHALL emit SSE events for each agent's progress and return the final generated code

#### Scenario: Pipeline resume on failure
- **WHEN** a pipeline run fails and a valid checkpoint exists
- **THEN** the system SHALL resume from the last checkpoint instead of restarting
