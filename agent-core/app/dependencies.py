"""Dependency injection for FastAPI."""

from functools import lru_cache

from app.config import get_settings
from llm.gateway import LLMGateway


@lru_cache()
def get_llm_gateway() -> LLMGateway:
    """Get or create a cached LLM gateway instance.

    Uses lru_cache to ensure only one instance is created per configuration.
    The instance is reused across all requests.

    Returns:
        LLMGateway instance configured from settings
    """
    settings = get_settings()
    llm_config = settings.get_llm_config()

    return LLMGateway(
        provider=llm_config["provider"],
        api_key=llm_config["api_key"],
        model=llm_config["model"],
        base_url=llm_config.get("base_url"),
        temperature=llm_config.get("temperature", 0.7),
    )


def reset_llm_gateway() -> None:
    """Reset the cached LLM gateway.

    Call this if configuration changes and you need a fresh instance.
    """
    get_llm_gateway.cache_clear()
