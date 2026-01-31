"""Code Validator Tool - Validates generated HTML code."""

import re
from typing import TYPE_CHECKING, Optional

from bs4 import BeautifulSoup
from pydantic import BaseModel, Field

if TYPE_CHECKING:
    from schemas.interaction import InteractionSpec


class ValidationResult(BaseModel):
    """Result of HTML code validation."""

    is_valid: bool = Field(description="Whether the code passed validation")
    errors: list[str] = Field(default_factory=list, description="List of validation errors")
    warnings: list[str] = Field(default_factory=list, description="Non-critical warnings")


class CodeValidator:
    """Validates generated HTML/Tailwind/Alpine.js code."""

    def validate(self, html: str) -> ValidationResult:
        """Validate the generated HTML code.

        Checks:
        1. HTML structure validity
        2. Required Alpine.js attributes
        3. State machine implementation
        4. Required Tailwind/Alpine CDN scripts

        Args:
            html: The HTML code to validate

        Returns:
            ValidationResult with is_valid flag and any errors/warnings
        """
        errors = []
        warnings = []

        # Parse HTML
        try:
            soup = BeautifulSoup(html, "lxml")
        except Exception as e:
            return ValidationResult(
                is_valid=False,
                errors=[f"HTML 解析失败: {str(e)}"],
            )

        # Check 1: Basic HTML structure
        structure_errors = self._check_html_structure(soup)
        errors.extend(structure_errors)

        # Check 2: Required CDN scripts
        script_errors = self._check_required_scripts(soup, html)
        errors.extend(script_errors)

        # Check 3: Alpine.js state machine
        alpine_errors, alpine_warnings = self._check_alpine_state_machine(soup)
        errors.extend(alpine_errors)
        warnings.extend(alpine_warnings)

        # Check 4: State transitions
        transition_warnings = self._check_state_transitions(soup)
        warnings.extend(transition_warnings)

        return ValidationResult(
            is_valid=len(errors) == 0,
            errors=errors,
            warnings=warnings,
        )

    def _check_html_structure(self, soup: BeautifulSoup) -> list[str]:
        """Check basic HTML structure."""
        errors = []

        # Check for DOCTYPE
        if not soup.find("html"):
            errors.append("缺少 <html> 标签")

        # Check for head
        if not soup.find("head"):
            errors.append("缺少 <head> 标签")

        # Check for body
        if not soup.find("body"):
            errors.append("缺少 <body> 标签")

        # Check for meta charset
        meta_charset = soup.find("meta", attrs={"charset": True})
        if not meta_charset:
            errors.append("缺少 charset meta 标签")

        # Check for viewport meta
        meta_viewport = soup.find("meta", attrs={"name": "viewport"})
        if not meta_viewport:
            errors.append("缺少 viewport meta 标签")

        return errors

    def _check_required_scripts(self, soup: BeautifulSoup, html: str) -> list[str]:
        """Check for required CDN scripts."""
        errors = []

        # Check for Tailwind CSS
        if "tailwindcss" not in html and "tailwind" not in html.lower():
            errors.append("缺少 Tailwind CSS CDN 引入")

        # Check for Alpine.js
        if "alpinejs" not in html and "alpine" not in html.lower():
            errors.append("缺少 Alpine.js CDN 引入")

        return errors

    def _check_alpine_state_machine(self, soup: BeautifulSoup) -> tuple[list[str], list[str]]:
        """Check Alpine.js state machine implementation."""
        errors = []
        warnings = []

        # Find x-data elements
        x_data_elements = soup.find_all(attrs={"x-data": True})

        if not x_data_elements:
            errors.append("缺少 x-data 状态管理，无法实现状态机")
            return errors, warnings

        # Check for currentState or similar state variable
        has_state_var = False
        for elem in x_data_elements:
            x_data = elem.get("x-data", "")
            if "currentState" in x_data or "state" in x_data or "page" in x_data:
                has_state_var = True
                break

        if not has_state_var:
            warnings.append("建议使用 currentState 变量来管理页面状态")

        # Check for x-show elements
        x_show_elements = soup.find_all(attrs={"x-show": True})
        if not x_show_elements:
            warnings.append("缺少 x-show 条件渲染，可能无法切换状态视图")

        return errors, warnings

    def _check_state_transitions(self, soup: BeautifulSoup) -> list[str]:
        """Check for state transition implementations."""
        warnings = []

        # Find elements with click handlers that modify state
        click_elements = soup.find_all(attrs={"@click": True})
        x_on_click_elements = soup.find_all(attrs={"x-on:click": True})

        all_click_elements = click_elements + x_on_click_elements

        if not all_click_elements:
            warnings.append("缺少点击事件处理，可能无法触发状态转换")
            return warnings

        # Check if any click handler modifies state
        state_modifying_clicks = 0
        for elem in all_click_elements:
            click_handler = elem.get("@click") or elem.get("x-on:click", "")
            if "currentState" in click_handler or "state" in click_handler or "=" in click_handler:
                state_modifying_clicks += 1

        if state_modifying_clicks == 0:
            warnings.append("点击事件可能没有正确修改状态变量")

        # Check for back navigation
        back_patterns = ["返回", "back", "home", "首页"]
        has_back_nav = False
        for elem in all_click_elements:
            text = elem.get_text().lower()
            click_handler = (elem.get("@click") or elem.get("x-on:click", "")).lower()
            for pattern in back_patterns:
                if pattern in text or pattern in click_handler:
                    has_back_nav = True
                    break

        if not has_back_nav and state_modifying_clicks > 1:
            warnings.append("建议添加返回/首页导航，避免用户被困在某个状态")

        return warnings

    def validate_states_coverage(
        self,
        html: str,
        expected_states: list[str],
    ) -> ValidationResult:
        """Validate that all expected states are implemented.

        Args:
            html: The HTML code to validate
            expected_states: List of expected state IDs

        Returns:
            ValidationResult with coverage information
        """
        errors = []
        warnings = []

        # Find all x-show conditions
        soup = BeautifulSoup(html, "lxml")
        x_show_elements = soup.find_all(attrs={"x-show": True})

        implemented_states = set()
        for elem in x_show_elements:
            x_show = elem.get("x-show", "")
            # Extract state names from conditions like "currentState === 'home'"
            matches = re.findall(r"['\"]([a-zA-Z_][a-zA-Z0-9_]*)['\"]", x_show)
            implemented_states.update(matches)

        # Check coverage
        expected_set = set(expected_states)
        missing = expected_set - implemented_states
        extra = implemented_states - expected_set

        if missing:
            errors.append(f"缺少状态实现: {', '.join(missing)}")

        if extra:
            warnings.append(f"发现额外状态: {', '.join(extra)}")

        return ValidationResult(
            is_valid=len(errors) == 0,
            errors=errors,
            warnings=warnings,
        )

    def validate_transitions(
        self,
        html: str,
        interaction_spec: "InteractionSpec",
    ) -> ValidationResult:
        """Validate that all state transitions are implemented.

        This checks that every transition defined in InteractionSpec has
        a corresponding click handler in the generated code.

        Args:
            html: The HTML code to validate
            interaction_spec: The interaction specification with transitions

        Returns:
            ValidationResult with transition coverage information
        """
        errors = []
        warnings = []

        soup = BeautifulSoup(html, "lxml")

        # Find all click handlers
        click_elements = soup.find_all(attrs={"@click": True})
        x_on_click_elements = soup.find_all(attrs={"x-on:click": True})
        all_click_elements = click_elements + x_on_click_elements

        # Extract all state transitions from click handlers
        implemented_transitions = set()
        for elem in all_click_elements:
            click_handler = elem.get("@click") or elem.get("x-on:click", "")
            # Extract target state from patterns like "currentState = 'search'" or "state = 'home'"
            matches = re.findall(r"(?:currentState|state)\s*=\s*['\"]([a-zA-Z_][a-zA-Z0-9_]*)['\"]", click_handler)
            for target_state in matches:
                # We don't know the exact from_state from HTML, but we can check if target exists
                implemented_transitions.add(target_state)

        # Check each transition
        for transition in interaction_spec.transitions:
            # Check if there's a click handler that transitions to the target state
            if transition.to_state not in implemented_transitions:
                errors.append(
                    f"缺少状态转换: {transition.from_state} → {transition.to_state} "
                    f"(触发器: {transition.trigger})"
                )

        # Check for dead-end states (states with no outgoing transitions)
        states_with_outgoing = set(t.from_state for t in interaction_spec.transitions)
        all_states = set(s.id for s in interaction_spec.states)
        dead_ends = all_states - states_with_outgoing

        if dead_ends and len(all_states) > 1:
            warnings.append(f"以下状态没有出口转换（可能导致用户被困）: {', '.join(dead_ends)}")

        return ValidationResult(
            is_valid=len(errors) == 0,
            errors=errors,
            warnings=warnings,
        )

    def validate_full(
        self,
        html: str,
        interaction_spec: "Optional[InteractionSpec]" = None,
    ) -> ValidationResult:
        """Run full validation including state machine checks.

        Args:
            html: The HTML code to validate
            interaction_spec: Optional interaction specification for transition checks

        Returns:
            Combined ValidationResult
        """
        # Run basic validation
        basic_result = self.validate(html)

        if interaction_spec is None:
            return basic_result

        errors = list(basic_result.errors)
        warnings = list(basic_result.warnings)

        # Validate states coverage
        expected_states = [s.id for s in interaction_spec.states]
        states_result = self.validate_states_coverage(html, expected_states)
        errors.extend(states_result.errors)
        warnings.extend(states_result.warnings)

        # Validate transitions
        transitions_result = self.validate_transitions(html, interaction_spec)
        errors.extend(transitions_result.errors)
        warnings.extend(transitions_result.warnings)

        return ValidationResult(
            is_valid=len(errors) == 0,
            errors=errors,
            warnings=warnings,
        )
