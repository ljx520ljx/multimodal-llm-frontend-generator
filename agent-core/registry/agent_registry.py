"""AgentRegistry - Loads and manages Agent configurations from YAML files."""

from __future__ import annotations

import logging
from collections import deque
from pathlib import Path
from typing import Optional, Union

import yaml
from pydantic import BaseModel, Field

logger = logging.getLogger(__name__)


class LLMConfig(BaseModel):
    """LLM configuration for an agent."""

    temperature: float = 0.7
    max_tokens: Optional[int] = None


class AgentConfig(BaseModel):
    """Configuration for a single agent loaded from YAML."""

    name: str
    version: str = "1.0"
    description: str = ""
    tags: list[str] = Field(default_factory=list)
    llm: LLMConfig = Field(default_factory=LLMConfig)
    input_fields: list[str] = Field(default_factory=list, alias="input")
    output_fields: list[str] = Field(default_factory=list, alias="output")
    prompt_template: str = Field(default="", alias="prompt")
    retry: int = 0
    dependencies: list[str] = Field(default_factory=list)

    model_config = {"populate_by_name": True}


class AgentRegistry:
    """Registry that loads Agent configurations from YAML files.

    Usage:
        registry = AgentRegistry("agents/configs")
        config = registry.get("layout_analyzer")
        pipeline = registry.get_pipeline_order()
    """

    def __init__(self, configs_dir: Union[str, Path]):
        self._configs: dict[str, AgentConfig] = {}
        self._configs_dir = Path(configs_dir)
        self._load_all()

    def _load_all(self) -> None:
        """Scan configs directory and load all .yaml files."""
        if not self._configs_dir.exists():
            logger.warning(f"Agent configs directory not found: {self._configs_dir}")
            return

        for yaml_file in sorted(self._configs_dir.glob("*.yaml")):
            try:
                with open(yaml_file, encoding="utf-8") as f:
                    data = yaml.safe_load(f)
                if data is None:
                    continue
                config = AgentConfig(**data)
                self._configs[config.name] = config
                logger.info(f"Loaded agent config: {config.name} v{config.version}")
            except Exception as e:
                logger.error(f"Failed to load agent config {yaml_file}: {e}")

    def get(self, name: str) -> Optional[AgentConfig]:
        """Get agent config by name."""
        return self._configs.get(name)

    def list_all(self) -> list[AgentConfig]:
        """List all loaded agent configs."""
        return list(self._configs.values())

    def list_by_tag(self, tag: str) -> list[AgentConfig]:
        """List agents that have the specified tag."""
        return [c for c in self._configs.values() if tag in c.tags]

    def get_pipeline_order(self) -> list[str]:
        """Compute topological order of pipeline agents based on dependencies.

        Uses Kahn's algorithm. Raises ValueError on circular dependencies.

        Returns:
            Ordered list of agent names for pipeline execution.
        """
        pipeline_agents = {
            name: config
            for name, config in self._configs.items()
            if "pipeline" in config.tags
        }

        if not pipeline_agents:
            return []

        # Build adjacency and in-degree
        in_degree: dict[str, int] = {name: 0 for name in pipeline_agents}
        dependents: dict[str, list[str]] = {name: [] for name in pipeline_agents}

        for name, config in pipeline_agents.items():
            for dep in config.dependencies:
                if dep in pipeline_agents:
                    dependents[dep].append(name)
                    in_degree[name] += 1

        # Kahn's algorithm
        queue = deque(name for name, deg in in_degree.items() if deg == 0)
        result: list[str] = []

        while queue:
            node = queue.popleft()
            result.append(node)
            for dependent in dependents[node]:
                in_degree[dependent] -= 1
                if in_degree[dependent] == 0:
                    queue.append(dependent)

        if len(result) != len(pipeline_agents):
            remaining = set(pipeline_agents) - set(result)
            raise ValueError(f"Circular dependency detected among agents: {remaining}")

        return result

    def __len__(self) -> int:
        return len(self._configs)

    def __contains__(self, name: str) -> bool:
        return name in self._configs
