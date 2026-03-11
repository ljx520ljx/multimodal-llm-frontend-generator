"""Shared HTML code extraction from LLM responses."""

import re


def extract_html_code(response: str) -> str:
    """Extract HTML code from LLM response.

    Handles markdown code blocks with ```html ... ``` format.

    Args:
        response: Raw LLM response text

    Returns:
        Extracted HTML code or empty string
    """
    # Try to extract from markdown code block
    html_pattern = r"```html\s*([\s\S]*?)\s*```"
    match = re.search(html_pattern, response)
    if match:
        return match.group(1).strip()

    # Try generic code block
    code_pattern = r"```\s*([\s\S]*?)\s*```"
    match = re.search(code_pattern, response)
    if match:
        return match.group(1).strip()

    return ""
