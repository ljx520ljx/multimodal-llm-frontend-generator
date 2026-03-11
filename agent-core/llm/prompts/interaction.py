"""Interaction inference prompt template."""

INTERACTION_INFER_PROMPT = """你是一位专业的交互设计分析师。请分析设计稿之间的交互逻辑，构建状态机模型。

## 核心概念

**状态机模型**：
- 每张设计稿 = 一个状态（State）
- 用户操作（点击等）= 状态转换触发器（Transition）
- 用户可以在状态间自由跳转：1→2→1→3→1→2...（不是线性的 1→2→3）

## 已识别的信息

### 布局
{layout_info}

### 组件
{component_info}

## 分析任务

1. **定义状态**
   - 为每张设计稿创建一个状态
   - 给状态起有意义的 ID（如 home, search, product）
   - 状态名称使用中文（如 首页, 搜索页, 商品页）

2. **推断状态转换**
   - 分析哪些组件点击后会切换到其他状态
   - 确保用户可以从任何状态返回
   - 状态转换应该是双向或多向的，不是单向的

## 输出格式

注意：`summary` 和 `initial_state` 必须放在 `states` 和 `transitions` **之前**输出。
`summary` 必须是一句话的简短概述（不要使用编号列表）。

```json
{{
    "summary": "用户可在首页、搜索页和商品页之间通过导航栏和按钮自由切换",
    "initial_state": "初始状态ID",
    "states": [
        {{
            "id": "状态ID",
            "name": "状态名称（中文）",
            "image_index": 0,
            "description": "状态描述"
        }}
    ],
    "transitions": [
        {{
            "from_state": "起始状态ID",
            "to_state": "目标状态ID",
            "trigger": "触发组件ID",
            "trigger_event": "click",
            "description": "点击XX进入YY页"
        }}
    ]
}}
```

## 重要提醒

1. **状态转换必须支持自由导航**
   - 从首页可以去搜索页，从搜索页也要能回首页
   - 从任何页面都应该能回到首页（通过 Logo 或首页按钮）

2. **识别导航元素**
   - 顶部导航栏的链接
   - 侧边栏菜单项
   - 返回按钮、Logo 等

3. **避免死胡同**
   - 每个状态都应该有至少一个出口
   - 用户不应该被困在某个状态无法离开
"""
