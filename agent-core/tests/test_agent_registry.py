"""Unit tests for AgentRegistry."""

import pytest
import tempfile
import os

from registry.agent_registry import AgentConfig, AgentRegistry, LLMConfig


@pytest.fixture
def configs_dir(tmp_path):
    """Create a temporary configs directory with sample YAML files."""
    # layout_analyzer.yaml
    (tmp_path / "layout_analyzer.yaml").write_text("""
name: layout_analyzer
version: "1.0"
description: Analyzes UI layout structure
tags: [pipeline, analysis]
llm:
  temperature: 0.2
input: [images]
output: [layout_info]
dependencies: []
""")

    # component_detector.yaml
    (tmp_path / "component_detector.yaml").write_text("""
name: component_detector
version: "1.0"
description: Detects UI components
tags: [pipeline, analysis]
llm:
  temperature: 0.2
input: [images, layout_info]
output: [component_info]
dependencies: [layout_analyzer]
""")

    # interaction_infer.yaml
    (tmp_path / "interaction_infer.yaml").write_text("""
name: interaction_infer
version: "1.0"
description: Infers interaction logic
tags: [pipeline, analysis]
llm:
  temperature: 0.5
input: [layout_info, component_info]
output: [interaction_info]
dependencies: [layout_analyzer, component_detector]
""")

    # code_generator.yaml
    (tmp_path / "code_generator.yaml").write_text("""
name: code_generator
version: "1.0"
description: Generates code
tags: [pipeline, generation]
llm:
  temperature: 0.3
  max_tokens: 8192
input: [layout_info, component_info, interaction_info]
output: [generated_code]
retry: 3
dependencies: [layout_analyzer, component_detector, interaction_infer]
""")

    # code_validator.yaml
    (tmp_path / "code_validator.yaml").write_text("""
name: code_validator
version: "1.0"
description: Validates generated code
tags: [pipeline, validation]
input: [generated_code]
output: [validation_result]
dependencies: [code_generator]
""")

    # chat_agent.yaml (non-pipeline)
    (tmp_path / "chat_agent.yaml").write_text("""
name: chat_agent
version: "1.0"
description: Chat-based code modification
tags: [chat]
llm:
  temperature: 0.7
input: [message, current_code]
output: [updated_code]
dependencies: []
""")

    return tmp_path


@pytest.fixture
def registry(configs_dir):
    """Create an AgentRegistry from the test configs."""
    return AgentRegistry(configs_dir)


class TestAgentConfig:
    """Tests for AgentConfig model."""

    def test_minimal_config(self):
        config = AgentConfig(name="test_agent")
        assert config.name == "test_agent"
        assert config.version == "1.0"
        assert config.tags == []
        assert config.dependencies == []
        assert config.retry == 0

    def test_full_config(self):
        config = AgentConfig(
            name="test",
            version="2.0",
            description="A test agent",
            tags=["pipeline"],
            llm=LLMConfig(temperature=0.5, max_tokens=4096),
            retry=3,
            dependencies=["dep1", "dep2"],
        )
        assert config.version == "2.0"
        assert config.llm.temperature == 0.5
        assert config.llm.max_tokens == 4096
        assert config.retry == 3
        assert config.dependencies == ["dep1", "dep2"]

    def test_alias_fields(self):
        """Test that 'input', 'output', 'prompt' aliases work."""
        data = {
            "name": "test",
            "input": ["images"],
            "output": ["layout_info"],
            "prompt": "Analyze the layout",
        }
        config = AgentConfig(**data)
        assert config.input_fields == ["images"]
        assert config.output_fields == ["layout_info"]
        assert config.prompt_template == "Analyze the layout"


class TestAgentRegistryLoading:
    """Tests for AgentRegistry loading behavior."""

    def test_load_all_configs(self, registry):
        assert len(registry) == 6

    def test_get_existing_agent(self, registry):
        config = registry.get("layout_analyzer")
        assert config is not None
        assert config.name == "layout_analyzer"
        assert config.llm.temperature == 0.2

    def test_get_nonexistent_agent(self, registry):
        config = registry.get("nonexistent")
        assert config is None

    def test_contains(self, registry):
        assert "layout_analyzer" in registry
        assert "nonexistent" not in registry

    def test_list_all(self, registry):
        all_configs = registry.list_all()
        assert len(all_configs) == 6
        names = {c.name for c in all_configs}
        assert "layout_analyzer" in names
        assert "chat_agent" in names

    def test_list_by_tag(self, registry):
        pipeline_agents = registry.list_by_tag("pipeline")
        assert len(pipeline_agents) == 5

        chat_agents = registry.list_by_tag("chat")
        assert len(chat_agents) == 1
        assert chat_agents[0].name == "chat_agent"

    def test_empty_directory(self, tmp_path):
        empty_dir = tmp_path / "empty"
        empty_dir.mkdir()
        reg = AgentRegistry(empty_dir)
        assert len(reg) == 0

    def test_nonexistent_directory(self, tmp_path):
        reg = AgentRegistry(tmp_path / "nonexistent")
        assert len(reg) == 0

    def test_invalid_yaml_skipped(self, tmp_path):
        (tmp_path / "valid.yaml").write_text("name: valid\n")
        (tmp_path / "invalid.yaml").write_text("{{invalid yaml content")
        reg = AgentRegistry(tmp_path)
        assert len(reg) == 1
        assert "valid" in reg


class TestPipelineOrder:
    """Tests for topological sort in get_pipeline_order."""

    def test_pipeline_order(self, registry):
        order = registry.get_pipeline_order()
        assert len(order) == 5  # 5 pipeline agents, chat_agent excluded

        # Verify dependencies are satisfied
        seen = set()
        for name in order:
            config = registry.get(name)
            for dep in config.dependencies:
                if dep in {c.name for c in registry.list_by_tag("pipeline")}:
                    assert dep in seen, f"{name} depends on {dep} but it hasn't run yet"
            seen.add(name)

    def test_pipeline_order_layout_first(self, registry):
        order = registry.get_pipeline_order()
        assert order[0] == "layout_analyzer"

    def test_pipeline_order_validator_last(self, registry):
        order = registry.get_pipeline_order()
        assert order[-1] == "code_validator"

    def test_circular_dependency_detection(self, tmp_path):
        (tmp_path / "a.yaml").write_text("""
name: agent_a
tags: [pipeline]
dependencies: [agent_b]
""")
        (tmp_path / "b.yaml").write_text("""
name: agent_b
tags: [pipeline]
dependencies: [agent_a]
""")
        reg = AgentRegistry(tmp_path)
        with pytest.raises(ValueError, match="Circular dependency"):
            reg.get_pipeline_order()

    def test_no_pipeline_agents(self, tmp_path):
        (tmp_path / "chat.yaml").write_text("""
name: chat
tags: [chat]
dependencies: []
""")
        reg = AgentRegistry(tmp_path)
        order = reg.get_pipeline_order()
        assert order == []

    def test_single_pipeline_agent(self, tmp_path):
        (tmp_path / "solo.yaml").write_text("""
name: solo_agent
tags: [pipeline]
dependencies: []
""")
        reg = AgentRegistry(tmp_path)
        order = reg.get_pipeline_order()
        assert order == ["solo_agent"]
