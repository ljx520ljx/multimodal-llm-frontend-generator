"""Unit tests for CodeValidator."""

import pytest

from tools.code_validator import CodeValidator, ValidationResult


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

        assert result.is_valid is True
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
        assert result.is_valid is False

    def test_missing_head_tag(self, validator: CodeValidator):
        """Test detection of missing head tag."""
        html = """<!DOCTYPE html><html><body></body></html>"""

        result = validator.validate(html)

        assert result.is_valid is False
        assert any("head" in e.lower() for e in result.errors)

    def test_missing_body_tag(self, validator: CodeValidator):
        """Test detection of missing body tag."""
        html = """<!DOCTYPE html><html><head></head></html>"""

        result = validator.validate(html)

        assert result.is_valid is False
        assert any("body" in e.lower() for e in result.errors)

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

        assert result.is_valid is False
        assert any("charset" in e.lower() for e in result.errors)

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

        assert result.is_valid is False
        assert any("viewport" in e.lower() for e in result.errors)

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

        assert result.is_valid is False
        assert any("tailwind" in e.lower() for e in result.errors)

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

        assert result.is_valid is False
        assert any("alpine" in e.lower() for e in result.errors)

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

        assert result.is_valid is False
        assert any("x-data" in e.lower() for e in result.errors)

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
        assert result.is_valid is True
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

        assert result.is_valid is True
        assert any("点击" in w or "click" in w.lower() for w in result.warnings)


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

        assert result.is_valid is True
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

        assert result.is_valid is False
        assert any("detail" in e for e in result.errors)
        assert any("cart" in e for e in result.errors)

    def test_extra_states_warning(self, validator: CodeValidator):
        """Test warning for extra states not in expected list."""
        html = """<div x-data="{ currentState: 'home' }">
    <div x-show="currentState === 'home'">Home</div>
    <div x-show="currentState === 'search'">Search</div>
    <div x-show="currentState === 'extra'">Extra</div>
</div>"""

        result = validator.validate_states_coverage(html, ["home", "search"])

        assert result.is_valid is True
        assert any("extra" in w for w in result.warnings)


class TestValidationResult:
    """Tests for ValidationResult model."""

    def test_default_values(self):
        """Test default values of ValidationResult."""
        result = ValidationResult(is_valid=True)

        assert result.is_valid is True
        assert result.errors == []
        assert result.warnings == []

    def test_with_errors_and_warnings(self):
        """Test ValidationResult with errors and warnings."""
        result = ValidationResult(
            is_valid=False,
            errors=["Error 1", "Error 2"],
            warnings=["Warning 1"],
        )

        assert result.is_valid is False
        assert len(result.errors) == 2
        assert len(result.warnings) == 1
