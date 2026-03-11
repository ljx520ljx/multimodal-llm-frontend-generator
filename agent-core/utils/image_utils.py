"""Shared image processing utilities."""

from __future__ import annotations

from typing import Any


def normalize_base64_url(base64_data: str) -> str:
    """Normalize base64 data to a data URL with proper prefix.

    Ensures the base64 string starts with 'data:image/...' scheme.

    Args:
        base64_data: Raw base64 string or data URL

    Returns:
        Data URL string
    """
    if not base64_data.startswith("data:"):
        return f"data:image/png;base64,{base64_data}"
    return base64_data


def build_image_content(base64_data: str) -> dict[str, Any]:
    """Build an image content dict for LLM multimodal messages.

    Args:
        base64_data: Raw base64 string or data URL

    Returns:
        Dict with type 'image_url' and url
    """
    return {
        "type": "image_url",
        "image_url": {"url": normalize_base64_url(base64_data)},
    }
