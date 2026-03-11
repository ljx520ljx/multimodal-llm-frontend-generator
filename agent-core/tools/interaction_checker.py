"""Interaction Checker Tool - LangChain tool wrapper for state machine validation."""

import re

from bs4 import BeautifulSoup
from langchain_core.tools import tool


@tool
def check_interaction(code: str) -> dict:
    """检查状态机转换是否完整（所有状态是否可达）。

    检查项目：
    1. 状态定义是否存在（x-data 中的 currentState/state 变量）
    2. 各状态视图是否实现（x-show 条件渲染）
    3. 状态转换是否实现（@click 事件处理）
    4. 是否存在死胡同状态（无法离开的状态）

    Args:
        code: 要检查的 HTML 代码

    Returns:
        检查结果字典，包含:
        - complete: bool, 交互是否完整
        - defined_states: list[str], 已定义的状态列表
        - missing_states: list[str], 缺失的状态列表
        - invalid_transitions: list[str], 无效的状态转换
        - dead_end_states: list[str], 死胡同状态（无出口）
        - issues: list[str], 问题描述列表
    """
    issues = []
    defined_states = set()
    target_states = set()

    try:
        soup = BeautifulSoup(code, "lxml")
    except Exception as e:
        return {
            "complete": False,
            "defined_states": [],
            "missing_states": [],
            "invalid_transitions": [f"HTML 解析失败: {str(e)}"],
            "dead_end_states": [],
            "issues": [f"HTML 解析失败: {str(e)}"],
        }

    # 1. 检查状态变量定义
    x_data_elements = soup.find_all(attrs={"x-data": True})
    if not x_data_elements:
        issues.append("缺少 x-data 状态定义")
        return {
            "complete": False,
            "defined_states": [],
            "missing_states": [],
            "invalid_transitions": [],
            "dead_end_states": [],
            "issues": issues,
        }

    # 提取初始状态
    for elem in x_data_elements:
        x_data = elem.get("x-data", "")
        # 匹配 currentState/state/page/step/mode/view/tab/screen 等常见状态变量
        matches = re.findall(
            r"(?:currentState|state|page|currentPage|step|currentStep|mode|view|currentView|tab|activeTab|screen)\s*:\s*['\"]([a-zA-Z_][a-zA-Z0-9_]*)['\"]",
            x_data,
        )
        defined_states.update(matches)

    # 2. 检查 x-show 中定义的状态视图
    x_show_elements = soup.find_all(attrs={"x-show": True})
    for elem in x_show_elements:
        x_show = elem.get("x-show", "")
        matches = re.findall(r"['\"]([a-zA-Z_][a-zA-Z0-9_]*)['\"]", x_show)
        defined_states.update(matches)

    # 3. 检查状态转换（@click 事件）
    click_elements = soup.find_all(attrs={"@click": True})
    x_on_click_elements = soup.find_all(attrs={"x-on:click": True})
    all_click_elements = click_elements + x_on_click_elements

    for elem in all_click_elements:
        click_handler = elem.get("@click") or elem.get("x-on:click", "")
        # 提取目标状态 - 匹配所有常见状态变量的赋值
        matches = re.findall(
            r"(?:currentState|state|page|currentPage|step|currentStep|mode|view|currentView|tab|activeTab|screen)\s*=\s*['\"]([a-zA-Z_][a-zA-Z0-9_]*)['\"]",
            click_handler,
        )
        target_states.update(matches)

    # 4. 分析问题
    # 检查目标状态是否都有对应的视图
    missing_views = target_states - defined_states
    if missing_views:
        issues.append(f"以下状态缺少视图定义: {', '.join(missing_views)}")

    # 检查是否有状态没有入口（除了初始状态）
    states_without_entry = defined_states - target_states
    # 初始状态不需要入口，只要不是所有状态都没入口就行
    if len(states_without_entry) == len(defined_states) and len(defined_states) > 1:
        issues.append("所有状态都没有转换入口，可能缺少状态切换逻辑")

    # 检查是否有足够的状态转换
    if len(all_click_elements) == 0 and len(defined_states) > 1:
        issues.append("缺少点击事件处理，无法触发状态转换")

    # 5. 检查死胡同状态
    # 检查：如果某个状态的视图内没有任何状态转换事件
    dead_end_states = []
    # Check if there are global navigation elements (outside x-show containers)
    # that can transition away from any state
    has_global_nav = False
    for click_elem in all_click_elements:
        # If a click element is NOT inside any x-show element, it's global nav
        parent_x_show = click_elem.find_parent(attrs={"x-show": True})
        if not parent_x_show:
            handler = click_elem.get("@click") or click_elem.get("x-on:click", "")
            if re.search(
                r"(?:currentState|state|page|currentPage|step|currentStep|mode|view|currentView|tab|activeTab|screen)\s*=",
                handler,
            ):
                has_global_nav = True
                break

    if not has_global_nav:
        for elem in x_show_elements:
            x_show = elem.get("x-show", "")
            state_match = re.search(r"['\"]([a-zA-Z_][a-zA-Z0-9_]*)['\"]", x_show)
            if state_match:
                state_name = state_match.group(1)
                # 检查这个状态容器内是否有状态转换的点击事件
                inner_clicks = elem.find_all(attrs={"@click": True})
                inner_x_on_clicks = elem.find_all(attrs={"x-on:click": True})
                has_transition = False
                for ic in inner_clicks + inner_x_on_clicks:
                    handler = ic.get("@click") or ic.get("x-on:click", "")
                    if re.search(
                        r"(?:currentState|state|page|currentPage|step|currentStep|mode|view|currentView|tab|activeTab|screen)\s*=",
                        handler,
                    ):
                        has_transition = True
                        break
                if not has_transition and len(defined_states) > 1:
                    dead_end_states.append(state_name)

    if dead_end_states:
        issues.append(f"以下状态可能是死胡同（无法离开）: {', '.join(dead_end_states)}")

    # 判断是否完整
    is_complete = len(issues) == 0 and len(defined_states) > 0

    return {
        "complete": is_complete,
        "defined_states": list(defined_states),
        "missing_states": list(missing_views),
        "invalid_transitions": [],  # 保留字段以便未来扩展
        "dead_end_states": dead_end_states,
        "issues": issues,
    }
