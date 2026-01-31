package prompt

import "fmt"

// SystemPromptHTML is the system prompt for HTML + Alpine.js generation
const SystemPromptHTML = "你是一位资深前端架构师，专精于 UI 设计稿转代码。\n\n" +
	"## 核心能力\n" +
	"- 精准识别 UI 布局结构和组件层级\n" +
	"- 分析多图差异，推断交互逻辑\n" +
	"- 生成高质量的 HTML + Tailwind CSS + Alpine.js 代码\n\n" +
	"## 响应要求（重要）\n" +
	"收到请求后，立即开始输出分析内容，不要等待。直接从 Step 1 开始输出。\n\n" +
	"## 技术栈\n" +
	"- HTML5 语义化标签\n" +
	"- Tailwind CSS（通过 CDN 加载）\n" +
	"- Alpine.js（轻量级交互框架，通过 CDN 加载）\n" +
	"- 不使用任何构建工具，代码可直接在浏览器运行\n\n" +
	"## Alpine.js 语法快速参考\n" +
	"- x-data=\"{ state: 'idle' }\" - 定义组件状态\n" +
	"- x-show=\"condition\" - 条件显示\n" +
	"- @click=\"state = 'active'\" - 事件绑定\n" +
	"- x-text=\"variable\" - 动态文本\n" +
	"- :class=\"{ 'active': isActive }\" - 动态类名\n" +
	"- x-transition - 过渡动画\n\n" +
	"## 布局要求\n" +
	"- 根容器使用 min-h-screen 确保填满视口\n" +
	"- 主要内容使用 flex + items-center + justify-center 居中\n" +
	"- 所有文本使用中文\n\n" +
	"## 完整性要求（非常重要）\n" +
	"- 必须还原设计稿中的**所有可见元素**，不要省略或简化\n" +
	"- 每个按钮、文字、图标、卡片都要完整实现\n" +
	"- 如果页面元素较多，也要全部生成，不要用注释替代\n\n" +
	"## 链接处理\n" +
	"- 所有 <a> 标签使用 href=\"#\" 或 href=\"javascript:void(0)\"\n" +
	"- 不要使用外部 URL，这是原型预览，不需要真实链接\n\n" +
	"## 交互逻辑要求（非常重要）\n" +
	"- **双向导航**：所有菜单项/Tab/导航都必须支持双向切换\n" +
	"- 不是线性流程（1→2→3），而是任意状态间可以自由切换\n" +
	"- 例如：从\"搜索\"切换到\"商品\"后，点击\"搜索\"应该能返回\n" +
	"- 每个可点击的导航项都要绑定正确的状态切换事件\n" +
	"- 使用 Alpine.js 的状态变量管理当前激活项：如 `x-data=\"{ activeMenu: 'search' }\"`\n" +
	"- 每个菜单项都用 `@click=\"activeMenu = 'xxx'\"` 绑定切换"

// SingleImagePromptHTML is the prompt for single image HTML generation
const SingleImagePromptHTML = "## 单图分析任务\n\n" +
	"请根据这张 UI 设计稿生成完整的 HTML 页面。\n\n" +
	"### 分析步骤\n\n" +
	"**Step 1: 布局识别**\n" +
	"描述整体布局结构（头部、主体、侧边栏、底部等）。\n\n" +
	"**Step 2: 组件识别**\n" +
	"识别主要 UI 组件，为它们命名（如 Header, Card, Button, Form 等）。\n\n" +
	"**Step 3: 样式分析**\n" +
	"分析颜色、字体、间距、圆角等视觉特征。\n\n" +
	"**Step 4: 交互分析**\n" +
	"识别可交互元素（按钮、输入框、链接等），推断可能的交互行为。\n\n" +
	"### 输出格式\n" +
	"1. 先按照上述步骤输出分析过程（每个步骤 1-2 句话）\n" +
	"2. 然后输出完整的 HTML 代码，用 ```html 代码块包裹\n\n" +
	"### 代码要求\n" +
	"- 使用 HTML + Tailwind CSS 实现设计稿中的布局和样式\n" +
	"- 如果页面有可交互元素，使用 Alpine.js 添加基础交互\n" +
	"- 图片使用占位符或 SVG 图标\n" +
	"- 根容器使用 min-h-screen 和 flex 居中"

