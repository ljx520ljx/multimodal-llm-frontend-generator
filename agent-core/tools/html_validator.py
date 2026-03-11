"""HTML Validator Tool - LangChain tool wrapper for HTML validation."""

from langchain_core.tools import tool

from tools.code_validator import CodeValidator


@tool
def validate_html(code: str) -> dict:
    """验证 HTML 代码语法是否正确。

    检查项目：
    1. HTML 基本结构（html, head, body 标签）
    2. 必要的 meta 标签（charset, viewport）
    3. Tailwind CSS 和 Alpine.js CDN 引入
    4. Alpine.js 状态机实现（x-data, x-show）
    5. 状态转换逻辑（@click 事件处理）

    Args:
        code: 要验证的 HTML 代码

    Returns:
        验证结果字典，包含:
        - valid: bool, 是否通过验证
        - errors: list[dict], 结构化错误列表
        - warnings: list[str], 警告列表
    """
    validator = CodeValidator()
    result = validator.validate(code)
    return {
        "valid": result.valid,
        "errors": [e.model_dump() for e in result.errors],
        "warnings": result.warnings,
    }
