"""Unit tests for CodeValidator."""

import pytest

from schemas.code import ValidationError, ValidationResult
from tools.code_validator import CodeValidator


@pytest.fixture
def validator():
    """Create a CodeValidator instance."""
    return CodeValidator()


class TestCodeValidator:
    """Tests for CodeValidator."""

    def test_valid_html(self, validator: CodeValidator):
        """Test validation of valid HTML with Alpine.js."""
        html = """<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Test</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body>
    <div x-data="{ currentState: 'home' }">
        <div x-show="currentState === 'home'">
            <button @click="currentState = 'search'">Go to Search</button>
        </div>
        <div x-show="currentState === 'search'">
            <button @click="currentState = 'home'">Back</button>
        </div>
    </div>
</body>
</html>"""

        result = validator.validate(html)

        assert result.valid is True
        assert len(result.errors) == 0

    def test_missing_html_tag(self, validator: CodeValidator):
        """Test detection of missing HTML tag.

        Note: BeautifulSoup's lxml parser auto-inserts html/head/body tags,
        so we test with an empty string instead which will fail other checks.
        This test verifies the overall validation catches structural issues.
        """
        # lxml parser auto-adds html tag, so just verify validation catches issues
        html = ""

        result = validator.validate(html)

        # Empty HTML should fail validation
        assert result.valid is False

    def test_missing_head_tag(self, validator: CodeValidator):
        """Test detection of missing head tag."""
        html = """<!DOCTYPE html><html><body></body></html>"""

        result = validator.validate(html)

        assert result.valid is False
        assert any("head" in e.message.lower() for e in result.errors)

    def test_missing_body_tag(self, validator: CodeValidator):
        """Test detection of missing body tag."""
        html = """<!DOCTYPE html><html><head></head></html>"""

        result = validator.validate(html)

        assert result.valid is False
        assert any("body" in e.message.lower() for e in result.errors)

    def test_missing_charset(self, validator: CodeValidator):
        """Test detection of missing charset meta tag."""
        html = """<!DOCTYPE html>
<html>
<head>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body></body>
</html>"""

        result = validator.validate(html)

        assert result.valid is False
        assert any("charset" in e.message.lower() for e in result.errors)

    def test_missing_viewport(self, validator: CodeValidator):
        """Test detection of missing viewport meta tag."""
        html = """<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body></body>
</html>"""

        result = validator.validate(html)

        assert result.valid is False
        assert any("viewport" in e.message.lower() for e in result.errors)

    def test_missing_tailwind(self, validator: CodeValidator):
        """Test detection of missing Tailwind CSS."""
        html = """<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body></body>
</html>"""

        result = validator.validate(html)

        assert result.valid is False
        assert any("tailwind" in e.message.lower() for e in result.errors)

    def test_missing_alpine(self, validator: CodeValidator):
        """Test detection of missing Alpine.js."""
        html = """<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body></body>
</html>"""

        result = validator.validate(html)

        assert result.valid is False
        assert any("alpine" in e.message.lower() for e in result.errors)

    def test_missing_x_data(self, validator: CodeValidator):
        """Test detection of missing x-data attribute."""
        html = """<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <script src="https://cdn.tailwindcss.com"></script>
    <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body>
    <div>No state management</div>
</body>
</html>"""

        result = validator.validate(html)

        assert result.valid is False
        assert any("x-data" in e.message.lower() for e in result.errors)

    def test_warning_for_missing_x_show(self, validator: CodeValidator):
        """Test warning for missing x-show elements."""
        html = """<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <script src="https://cdn.tailwindcss.com"></script>
    <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body>
    <div x-data="{ currentState: 'home' }">
        <div>No conditional rendering</div>
    </div>
</body>
</html>"""

        result = validator.validate(html)

        # Should be valid but with warnings
        assert result.valid is True
        assert any("x-show" in w.lower() for w in result.warnings)

    def test_warning_for_missing_click_handlers(self, validator: CodeValidator):
        """Test warning for missing click event handlers."""
        html = """<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <script src="https://cdn.tailwindcss.com"></script>
    <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body>
    <div x-data="{ currentState: 'home' }">
        <div x-show="currentState === 'home'">Home page</div>
    </div>
</body>
</html>"""

        result = validator.validate(html)

        assert result.valid is True
        assert any("点击" in w or "click" in w.lower() for w in result.warnings)

    def test_validation_error_has_suggestion(self, validator: CodeValidator):
        """Test that validation errors include suggestions."""
        html = ""

        result = validator.validate(html)

        assert result.valid is False
        for error in result.errors:
            assert isinstance(error, ValidationError)
            assert error.type in ("syntax", "missing_state", "missing_transition", "alpine_error")
            assert error.message  # non-empty message

    def test_validation_error_types(self, validator: CodeValidator):
        """Test that different checks produce correct error types."""
        # Missing x-data -> missing_state type
        html = """<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <script src="https://cdn.tailwindcss.com"></script>
    <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body><div>No state</div></body>
</html>"""

        result = validator.validate(html)
        state_errors = [e for e in result.errors if e.type == "missing_state"]
        assert len(state_errors) > 0


