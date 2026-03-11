"""Unified LLM Gateway supporting multiple providers."""

from __future__ import annotations

import asyncio
import json
import logging
import re
from typing import Any, AsyncIterator, Optional, Type

import httpx
from langchain_core.language_models import BaseChatModel
from langchain_core.messages import BaseMessage, HumanMessage, SystemMessage
from pydantic import BaseModel, ValidationError

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
        base_url: Optional[str] = None,
        temperature: float = 0.7,
        request_timeout: Optional[int] = None,
    ):
        self.provider = provider.lower()
        self.model = model
        self.temperature = temperature
        self.request_timeout = request_timeout
        self.client = self._create_client(provider, api_key, model, base_url, temperature, request_timeout)
        logger.info(f"LLMGateway initialized: provider={provider}, model={model}, request_timeout={request_timeout}s")

    def _create_client(
        self,
        provider: str,
        api_key: str,
        model: str,
        base_url: Optional[str],
        temperature: float,
        request_timeout: Optional[int] = None,
    ) -> BaseChatModel:
        """Create the appropriate LangChain client.

        Args:
            request_timeout: Per-request timeout in seconds. This is the innermost
                timeout in the chain (LLM 60s < Agent 180s < Handler 240s < SSE 300s).
        """
        provider = provider.lower()
        timeout_kwargs: dict[str, Any] = {}
        if request_timeout is not None:
            timeout_kwargs["request_timeout"] = request_timeout

        if provider == "openai":
            from langchain_openai import ChatOpenAI
            return ChatOpenAI(
                api_key=api_key,
                model=model,
                temperature=temperature,
                base_url=base_url,
                **timeout_kwargs,
            )

        elif provider == "anthropic":
            from langchain_anthropic import ChatAnthropic
            return ChatAnthropic(
                api_key=api_key,
                model=model,
                temperature=temperature,
                **timeout_kwargs,
            )

        elif provider == "google":
            from langchain_google_genai import ChatGoogleGenerativeAI
            return ChatGoogleGenerativeAI(
                google_api_key=api_key,
                model=model,
                temperature=temperature,
                **timeout_kwargs,
            )

        elif provider in PROVIDER_BASE_URLS:
            # OpenAI-compatible providers
            from langchain_openai import ChatOpenAI
            return ChatOpenAI(
                api_key=api_key,
                model=model,
                temperature=temperature,
                base_url=base_url or PROVIDER_BASE_URLS[provider],
                **timeout_kwargs,
            )

        else:
            raise ValueError(f"Unsupported provider: {provider}")

    async def chat(self, messages: list[dict[str, Any]], max_retries: int = 3) -> str:
        """Send a chat request and return the response content.

        Uses exponential backoff retry for transient errors.
        """
        lc_messages = self._to_langchain_messages(messages)
        return await self._retry_with_backoff(
            lambda: self.client.ainvoke(lc_messages),
            max_retries=max_retries,
            extract=lambda r: r.content,
        )

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
        max_retries: int = 3,
    ) -> BaseModel:
        """Send a chat request and return structured output.

        Uses prompt-based JSON extraction instead of function calling,
        which provides compatibility with more LLM providers.
        Uses exponential backoff retry for transient errors.
        """
        # Generate JSON schema description for the prompt
        schema_description = self._generate_schema_description(output_schema)

        # Add JSON format instruction to the last message
        enhanced_messages = self._add_json_instruction(messages, schema_description)

        async def _call():
            response = await self.chat(enhanced_messages, max_retries=1)
            json_data = self._extract_json(response)
            try:
                return output_schema.model_validate(json_data)
            except ValidationError as e:
                logger.error(f"Failed to validate JSON response: {e}")
                logger.error(f"Raw response: {response[:500]}...")
                raise ValueError(f"LLM response does not match expected schema: {e}")

        return await self._retry_with_backoff(
            _call,
            max_retries=max_retries,
            extract=lambda r: r,
        )

    _RETRYABLE_EXCEPTIONS = (
        ConnectionError,
        TimeoutError,
        httpx.ConnectError,
        httpx.ReadTimeout,
        httpx.WriteTimeout,
        httpx.PoolTimeout,
    )

    async def _retry_with_backoff(
        self,
        call,
        max_retries: int = 3,
        initial_delay: float = 1.0,
        extract=None,
    ):
        """Execute an async call with exponential backoff retry.

        Retries on transient errors (connection, timeout, rate limit).
        For 429 responses, uses Retry-After header if available.
        """
        last_exception = None
        for attempt in range(max_retries):
            try:
                result = await call()
                return extract(result) if extract else result
            except self._RETRYABLE_EXCEPTIONS as e:
                last_exception = e
                delay = initial_delay * (2 ** attempt)
                logger.warning(
                    f"LLM call failed (attempt {attempt + 1}/{max_retries}): {e}. "
                    f"Retrying in {delay:.1f}s..."
                )
                await asyncio.sleep(delay)
            except httpx.HTTPStatusError as e:
                if e.response.status_code == 429:
                    last_exception = e
                    retry_after = e.response.headers.get("Retry-After")
                    delay = float(retry_after) if retry_after else initial_delay * (2 ** attempt)
                    logger.warning(
                        f"Rate limited (429, attempt {attempt + 1}/{max_retries}). "
                        f"Retrying in {delay:.1f}s..."
                    )
                    await asyncio.sleep(delay)
                else:
                    raise
            except Exception:
                raise
        raise last_exception

    def _generate_schema_description(self, schema: Type[BaseModel]) -> str:
        """Generate a human-readable schema description for the prompt.

        Handles nested Pydantic models by resolving $ref references.
        """
        json_schema = schema.model_json_schema()
        defs = json_schema.get("$defs", {})

        # Generate example following the complete schema
        example = self._generate_example(json_schema, defs)

        lines = [
            "Output must be a valid JSON object matching this exact structure:",
            "```json",
            json.dumps(example, indent=2, ensure_ascii=False),
            "```",
            "",
            "Schema details:",
        ]

        # Add field descriptions for root object
        self._add_field_descriptions(lines, json_schema, defs, prefix="")

        return "\n".join(lines)

    def _resolve_ref(self, field_info: dict, defs: dict) -> dict:
        """Resolve $ref to actual schema definition."""
        if "$ref" in field_info:
            ref_path = field_info["$ref"]
            # Format: #/$defs/ClassName
            if ref_path.startswith("#/$defs/"):
                ref_name = ref_path.split("/")[-1]
                return defs.get(ref_name, field_info)
        return field_info

    def _generate_example(self, schema: dict, defs: dict) -> Any:
        """Recursively generate example value from schema."""
        # Resolve reference if present
        schema = self._resolve_ref(schema, defs)

        field_type = schema.get("type")

        # Handle allOf (used by Pydantic for Literal types)
        if "allOf" in schema:
            # Take first item
            return self._generate_example(schema["allOf"][0], defs)

        # Handle anyOf (used by Pydantic for Optional types)
        if "anyOf" in schema:
            for option in schema["anyOf"]:
                if option.get("type") != "null":
                    return self._generate_example(option, defs)
            return None

        # Handle const (literal value)
        if "const" in schema:
            return schema["const"]

        # Handle enum
        if "enum" in schema:
            return schema["enum"][0] if schema["enum"] else None

        if field_type == "string":
            title = schema.get("title", "value")
            return f"<{title}>"
        elif field_type == "integer":
            return 0
        elif field_type == "number":
            return 0.0
        elif field_type == "boolean":
            return True
        elif field_type == "array":
            items_schema = schema.get("items", {})
            item_example = self._generate_example(items_schema, defs)
            return [item_example]
        elif field_type == "object":
            properties = schema.get("properties", {})
            obj = {}
            for prop_name, prop_schema in properties.items():
                obj[prop_name] = self._generate_example(prop_schema, defs)
            return obj
        elif "properties" in schema:
            # Object without explicit type
            properties = schema.get("properties", {})
            obj = {}
            for prop_name, prop_schema in properties.items():
                obj[prop_name] = self._generate_example(prop_schema, defs)
            return obj

        return None

    def _add_field_descriptions(
        self,
        lines: list[str],
        schema: dict,
        defs: dict,
        prefix: str = "",
    ) -> None:
        """Add field descriptions recursively."""
        schema = self._resolve_ref(schema, defs)
        properties = schema.get("properties", {})
        required = schema.get("required", [])

        for field_name, field_info in properties.items():
            resolved = self._resolve_ref(field_info, defs)
            description = resolved.get("description", field_info.get("description", ""))
            field_type = resolved.get("type", "object")
            is_required = "required" if field_name in required else "optional"

            full_name = f"{prefix}{field_name}" if prefix else field_name

            # Handle enum/literal
            if "enum" in resolved:
                enum_values = ", ".join(f'"{v}"' for v in resolved["enum"])
                lines.append(f"- `{full_name}` (enum: {enum_values}, {is_required}): {description}")
            elif field_type == "array":
                items_schema = resolved.get("items", {})
                items_resolved = self._resolve_ref(items_schema, defs)
                items_type = items_resolved.get("type", "object")
                lines.append(f"- `{full_name}` (array of {items_type}, {is_required}): {description}")
                # If array items are objects, describe their fields
                if "properties" in items_resolved:
                    self._add_field_descriptions(lines, items_resolved, defs, prefix=f"{full_name}[].")
            elif "properties" in resolved:
                lines.append(f"- `{full_name}` (object, {is_required}): {description}")
                self._add_field_descriptions(lines, resolved, defs, prefix=f"{full_name}.")
            else:
                lines.append(f"- `{full_name}` ({field_type}, {is_required}): {description}")

    def _add_json_instruction(
        self,
        messages: list[dict[str, Any]],
        schema_description: str,
    ) -> list[dict[str, Any]]:
        """Add JSON format instruction to messages."""
        json_instruction = f"""

IMPORTANT: You must respond with a valid JSON object only. Do not include any other text before or after the JSON.

{schema_description}

Remember: Output ONLY the JSON object, no markdown code blocks, no explanations."""

        # Deep copy messages to avoid modifying original
        enhanced = []
        for msg in messages:
            enhanced.append(dict(msg))

        # Append instruction to the last user message
        if enhanced and enhanced[-1].get("role") == "user":
            content = enhanced[-1].get("content", "")
            if isinstance(content, str):
                enhanced[-1]["content"] = content + json_instruction
            elif isinstance(content, list):
                # Multimodal message - append to text parts
                new_content = list(content)
                new_content.append({"type": "text", "text": json_instruction})
                enhanced[-1]["content"] = new_content
        else:
            # Add as a new user message
            enhanced.append({"role": "user", "content": json_instruction})

        return enhanced

    def _extract_json(self, response: str) -> dict:
        """Extract JSON from LLM response.

        Handles various response formats:
        - Pure JSON
        - JSON wrapped in markdown code blocks
        - JSON with surrounding text
        """
        response = response.strip()

        # Try direct JSON parse first
        try:
            return json.loads(response)
        except json.JSONDecodeError:
            pass

        # Try to extract from markdown code block
        code_block_pattern = r"```(?:json)?\s*\n?([\s\S]*?)\n?```"
        matches = re.findall(code_block_pattern, response)
        for match in matches:
            try:
                return json.loads(match.strip())
            except json.JSONDecodeError:
                continue

        # Try to find JSON object in response
        # Look for content between first { and last }
        first_brace = response.find("{")
        last_brace = response.rfind("}")
        if first_brace != -1 and last_brace != -1 and last_brace > first_brace:
            json_str = response[first_brace:last_brace + 1]
            try:
                return json.loads(json_str)
            except json.JSONDecodeError:
                pass

        # Try to find JSON array
        first_bracket = response.find("[")
        last_bracket = response.rfind("]")
        if first_bracket != -1 and last_bracket != -1 and last_bracket > first_bracket:
            json_str = response[first_bracket:last_bracket + 1]
            try:
                return json.loads(json_str)
            except json.JSONDecodeError:
                pass

        logger.error(f"Failed to extract JSON from response: {response[:500]}...")
        raise ValueError("Could not extract valid JSON from LLM response")

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
