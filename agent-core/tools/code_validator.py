"""Code Validator Tool - Validates generated HTML code."""

from __future__ import annotations

import re
from typing import TYPE_CHECKING, Optional

from bs4 import BeautifulSoup

if TYPE_CHECKING:
    from schemas.interaction import InteractionSpec

from schemas.code import ValidationError, ValidationResult


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
            ValidationResult with valid flag and any errors/warnings
        """
        errors: list[ValidationError] = []
        warnings: list[str] = []

        # Parse HTML
        try:
            soup = BeautifulSoup(html, "lxml")
        except Exception as e:
            return ValidationResult(
                valid=False,
                errors=[ValidationError(
                    type="syntax",
                    message=f"HTML 解析失败: {str(e)}",
                    suggestion="检查 HTML 语法是否正确",
                )],
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
            valid=len(errors) == 0,
            errors=errors,
            warnings=warnings,
        )

    def _check_html_structure(self, soup: BeautifulSoup) -> list[ValidationError]:
        """Check basic HTML structure."""
        errors: list[ValidationError] = []

        if not soup.find("html"):
            errors.append(ValidationError(
                type="syntax", message="缺少 <html> 标签",
                suggestion="添加 <html> 根标签",
            ))

        if not soup.find("head"):
            errors.append(ValidationError(
                type="syntax", message="缺少 <head> 标签",
                suggestion="添加 <head> 标签并包含必要的 meta 信息",
            ))

        if not soup.find("body"):
            errors.append(ValidationError(
                type="syntax", message="缺少 <body> 标签",
                suggestion="添加 <body> 标签",
            ))

        meta_charset = soup.find("meta", attrs={"charset": True})
        if not meta_charset:
            errors.append(ValidationError(
                type="syntax", message="缺少 charset meta 标签",
                suggestion='在 <head> 中添加 <meta charset="UTF-8">',
            ))

        meta_viewport = soup.find("meta", attrs={"name": "viewport"})
        if not meta_viewport:
            errors.append(ValidationError(
                type="syntax", message="缺少 viewport meta 标签",
                suggestion='添加 <meta name="viewport" content="width=device-width, initial-scale=1.0">',
            ))

        return errors

    def _check_required_scripts(self, soup: BeautifulSoup, html: str) -> list[ValidationError]:
        """Check for required CDN scripts."""
        errors: list[ValidationError] = []

        if "tailwindcss" not in html and "tailwind" not in html.lower():
            errors.append(ValidationError(
                type="syntax", message="缺少 Tailwind CSS CDN 引入",
                suggestion='在 <head> 中添加 <script src="https://cdn.tailwindcss.com"></script>',
            ))

        if "alpinejs" not in html and "alpine" not in html.lower():
            errors.append(ValidationError(
                type="alpine_error", message="缺少 Alpine.js CDN 引入",
                suggestion='在 <head> 中添加 <script src="https://unpkg.com/alpinejs" defer></script>',
            ))

        return errors

    def _check_alpine_state_machine(self, soup: BeautifulSoup) -> tuple[list[ValidationError], list[str]]:
        """Check Alpine.js state machine implementation."""
        errors: list[ValidationError] = []
        warnings: list[str] = []

        # Find x-data elements
        x_data_elements = soup.find_all(attrs={"x-data": True})

        if not x_data_elements:
            errors.append(ValidationError(
                type="missing_state", message="缺少 x-data 状态管理，无法实现状态机",
                suggestion="添加 x-data 属性定义状态，如 x-data=\"{ currentState: 'home' }\"",
            ))
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
        errors: list[ValidationError] = []
        warnings: list[str] = []

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
            errors.append(ValidationError(
                type="missing_state",
                message=f"缺少状态实现: {', '.join(missing)}",
                suggestion=f"为以下状态添加 x-show 条件渲染: {', '.join(missing)}",
            ))

        if extra:
            warnings.append(f"发现额外状态: {', '.join(extra)}")

        return ValidationResult(
            valid=len(errors) == 0,
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
        errors: list[ValidationError] = []
        warnings: list[str] = []

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
                implemented_transitions.add(target_state)

        # Check each transition
        for transition in interaction_spec.transitions:
            if transition.to_state not in implemented_transitions:
                errors.append(ValidationError(
                    type="missing_transition",
                    message=f"缺少状态转换: {transition.from_state} -> {transition.to_state} (触发器: {transition.trigger})",
                    suggestion=f"添加 @click 事件将状态从 '{transition.from_state}' 转换到 '{transition.to_state}'",
                ))

        # Check for dead-end states (states with no outgoing transitions)
        states_with_outgoing = set(t.from_state for t in interaction_spec.transitions)
        all_states = set(s.id for s in interaction_spec.states)
        dead_ends = all_states - states_with_outgoing

        if dead_ends and len(all_states) > 1:
            warnings.append(f"以下状态没有出口转换（可能导致用户被困）: {', '.join(dead_ends)}")

        return ValidationResult(
            valid=len(errors) == 0,
            errors=errors,
            warnings=warnings,
        )

    def validate_full(
        self,
        html: str,
        interaction_spec: "Optional[InteractionSpec]" = None,
    ) -> ValidationResult:
        """Run full validation including state machine checks.

        Scope-aware: only ``scope == "page"`` states/transitions are enforced
        as hard errors (they must appear as ``currentState === 'id'`` /
        ``currentState = 'id'`` literals). ``scope == "component"`` entries
        are degraded to warnings because their correct Alpine implementation
        (local boolean / numeric ``x-data`` variables) does not contain the
        state id as a literal.

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

        # Partition states by scope.
        # Back-compat: states without explicit scope default to "page" (Pydantic default).
        page_states = [s for s in interaction_spec.states if getattr(s, "scope", "page") == "page"]
        component_states = [s for s in interaction_spec.states if getattr(s, "scope", "page") == "component"]

        # Hard coverage check for page-scope states only.
        if page_states:
            page_result = self.validate_states_coverage(html, [s.id for s in page_states])
            errors.extend(page_result.errors)
            warnings.extend(page_result.warnings)

        # Soft check for component-scope states: we only require *some* local
        # Alpine binding exists (x-data / x-show / x-model / @input / @click).
        # Missing such binding is a warning, not an error.
        if component_states and not self._has_local_alpine_bindings(html):
            warnings.append(
                f"组件级状态 ({', '.join(s.id for s in component_states)}) "
                "未检测到任何 x-data 局部变量或 x-show / x-model / @click 绑定"
            )

        # Partition transitions by scope (fall back via source state's scope).
        state_scope_map = {s.id: getattr(s, "scope", "page") for s in interaction_spec.states}
        page_transitions = []
        component_transitions = []
        for t in interaction_spec.transitions:
            t_scope = getattr(t, "scope", None) or state_scope_map.get(t.from_state, "page")
            if t_scope == "page":
                page_transitions.append(t)
            else:
                component_transitions.append(t)

        if page_transitions:
            # Build a lightweight spec with only page transitions for strict check.
            page_only_spec = interaction_spec.model_copy(update={
                "states": page_states or interaction_spec.states,
                "transitions": page_transitions,
            })
            transitions_result = self.validate_transitions(html, page_only_spec)
            errors.extend(transitions_result.errors)
            warnings.extend(transitions_result.warnings)

        if component_transitions:
            warnings.append(
                f"{len(component_transitions)} 条组件级 transition 未做字面量校验（"
                "组件级交互通过局部变量/事件实现，不强制 currentState 字面量）"
            )

        return ValidationResult(
            valid=len(errors) == 0,
            errors=errors,
            warnings=warnings,
        )

    @staticmethod
    def _has_local_alpine_bindings(html: str) -> bool:
        """Cheap check: does the HTML contain any Alpine binding beyond the
        top-level currentState wrapper? Used as a soft check for component-
        scope states."""
        soup = BeautifulSoup(html, "html.parser")
        local_only_directives = {"x-model", "@input", "@change", "@mousedown", "@mouseup"}

        for tag in soup.find_all(True):
            attrs = tag.attrs
            if any(name in attrs for name in local_only_directives):
                return True

            x_data = attrs.get("x-data")
            if isinstance(x_data, str) and not re.search(r"\bcurrentState\b", x_data):
                return True

            for directive in ("x-show", "@click"):
                expression = attrs.get(directive)
                if isinstance(expression, str) and not re.search(r"\bcurrentState\b", expression):
                    return True

        return False