class TestStatesCoverage:
    """Tests for validate_states_coverage method."""

    def test_all_states_implemented(self, validator: CodeValidator):
        """Test when all expected states are implemented."""
        html = """<div x-data="{ currentState: 'home' }">
    <div x-show="currentState === 'home'">Home</div>
    <div x-show="currentState === 'search'">Search</div>
    <div x-show="currentState === 'detail'">Detail</div>
</div>"""

        result = validator.validate_states_coverage(
            html, ["home", "search", "detail"]
        )

        assert result.valid is True
        assert len(result.errors) == 0

    def test_missing_states(self, validator: CodeValidator):
        """Test detection of missing state implementations."""
        html = """<div x-data="{ currentState: 'home' }">
    <div x-show="currentState === 'home'">Home</div>
    <div x-show="currentState === 'search'">Search</div>
</div>"""

        result = validator.validate_states_coverage(
            html, ["home", "search", "detail", "cart"]
        )

        assert result.valid is False
        assert any("detail" in e.message for e in result.errors)
        assert any("cart" in e.message for e in result.errors)

    def test_extra_states_warning(self, validator: CodeValidator):
        """Test warning for extra states not in expected list."""
        html = """<div x-data="{ currentState: 'home' }">
    <div x-show="currentState === 'home'">Home</div>
    <div x-show="currentState === 'search'">Search</div>
    <div x-show="currentState === 'extra'">Extra</div>
</div>"""

        result = validator.validate_states_coverage(html, ["home", "search"])

        assert result.valid is True
        assert any("extra" in w for w in result.warnings)


