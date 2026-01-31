"""Configuration management using Pydantic Settings."""

from functools import lru_cache
from typing import Literal

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Application settings loaded from environment variables."""

    # Server settings
    agent_port: int = 8081
    agent_host: str = "0.0.0.0"

    # Logging
    log_level: Literal["DEBUG", "INFO", "WARNING", "ERROR"] = "INFO"

    # CORS
    cors_origins: list[str] = ["http://localhost:3000", "http://localhost:8080"]

    # LLM Provider selection
    llm_provider: Literal[
        "openai", "anthropic", "google", "deepseek", "doubao", "glm", "kimi"
    ] = "openai"

    # Pipeline settings
    max_retries: int = 3

    # OpenAI
    openai_api_key: str = ""
    openai_model: str = "gpt-4o"
    openai_base_url: str | None = None

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

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
    )

    def get_llm_config(self) -> dict:
        """Get LLM configuration for the selected provider."""
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
        return {"provider": provider, **configs[provider]}


@lru_cache
def get_settings() -> Settings:
    """Get cached settings instance."""
    return Settings()
