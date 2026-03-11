"""Dependency injection for FastAPI."""

from __future__ import annotations

from functools import lru_cache
from typing import Optional

from fastapi import HTTPException

from app.config import AgentType, get_settings
from llm.gateway import LLMGateway


class LLMGatewayError(Exception):
    """Exception raised when LLM gateway cannot be initialized."""

    pass


# Cache for agent-specific LLM gateways
_llm_gateway_cache: dict[str, LLMGateway] = {}


def _create_llm_gateway(agent_type: AgentType = "default") -> LLMGateway:
    """Create an LLM gateway for the specified agent type.

    Args:
        agent_type: The agent type (default, chat, layout, component, etc.)

    Returns:
        LLMGateway instance configured for the agent type

    Raises:
        HTTPException: If API key is not configured
    """
    settings = get_settings()
    llm_config = settings.get_llm_config(agent_type)

    # Check if API key is configured
    if not llm_config.get("api_key"):
        raise HTTPException(
            status_code=503,
            detail=f"LLM service unavailable: API key not configured for provider '{llm_config['provider']}'. "
            f"Please set the corresponding environment variable (e.g., OPENAI_API_KEY).",
        )

    return LLMGateway(
        provider=llm_config["provider"],
        api_key=llm_config["api_key"],
        model=llm_config["model"],
        base_url=llm_config.get("base_url"),
        temperature=llm_config.get("temperature", 0.7),
        request_timeout=settings.llm_request_timeout,
    )


def get_llm_gateway(agent_type: AgentType = "default") -> LLMGateway:
    """Get or create a cached LLM gateway instance for the specified agent type.

    Uses a cache to ensure only one instance is created per agent type.
    Different agent types can use different models if configured.

    Args:
        agent_type: The agent type to get gateway for:
            - "default": Default configuration
            - "chat": ChatAgent (conversational modification)
            - "layout": LayoutAnalyzer
            - "component": ComponentDetector
            - "interaction": InteractionInfer
            - "codegen": CodeGenerator
            - "validator": CodeValidator

    Returns:
        LLMGateway instance configured for the agent type

    Raises:
        HTTPException: If API key is not configured (503 Service Unavailable)

    Example:
        # Get default gateway
        llm = get_llm_gateway()

        # Get ChatAgent-specific gateway (may use different model)
        chat_llm = get_llm_gateway("chat")
    """
    # Check cache first
    if agent_type in _llm_gateway_cache:
        return _llm_gateway_cache[agent_type]

    try:
        gateway = _create_llm_gateway(agent_type)
        _llm_gateway_cache[agent_type] = gateway
        return gateway
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(
            status_code=503,
            detail=f"LLM service unavailable: {str(e)}",
        )


@lru_cache()
def get_default_llm_gateway() -> LLMGateway:
    """Get the default LLM gateway (backward compatibility).

    This is equivalent to get_llm_gateway("default").
    """
    return get_llm_gateway("default")


def reset_llm_gateway(agent_type: Optional[AgentType] = None) -> None:
    """Reset the cached LLM gateway(s).

    Args:
        agent_type: If specified, only reset that agent's gateway.
                   If None, reset all gateways.
    """
    global _llm_gateway_cache

    if agent_type is None:
        _llm_gateway_cache.clear()
        get_default_llm_gateway.cache_clear()
    elif agent_type in _llm_gateway_cache:
        del _llm_gateway_cache[agent_type]
        if agent_type == "default":
            get_default_llm_gateway.cache_clear()
