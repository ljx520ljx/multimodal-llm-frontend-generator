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

### 2. 状态机实现——**按 state.scope 选择实现范式**

InteractionSpec 的每个 state/transition 都带 `scope` 字段，**必须按 scope 选择实现模式**，两种模式不可混用：

#### 2.1 `scope == "page"` → 顶层 currentState 字面量
- 顶层 `x-data="{{ currentState: '<initial_state_id>' }}"`
- 每个 page state 必须有对应的 `x-show="currentState === '<state.id>'"` 分支，**state.id 必须原样作为字符串字面量出现**
- 每个 page transition 必须有对应的 `@click="currentState = '<to_state.id>'"`（同样 id 字面量）

```html
<div x-data="{{ currentState: 'home' }}">
    <div x-show="currentState === 'home'">
        <button @click="currentState = 'search'">搜索</button>
    </div>
    <div x-show="currentState === 'search'">
        <button @click="currentState = 'home'">返回首页</button>
    </div>
</div>
```

#### 2.2 `scope == "component"` → 组件局部 x-data 变量
- 组件内部 `x-data="{{ showTooltip: false, value: 50 }}"` 这类**布尔/数值**变量
- `x-show` / `x-bind:class` 直接引用变量（**不要**写 `currentState === 'xxx'`）
- 事件通过 `@mousedown` / `@mouseup` / `@input` / `x-model` 修改变量
- 典型场景：Slider tooltip、Form 校验、Collapse 面板、Popover、Rate 悬停

```html
<!-- Slider 拖动 tooltip：state.id='dragging' scope=component -->
<div x-data="{{ value: 50, dragging: false }}">
    <input type="range" x-model="value"
           @mousedown="dragging = true" @mouseup="dragging = false">
    <div x-show="dragging">当前值：<span x-text="value"></span></div>
</div>

<!-- Collapse 面板：state.id='expanded' scope=component -->
<div x-data="{{ open: false }}">
    <button @click="open = !open">面板标题</button>
    <div x-show="open">面板内容</div>
</div>
```

#### 2.3 混合场景
如果 InteractionSpec 里同时存在 `page` 和 `component` state：
- 顶层 `x-data` 放 `currentState`
- 组件级 `x-data` 独立定义局部布尔/数值变量，两者互不干扰
- 不要把 component state id 写进 `currentState` 切换

#### 2.4 **关键一致性约束**
- **page scope 的 state/transition**：`state.id` 和 `to_state` 都必须作为字符串字面量出现在 `x-show` / `@click` 中（Validator 会硬校验）
- **component scope 的 state/transition**：自由使用有语义的布尔/数值变量名（如 `showTooltip`、`open`、`focused`），不必使用 state.id 字面量

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
