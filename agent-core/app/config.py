"""Configuration management using Pydantic Settings."""

from __future__ import annotations

from functools import lru_cache
from typing import Literal, Optional

from pydantic_settings import BaseSettings, SettingsConfigDict


# Agent types that can have custom model configurations
AgentType = Literal["default", "chat", "layout", "component", "interaction", "codegen", "validator"]


class Settings(BaseSettings):
    """Application settings loaded from environment variables."""

    # Server settings
    agent_port: int = 8081
    agent_host: str = "0.0.0.0"

    # Logging
    log_level: Literal["DEBUG", "INFO", "WARNING", "ERROR"] = "INFO"

    # CORS
    cors_origins: list[str] = ["http://localhost:3000", "http://localhost:8080"]

    # LLM Provider selection (default for all agents)
    llm_provider: Literal[
        "openai", "anthropic", "google", "deepseek", "doubao", "glm", "kimi"
    ] = "openai"

    # Pipeline settings
    max_retries: int = 3

    # LLM request timeout (seconds) — innermost timeout in the chain:
    # LLM request 60s < Go AGENT_TIMEOUT 180s < Go HANDLER_TIMEOUT 240s < Frontend SSE 300s
    llm_request_timeout: int = 60

    # Internal API token for Go backend checkpoint API (must match Go INTERNAL_API_TOKEN)
    internal_api_token: str = ""

    # Go backend URL for checkpoint API
    backend_url: str = "http://localhost:8080"

    # OpenAI
    openai_api_key: str = ""
    openai_model: str = "gpt-4o"
    openai_base_url: Optional[str] = None

    # Anthropic
    anthropic_api_key: str = ""
    anthropic_model: str = "claude-3-5-sonnet-20241022"

    # Google
    google_api_key: str = ""
    google_model: str = "gemini-1.5-pro"

    # DeepSeek
    deepseek_api_key: str = ""
    deepseek_model: str = "deepseek-chat"

    # Doubao (Volcengine)
    doubao_api_key: str = ""
    doubao_model: str = "doubao-pro-32k"

    # GLM (Zhipu)
    glm_api_key: str = ""
    glm_model: str = "glm-4v"

    # Kimi (Moonshot)
    kimi_api_key: str = ""
    kimi_model: str = "moonshot-v1-32k"

    # ===========================================
    # Agent-specific model overrides (optional)
    # If not set, uses default llm_provider + model
    # ===========================================

    # ChatAgent model override
    chat_agent_model: Optional[str] = None

    # Pipeline Agent model overrides
    layout_agent_model: Optional[str] = None      # LayoutAnalyzer
    component_agent_model: Optional[str] = None   # ComponentDetector
    interaction_agent_model: Optional[str] = None # InteractionInfer
    codegen_agent_model: Optional[str] = None     # CodeGenerator
    validator_agent_model: Optional[str] = None   # CodeValidator (轻量任务可用便宜模型)

    # ===========================================
    # Agent-specific temperature overrides (optional)
    # If not set, uses default 0.7
    # ===========================================
    default_temperature: float = 0.7
    chat_agent_temperature: Optional[float] = None
    layout_agent_temperature: Optional[float] = None
    component_agent_temperature: Optional[float] = None
    interaction_agent_temperature: Optional[float] = None
    codegen_agent_temperature: Optional[float] = None     # 代码生成建议低温 (0.2-0.3)
    validator_agent_temperature: Optional[float] = None

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
    )

    def get_llm_config(self, agent_type: AgentType = "default") -> dict:
        """Get LLM configuration for the selected provider.

        Args:
            agent_type: The agent type to get config for. If the agent has
                       a custom model override configured, that model will be used.
                       Otherwise falls back to the default provider model.

        Returns:
            dict with keys: provider, api_key, model, base_url
        """
        provider = self.llm_provider
        configs = {
            "openai": {
                "api_key": self.openai_api_key,
                "model": self.openai_model,
                "base_url": self.openai_base_url,
            },
            "anthropic": {
                "api_key": self.anthropic_api_key,
                "model": self.anthropic_model,
                "base_url": None,
            },
            "google": {
                "api_key": self.google_api_key,
                "model": self.google_model,
                "base_url": None,
            },
            "deepseek": {
                "api_key": self.deepseek_api_key,
                "model": self.deepseek_model,
                "base_url": None,  # Will use default from gateway
            },
            "doubao": {
                "api_key": self.doubao_api_key,
                "model": self.doubao_model,
                "base_url": None,
            },
            "glm": {
                "api_key": self.glm_api_key,
                "model": self.glm_model,
                "base_url": None,
            },
            "kimi": {
                "api_key": self.kimi_api_key,
                "model": self.kimi_model,
                "base_url": None,
            },
        }

        config = {"provider": provider, **configs[provider]}

        # Apply agent-specific model override if configured
        agent_model_overrides = {
            "chat": self.chat_agent_model,
            "layout": self.layout_agent_model,
            "component": self.component_agent_model,
            "interaction": self.interaction_agent_model,
            "codegen": self.codegen_agent_model,
            "validator": self.validator_agent_model,
        }

        if agent_type != "default" and agent_type in agent_model_overrides:
            override_model = agent_model_overrides[agent_type]
            if override_model:
                config["model"] = override_model

        # Apply agent-specific temperature
        agent_temperature_overrides = {
            "chat": self.chat_agent_temperature,
            "layout": self.layout_agent_temperature,
            "component": self.component_agent_temperature,
            "interaction": self.interaction_agent_temperature,
            "codegen": self.codegen_agent_temperature,
            "validator": self.validator_agent_temperature,
        }

        temperature = self.default_temperature
        if agent_type != "default" and agent_type in agent_temperature_overrides:
            override_temp = agent_temperature_overrides[agent_type]
            if override_temp is not None:
                temperature = override_temp
        config["temperature"] = temperature

        return config


@lru_cache
def get_settings() -> Settings:
    """Get cached settings instance."""
    return Settings()
