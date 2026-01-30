package prompt

import "fmt"

// Framework names for display
const (
	ReactDisplayName = "React"
	VueDisplayName   = "Vue"
)

// GetFrameworkDisplayName returns the display name for a framework
func GetFrameworkDisplayName(framework string) string {
	switch framework {
	case "react":
		return ReactDisplayName
	case "vue":
		return VueDisplayName
	default:
		return ReactDisplayName
	}
}

// SystemPromptV2 is the optimized system prompt for code generation
const SystemPromptV2 = `你是一位资深前端架构师，专精于 UI 设计稿转代码。

## 核心能力
- 精准识别 UI 布局结构和组件层级
- 分析多图差异，推断交互逻辑
- 生成高质量、零语法错误的 %s 代码

## 技术约束
- 使用 %s 函数式组件
- 使用 Tailwind CSS（不使用 @apply）
- 代码必须在 Sandpack 沙箱中可直接运行
- 不使用外部图片 URL，使用占位符或内联 SVG
- 所有文本使用中文

## 布局要求（重要）
- 根组件容器必须使用 min-h-screen 确保填满视口高度
- 主要内容应使用 flex + items-center + justify-center 实现水平垂直居中
- 典型结构：<div className="min-h-screen bg-xxx flex items-center justify-center">

## 代码质量检查清单（必须遵守）
生成代码前，请确保：
1. ✅ 所有 JSX 标签正确闭合（自闭合标签用 />）
2. ✅ 所有括号成对匹配：{ } [ ] ( )
3. ✅ 字符串使用引号包裹，className 值用双引号
4. ✅ 事件处理器语法正确：onClick={() => ...} 或 onClick={handleClick}
5. ✅ 条件渲染语法正确：{condition && <Element />} 或 {condition ? <A /> : <B />}
6. ✅ map 返回值有唯一 key 属性
7. ✅ 没有未定义的变量或函数
8. ✅ import 语句在文件顶部且语法正确

## 常见错误避免
- ❌ 错误：onClick={setCount(count + 1)} → ✅ 正确：onClick={() => setCount(count + 1)}
- ❌ 错误：<div className=text-red-500> → ✅ 正确：<div className="text-red-500">
- ❌ 错误：{items.map(item => <div>...</div>)} → ✅ 正确：{items.map(item => <div key={item.id}>...</div>)}
- ❌ 错误：</button>> → ✅ 正确：</button>

## 输出格式
1. 先用 <thinking> 标签描述分析过程
2. 然后输出完整的代码块（用 ` + "```" + `jsx 或 ` + "```" + `tsx 包裹）
3. 代码必须包含所有必要的 import 语句
4. 导出默认组件名为 App
5. 生成后请在心里检查上述质量清单`

// SingleImagePromptV2 is the optimized prompt for single image code generation
const SingleImagePromptV2 = `请根据这张 UI 设计稿生成完整的 %s 代码。

### 分析步骤

**Step 1: 布局识别**
描述整体布局结构（头部、主体、侧边栏、底部等）。

**Step 2: 组件识别**
识别主要 UI 组件，为它们命名（如 Header, Sidebar, Card, Button 等）。

**Step 3: 样式分析**
分析颜色方案、字体大小、间距和视觉层次。

**Step 4: 代码生成**
生成完整的 React 组件代码：
- 合理的组件结构
- 使用 Tailwind CSS 实现样式
- 语义化的 HTML 结构
- 根容器使用 min-h-screen 和 flex 居中布局

### 输出要求
- <thinking> 中完成 Step 1-3 的分析
- 代码块中输出 Step 4 的完整代码
- 根组件结构：<div className="min-h-screen flex items-center justify-center ...">
- 确保代码可直接运行`

