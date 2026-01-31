"""Component detector prompt template."""

COMPONENT_DETECTOR_PROMPT = """你是一位专业的 UI 组件识别专家。请分析给定的设计稿图片，识别所有 UI 组件。

## 已识别的布局信息

{layout_info}

## 分析任务

识别设计稿中的所有 UI 组件，包括：
- **按钮 (button)**: 主按钮、次要按钮、图标按钮等
- **输入框 (input)**: 文本输入、搜索框、密码框等
- **导航 (nav)**: 导航栏、标签页、菜单等
- **卡片 (card)**: 内容卡片、商品卡片等
- **列表 (list)**: 列表项、商品列表等
- **图片 (image)**: 图片、头像、图标等
- **文本 (text)**: 标题、段落、标签等
- **其他**: 任何可识别的 UI 元素

## 输出要求

为每个组件提供：
- **id**: 唯一标识符（如 search_button, user_card_1）
- **type**: 组件类型
- **name**: 代码中使用的名称（PascalCase，如 SearchButton）
- **region**: 所属区域（来自布局分析）
- **properties**: 组件属性
  - text: 按钮文字、标签文字等
  - variant: primary, secondary, ghost 等
  - size: sm, md, lg 等
  - icon: 图标名称（如 search, user）

## 输出格式

```json
{
    "components": [
        {
            "id": "组件ID",
            "type": "组件类型",
            "name": "组件名称",
            "region": "所属区域",
            "properties": {"text": "文字", "variant": "样式"}
        }
    ],
    "summary": "组件概述"
}
```

## 注意事项

- 仔细识别每个可交互元素
- 特别注意导航相关的组件（用于状态切换）
- 如果有多张图片，识别所有图片中的组件
"""
