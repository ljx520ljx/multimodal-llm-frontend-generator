"""Code generation and validation schemas."""

from __future__ import annotations

from typing import Literal, Optional

from pydantic import BaseModel, Field


class GeneratedCode(BaseModel):
    """Generated code result."""

    html: str = Field(description="Complete HTML code with Tailwind CSS and Alpine.js")
    css: Optional[str] = Field(default=None, description="Optional separate CSS (usually None, styles in Tailwind)")
    js: Optional[str] = Field(default=None, description="Optional separate JS (usually None, logic in Alpine.js)")

    @property
    def code(self) -> str:
        """Alias for html field (for compatibility with SDD)."""
        return self.html


class ValidationError(BaseModel):
    """A validation error."""

    type: Literal["syntax", "missing_state", "missing_transition", "alpine_error"] = Field(
        description="Error type"
    )
    message: str = Field(description="Error message")
    line: Optional[int] = Field(default=None, description="Line number if applicable")
    suggestion: str = Field(default="", description="Suggested fix")


class ValidationResult(BaseModel):
    """Code validation result."""

    valid: bool = Field(description="Whether the code is valid")
    errors: list[ValidationError] = Field(
        default_factory=list,
        description="List of validation errors"
    )
    warnings: list[str] = Field(
        default_factory=list,
        description="Non-critical warnings"
    )