// MultiImagePromptV2 is the optimized prompt for multi-image interaction analysis
const MultiImagePromptV2 = `## 多图交互分析任务

你将看到 %d 张 UI 设计稿，它们展示了同一个页面的不同状态。

### 分析步骤

**Step 1: 布局识别**
分别描述每张图的整体布局结构（头部、主体、侧边栏等）。

**Step 2: 组件识别**
识别主要 UI 组件，为它们命名（如 Header, Sidebar, Card, Button 等）。

**Step 3: 差异检测**
对比相邻图片，列出具体的视觉变化：
- 哪些元素出现/消失了？
- 哪些元素的样式改变了？（颜色、大小、位置）
- 哪些文本内容改变了？

**Step 4: 交互推理**
基于差异，推断交互类型：
| 类型 | 特征 | 代码模式 |
|------|------|----------|
| Toggle | 元素显示/隐藏切换 | useState<boolean> + onClick |
| Expand | 区域展开/收起 | useState<boolean> + 高度过渡 |
| TabSwitch | Tab/标签页切换 | useState<string> + onClick |
| FormState | 表单输入状态变化 | useState + onChange |
| Hover | 悬停样式变化 | hover: 类或 onMouseEnter |
| Navigation | 页面/视图跳转 | useState<string> |

**Step 5: 代码生成**
生成包含以下内容的 %s 代码：
- useState 管理所有识别到的状态
- onClick/onChange 处理器绑定到触发元素
- 条件渲染实现所有图片中展示的状态
- 根容器使用 min-h-screen 和 flex 居中布局

### 输出要求
- <thinking> 中完成 Step 1-4 的分析
- 代码块中输出 Step 5 的完整代码
- 根组件结构：<div className="min-h-screen flex items-center justify-center ...">
- 确保代码能实现所有图片展示的状态切换`

// DiffAnalysisPromptV2 is the optimized prompt for two-image diff analysis
const DiffAnalysisPromptV2 = `## 双图差异分析任务

请对比这两张 UI 设计稿，分析状态变化并生成交互代码。

### 分析步骤

**Step 1: 图片描述**
分别简述两张图的主要内容。

**Step 2: 差异检测**
详细列出两张图之间的视觉差异：
- 显示/隐藏的元素
- 样式变化（颜色、大小、位置）
- 内容变化（文本、图标）

**Step 3: 交互推理**
基于差异推断：
- 交互类型（Toggle/Expand/TabSwitch/FormState/Hover/Navigation）
- 触发元素（哪个元素被点击/操作）
- 状态变量（需要哪些 useState）

**Step 4: 代码生成**
生成完整的 React 代码：
- 定义必要的状态变量
- 实现事件处理函数
- 使用条件渲染切换状态
- 根容器使用 min-h-screen 和 flex 居中布局

### 输出要求
- <thinking> 中完成 Step 1-3 的分析
- 代码块中输出完整可运行的代码
- 根组件结构：<div className="min-h-screen flex items-center justify-center ...">
- 默认显示第一张图的状态，点击可切换到第二张图的状态`

// ChatModifyPromptV2 is the optimized prompt for chat-based code modification
const ChatModifyPromptV2 = `## 代码增量修改任务

当前代码：
` + "```" + `jsx
%s
` + "```" + `

用户请求：%s

### 核心原则：增量修改
**只修改用户指定的部分，严格保持其他所有代码不变。**

### 修改步骤
1. **定位**：找到需要修改的具体代码位置
2. **分析**：理解用户的修改意图（样式、功能、文本等）
3. **修改**：仅修改相关代码，不改动其他任何部分
4. **验证**：确保修改后代码语法正确、可运行

### 禁止行为
- ❌ 重构或"优化"未被提及的代码
- ❌ 修改代码格式、缩进或空行
- ❌ 添加或删除未被要求的功能
- ❌ 更改变量名、函数名（除非明确要求）

### 输出格式
1. <thinking> 中说明：修改了哪里、为什么这样修改
2. 输出完整代码（必须包含所有原有代码，仅更改指定部分）`

// BuildSystemPromptV2 constructs the V2 system prompt for the given framework
func BuildSystemPromptV2(framework string) string {
	displayName := GetFrameworkDisplayName(framework)
	return fmt.Sprintf(SystemPromptV2, displayName, displayName)
}

// BuildSingleImagePromptV2 constructs the V2 single image prompt
func BuildSingleImagePromptV2(framework string) string {
	displayName := GetFrameworkDisplayName(framework)
	return fmt.Sprintf(SingleImagePromptV2, displayName)
}

// BuildMultiImagePromptV2 constructs the V2 multi-image prompt
func BuildMultiImagePromptV2(imageCount int, framework string) string {
	displayName := GetFrameworkDisplayName(framework)
	return fmt.Sprintf(MultiImagePromptV2, imageCount, displayName)
}

// BuildChatModifyPromptV2 constructs the V2 chat modification prompt
func BuildChatModifyPromptV2(code, message string) string {
	return fmt.Sprintf(ChatModifyPromptV2, code, message)
}
