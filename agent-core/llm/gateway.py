"""Unified LLM Gateway supporting multiple providers."""

import logging
from typing import Any, AsyncIterator, Type

from langchain_core.language_models import BaseChatModel
from langchain_core.messages import BaseMessage, HumanMessage, SystemMessage
from pydantic import BaseModel

logger = logging.getLogger(__name__)

# OpenAI-compatible provider base URLs
PROVIDER_BASE_URLS = {
    "deepseek": "https://api.deepseek.com/v1",
    "doubao": "https://ark.cn-beijing.volces.com/api/v3",
    "glm": "https://open.bigmodel.cn/api/paas/v4",
    "kimi": "https://api.moonshot.cn/v1",
}


class LLMGateway:
    """Unified LLM gateway supporting multiple providers.

    Supported providers:
    - openai: OpenAI API (gpt-4o, gpt-4o-mini, etc.)
    - anthropic: Anthropic API (claude-3-5-sonnet, etc.)
    - google: Google Generative AI (gemini-1.5-pro, etc.)
    - deepseek: DeepSeek API (OpenAI compatible)
    - doubao: Doubao/Volcengine API (OpenAI compatible)
    - glm: Zhipu GLM API (OpenAI compatible)
    - kimi: Moonshot Kimi API (OpenAI compatible)
    """

    def __init__(
        self,
        provider: str,
        api_key: str,
        model: str,
        base_url: str | None = None,
        temperature: float = 0.7,
    ):
        self.provider = provider.lower()
        self.model = model
        self.temperature = temperature
        self.client = self._create_client(provider, api_key, model, base_url, temperature)
        logger.info(f"LLMGateway initialized: provider={provider}, model={model}")

    def _create_client(
        self,
        provider: str,
        api_key: str,
        model: str,
        base_url: str | None,
        temperature: float,
    ) -> BaseChatModel:
        """Create the appropriate LangChain client."""
        provider = provider.lower()

        if provider == "openai":
            from langchain_openai import ChatOpenAI
            return ChatOpenAI(
                api_key=api_key,
                model=model,
                temperature=temperature,
                base_url=base_url,
            )

        elif provider == "anthropic":
            from langchain_anthropic import ChatAnthropic
            return ChatAnthropic(
                api_key=api_key,
                model=model,
                temperature=temperature,
            )

        elif provider == "google":
            from langchain_google_genai import ChatGoogleGenerativeAI
            return ChatGoogleGenerativeAI(
                google_api_key=api_key,
                model=model,
                temperature=temperature,
            )

        elif provider in PROVIDER_BASE_URLS:
            # OpenAI-compatible providers
            from langchain_openai import ChatOpenAI
            return ChatOpenAI(
                api_key=api_key,
                model=model,
                temperature=temperature,
                base_url=base_url or PROVIDER_BASE_URLS[provider],
            )

        else:
            raise ValueError(f"Unsupported provider: {provider}")

    async def chat(self, messages: list[dict[str, Any]]) -> str:
        """Send a chat request and return the response content."""
        lc_messages = self._to_langchain_messages(messages)
        response = await self.client.ainvoke(lc_messages)
        return response.content

    async def chat_stream(self, messages: list[dict[str, Any]]) -> AsyncIterator[str]:
        """Send a chat request and stream the response."""
        lc_messages = self._to_langchain_messages(messages)
        async for chunk in self.client.astream(lc_messages):
            if chunk.content:
                yield chunk.content

    async def chat_structured(
        self,
        messages: list[dict[str, Any]],
        output_schema: Type[BaseModel],
    ) -> BaseModel:
        """Send a chat request and return structured output."""
        lc_messages = self._to_langchain_messages(messages)
        structured_llm = self.client.with_structured_output(output_schema)
        return await structured_llm.ainvoke(lc_messages)

    def _to_langchain_messages(self, messages: list[dict[str, Any]]) -> list[BaseMessage]:
        """Convert dict messages to LangChain message objects."""
        result = []
        for msg in messages:
            role = msg.get("role", "user")
            content = msg.get("content", "")

            if role == "system":
                result.append(SystemMessage(content=content))
            elif role == "user":
                # Handle multimodal content
                if isinstance(content, list):
                    result.append(HumanMessage(content=content))
                else:
                    result.append(HumanMessage(content=content))
            elif role == "assistant":
                from langchain_core.messages import AIMessage
                result.append(AIMessage(content=content))

        return result