// MultiImagePromptHTML is the prompt for multi-image HTML generation with 5-step analysis framework
const MultiImagePromptHTML = "## 多图交互分析任务\n\n" +
	"你将看到 %d 张 UI 设计稿，它们展示了同一个页面的不同状态。\n\n" +
	"### 分析步骤（请依次完成）\n\n" +
	"**Step 1: 布局识别**\n" +
	"分别描述每张图的整体布局结构（头部、主体、侧边栏等）。\n\n" +
	"**Step 2: 组件识别**\n" +
	"识别主要 UI 组件，为它们命名（如 Header, Sidebar, Card, Button 等）。\n\n" +
	"**Step 3: 差异检测**\n" +
	"对比相邻图片，列出具体的视觉变化：\n" +
	"- 哪些元素出现/消失了？\n" +
	"- 哪些元素的样式改变了？（颜色、大小、位置）\n" +
	"- 哪些文本内容改变了？\n\n" +
	"**Step 4: 交互推理**\n" +
	"基于差异，推断交互类型：\n" +
	"| 类型 | 特征 |\n" +
	"|------|------|\n" +
	"| Toggle | 元素显示/隐藏切换 |\n" +
	"| Expand | 区域展开/收起 |\n" +
	"| TabSwitch | Tab/标签页切换 |\n" +
	"| FormState | 表单输入状态变化 |\n" +
	"| Hover | 悬停样式变化 |\n" +
	"| Navigation | 页面/视图跳转 |\n\n" +
	"**Step 5: 代码生成**\n" +
	"生成包含以下内容的 HTML + Alpine.js 代码：\n" +
	"- x-data 定义所有识别到的状态\n" +
	"- @click 等事件绑定到触发元素\n" +
	"- x-show/x-bind 实现条件渲染和样式切换\n\n" +
	"### 输出格式（重要）\n" +
	"1. 先按照 Step 1-4 输出分析过程（每步 2-3 句话）\n" +
	"2. 然后用 ```html 代码块输出 Step 5 的完整代码\n\n" +
	"### 代码要求\n" +
	"- 默认显示第一张图的状态\n" +
	"- 点击相应元素可切换到其他状态\n" +
	"- 确保代码能实现所有图片展示的状态切换\n" +
	"- **重要：支持双向切换**，任意状态间都可以相互切换，不只是单向流程\n" +
	"- 导航/菜单/Tab 等元素点击后要能切回之前的状态"

// DiffAnalysisPromptHTML is the prompt for two-image diff analysis with 5-step framework
const DiffAnalysisPromptHTML = "## 双图差异分析任务\n\n" +
	"请对比这两张 UI 设计稿，分析状态变化并生成交互代码。\n\n" +
	"### 分析步骤（请依次完成）\n\n" +
	"**Step 1: 布局识别**\n" +
	"描述页面的整体布局结构。\n\n" +
	"**Step 2: 组件识别**\n" +
	"识别主要 UI 组件并命名。\n\n" +
	"**Step 3: 差异检测**\n" +
	"对比两张图，具体列出：\n" +
	"- 出现/消失的元素\n" +
	"- 样式变化（颜色、大小、位置）\n" +
	"- 内容变化\n\n" +
	"**Step 4: 交互推理**\n" +
	"推断触发这些变化的交互类型（Toggle/Expand/TabSwitch/FormState/Hover/Navigation）。\n\n" +
	"### 输出格式\n" +
	"1. 先按照 Step 1-4 输出分析过程\n" +
	"2. 然后用 ```html 代码块输出完整 HTML 代码\n\n" +
	"### 代码要求\n" +
	"- 默认显示第一张图的状态\n" +
	"- 点击触发元素后切换到第二张图的状态\n" +
	"- **重要：支持双向切换**，点击另一个触发元素可以切回第一张图的状态\n" +
	"- 所有导航/菜单项都要绑定切换事件，不只是单向跳转"

// ChatModifyPromptHTML is the prompt for HTML code modification
const ChatModifyPromptHTML = "## 代码修改任务\n\n" +
	"当前代码：\n```html\n%s\n```\n\n" +
	"用户请求：**%s**\n\n" +
	"### 修改要求（必须遵守）\n" +
	"1. **必须执行用户请求的修改**，这是最重要的\n" +
	"2. 只修改用户指定的部分，保持其他代码不变\n" +
	"3. 确保修改后代码可直接运行\n" +
	"4. 如果用户要求修改颜色，直接修改对应的 Tailwind 类名或 style\n" +
	"5. 如果用户要求修改样式，确保找到正确的元素并应用修改\n\n" +
	"### 输出格式（必须遵守）\n" +
	"1. 先简要说明你要做的修改（1-2句话）\n" +
	"2. 然后用 ```html 代码块输出修改后的**完整** HTML 代码\n\n" +
	"**重要：必须输出完整的 HTML 代码，从 <!DOCTYPE html> 开始到 </html> 结束。**"

// BuildSystemPromptHTML constructs the HTML system prompt
func BuildSystemPromptHTML() string {
	return SystemPromptHTML
}

// BuildSingleImagePromptHTML constructs the single image HTML prompt
func BuildSingleImagePromptHTML() string {
	return SingleImagePromptHTML
}

// BuildMultiImagePromptHTML constructs the multi-image HTML prompt
func BuildMultiImagePromptHTML(imageCount int) string {
	return fmt.Sprintf(MultiImagePromptHTML, imageCount)
}

// BuildDiffAnalysisPromptHTML constructs the diff analysis prompt
func BuildDiffAnalysisPromptHTML() string {
	return DiffAnalysisPromptHTML
}

// BuildChatModifyPromptHTML constructs the chat modification prompt for HTML
func BuildChatModifyPromptHTML(code, message string) string {
	return fmt.Sprintf(ChatModifyPromptHTML, code, message)
}