class TestScopeAwareValidation:
    """Tests for scope-aware validate_full (page / component partition)."""

    _BASE_HTML_HEAD = (
        '<!DOCTYPE html><html lang="zh-CN"><head>'
        '<meta charset="UTF-8">'
        '<meta name="viewport" content="width=device-width, initial-scale=1.0">'
        '<title>t</title>'
        '<script src="https://cdn.tailwindcss.com"></script>'
        '<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>'
        '</head><body>'
    )
    _BASE_HTML_TAIL = '</body></html>'

    def _wrap(self, body: str) -> str:
        return self._BASE_HTML_HEAD + body + self._BASE_HTML_TAIL

    def _spec(self, states, transitions=None):
        from schemas.interaction import InteractionSpec, State, Transition
        return InteractionSpec(
            summary="t",
            initial_state=states[0]["id"],
            states=[State(**s) for s in states],
            transitions=[Transition(**t) for t in (transitions or [])],
        )

    def test_pure_page_scope_still_hard_fails_on_missing(self, validator: CodeValidator):
        """Pure page-scope spec: missing state is still a hard error."""
        html = self._wrap(
            '<div x-data="{ currentState: \'home\' }">'
            '<div x-show="currentState === \'home\'">H</div>'
            '</div>'
        )
        spec = self._spec([
            {"id": "home", "name": "首页", "image_index": 0, "scope": "page"},
            {"id": "search", "name": "搜索", "image_index": 1, "scope": "page"},
        ])
        result = validator.validate_full(html, spec)
        assert result.valid is False
        assert any("search" in e.message for e in result.errors)

    def test_pure_component_scope_does_not_hard_fail(self, validator: CodeValidator):
        """Pure component-scope spec: id not appearing as literal is OK as
        long as there are Alpine bindings somewhere."""
        html = self._wrap(
            '<div x-data="{ showTooltip: false, value: 50 }">'
            '<input type="range" x-model="value" '
            '@mousedown="showTooltip = true" @mouseup="showTooltip = false">'
            '<div x-show="showTooltip"><span x-text="value"></span></div>'
            '</div>'
        )
        spec = self._spec([
            {"id": "idle", "name": "静止", "image_index": 0, "scope": "component"},
            {"id": "tooltip_visible", "name": "提示显示", "image_index": 1, "scope": "component"},
        ])
        result = validator.validate_full(html, spec)
        # No missing_state error even though neither 'idle' nor
        # 'tooltip_visible' appears as a string literal in x-show.
        assert not any(e.type == "missing_state" for e in result.errors)

    def test_mixed_scope_only_page_states_checked(self, validator: CodeValidator):
        """Mixed spec: only page-scope ids are checked against x-show literals."""
        html = self._wrap(
            '<div x-data="{ currentState: \'list\' }">'
            '<div x-show="currentState === \'list\'">'
            '  <div x-data="{ open: false }">'
            '    <button @click="open = !open">toggle</button>'
            '    <div x-show="open">panel</div>'
            '  </div>'
            '</div>'
            '<div x-show="currentState === \'detail\'">d</div>'
            '</div>'
        )
        spec = self._spec([
            {"id": "list", "name": "列表", "image_index": 0, "scope": "page"},
            {"id": "detail", "name": "详情", "image_index": 1, "scope": "page"},
            {"id": "panel_open", "name": "面板展开", "image_index": 0, "scope": "component"},
        ])
        result = validator.validate_full(html, spec)
        # page ids fully covered, component id not required as literal → valid
        assert result.valid is True

    def test_component_transitions_not_hard_failed(self, validator: CodeValidator):
        """Component-scope transitions are warnings, not errors, even when
        their to_state is never assigned via currentState = '...'."""
        html = self._wrap(
            '<div x-data="{ showTooltip: false, value: 50 }">'
            '<input type="range" x-model="value" '
            '@mousedown="showTooltip = true" @mouseup="showTooltip = false">'
            '<div x-show="showTooltip">v</div>'
            '</div>'
        )
        spec = self._spec(
            states=[
                {"id": "idle", "name": "静止", "image_index": 0, "scope": "component"},
                {"id": "tooltip_visible", "name": "提示", "image_index": 1, "scope": "component"},
            ],
            transitions=[
                {"from_state": "idle", "to_state": "tooltip_visible",
                 "trigger": "slider", "trigger_event": "click", "scope": "component"},
            ],
        )
        result = validator.validate_full(html, spec)
        assert not any(e.type == "missing_transition" for e in result.errors)

    def test_page_state_bindings_do_not_mask_missing_component_binding(
        self, validator: CodeValidator
    ):
        """A top-level page state machine alone is not component interactivity."""
        html = self._wrap(
            '<div x-data="{ currentState: \'home\' }">'
            '<button @click="currentState = \'search\'">search</button>'
            '<div x-show="currentState === \'home\'">h</div>'
            '<div x-show="currentState === \'search\'">s</div>'
            '</div>'
        )
        spec = self._spec([
            {"id": "home", "name": "首页", "image_index": 0, "scope": "page"},
            {"id": "search", "name": "搜索", "image_index": 1, "scope": "page"},
            {"id": "panel_open", "name": "面板展开", "image_index": 1, "scope": "component"},
        ], transitions=[
            {"from_state": "home", "to_state": "search", "trigger": "search_button",
             "trigger_event": "click", "scope": "page"},
        ])

        result = validator.validate_full(html, spec)

        assert result.valid is True
        assert any("组件级状态 (panel_open)" in warning for warning in result.warnings)

    def test_backward_compat_missing_scope_defaults_to_page(self, validator: CodeValidator):
        """Old checkpoints / specs without scope field: default to page and
        still enforce literal coverage."""
        from schemas.interaction import InteractionSpec, State, Transition
        # Construct state objects without passing scope — Pydantic default
        # should be 'page'.
        spec = InteractionSpec(
            summary="t", initial_state="home",
            states=[State(id="home", name="首页", image_index=0),
                    State(id="search", name="搜索", image_index=1)],
            transitions=[],
        )
        assert spec.states[0].scope == "page"
        # And hard failure still fires when literal missing:
        html = self._wrap('<div x-data="{ currentState: \'home\' }">'
                          '<div x-show="currentState === \'home\'">h</div></div>')
        result = validator.validate_full(html, spec)
        assert result.valid is False
        assert any("search" in e.message for e in result.errors)


class TestValidationResult:
    """Tests for ValidationResult model."""

    def test_default_values(self):
        """Test default values of ValidationResult."""
        result = ValidationResult(valid=True)

        assert result.valid is True
        assert result.errors == []
        assert result.warnings == []

    def test_with_errors_and_warnings(self):
        """Test ValidationResult with errors and warnings."""
        result = ValidationResult(
            valid=False,
            errors=[
                ValidationError(type="syntax", message="Error 1"),
                ValidationError(type="missing_state", message="Error 2"),
            ],
            warnings=["Warning 1"],
        )

        assert result.valid is False
        assert len(result.errors) == 2
        assert len(result.warnings) == 1
        assert result.errors[0].type == "syntax"
        assert result.errors[1].type == "missing_state"
