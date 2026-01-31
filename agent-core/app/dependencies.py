"""Dependency injection for FastAPI."""

from functools import lru_cache

from fastapi import HTTPException

from app.config import get_settings
from llm.gateway import LLMGateway


class LLMGatewayError(Exception):
    """Exception raised when LLM gateway cannot be initialized."""

    pass


@lru_cache()
def get_llm_gateway() -> LLMGateway:
    """Get or create a cached LLM gateway instance.

    Uses lru_cache to ensure only one instance is created per configuration.
    The instance is reused across all requests.

    Returns:
        LLMGateway instance configured from settings

    Raises:
        HTTPException: If API key is not configured (503 Service Unavailable)
    """
    settings = get_settings()
    llm_config = settings.get_llm_config()

    # Check if API key is configured
    if not llm_config.get("api_key"):
        raise HTTPException(
            status_code=503,
            detail=f"LLM service unavailable: API key not configured for provider '{llm_config['provider']}'. "
            f"Please set the corresponding environment variable (e.g., OPENAI_API_KEY).",
        )

    try:
        return LLMGateway(
            provider=llm_config["provider"],
            api_key=llm_config["api_key"],
            model=llm_config["model"],
            base_url=llm_config.get("base_url"),
            temperature=llm_config.get("temperature", 0.7),
        )
    except Exception as e:
        # Clear cache on error so next request can retry
        get_llm_gateway.cache_clear()
        raise HTTPException(
            status_code=503,
            detail=f"LLM service unavailable: {str(e)}",
        )


def reset_llm_gateway() -> None:
    """Reset the cached LLM gateway.

    Call this if configuration changes and you need a fresh instance.
    """
    get_llm_gateway.cache_clear()
