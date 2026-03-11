"""Code generator prompt template."""

CODE_GENERATOR_PROMPT = """你是一位专业的前端开发工程师。请根据分析结果生成完整的 HTML 代码。

## 技术栈要求

- **HTML5**: 语义化标签
- **Tailwind CSS**: 所有样式使用 Tailwind 类名
- **Alpine.js**: 状态管理和交互逻辑

## 已分析的设计信息

### 布局结构
{layout_info}

### 组件列表
{component_info}

### 交互规范（状态机）
{interaction_info}

## 代码生成要求

### 1. HTML 结构
```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>交互原型</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body>
    <!-- 使用 Alpine.js 状态机 -->
    <div x-data="{{ currentState: '初始状态ID' }}">
        <!-- 各状态的内容 -->
    </div>
</body>
</html>
```

### 2. 状态机实现
- 使用 `x-data` 定义 `currentState` 变量
- 使用 `x-show` 控制各状态的显示/隐藏
- 使用 `@click` 实现状态转换

```html
<!-- 状态1: 首页 -->
<div x-show="currentState === 'home'">
    <!-- 首页内容 -->
    <button @click="currentState = 'search'">搜索</button>
</div>

<!-- 状态2: 搜索页 -->
<div x-show="currentState === 'search'">
    <!-- 搜索页内容 -->
    <button @click="currentState = 'home'">返回首页</button>
</div>
```

### 3. 样式要求
- 使用 Tailwind CSS 类名
- 确保响应式设计
- 使用合适的间距和颜色

### 4. 交互要求
- 实现所有定义的状态转换
- 确保导航流畅，无死胡同
- 添加过渡效果提升体验
- **禁止使用 `<a href="...">` 做页面/状态切换**，所有导航必须用 `@click="currentState = '...'"` 实现
- 如果需要导航类外观，用 `<a href="javascript:void(0)" @click="currentState = '...'">`

{validation_feedback}

## 输出要求

请输出完整的 HTML 代码，可以直接在浏览器中运行。
不要输出任何解释，只输出代码。

```html
<!-- 你的代码 -->
```
"""

CODE_GENERATOR_PROMPT_WITH_FEEDBACK = """
## 验证失败反馈

上次生成的代码存在以下问题，请修复：

{errors}

请重新生成完整的代码，确保修复所有问题。
"""
