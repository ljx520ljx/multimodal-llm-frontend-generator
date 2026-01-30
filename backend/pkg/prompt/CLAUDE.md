# L3 - Prompt 模板文档 | Prompt Templates

> DocOps 层级: **L3 (Implementation Level)**
> 父文档: [L2 后端模块文档](../../CLAUDE.md)

## 模块职责

管理 LLM 调用的 Prompt 模板，使用优化的结构化分析框架。

## 目录结构

```
pkg/prompt/
├── templates_html.go      # HTML + Alpine.js 模板（当前使用）
├── templates_v2.go        # React 模板（备用）
├── interaction_types.go   # 交互类型定义
├── few_shot.go           # Few-Shot 示例
└── CLAUDE.md             # 本文档
```

## 技术方案

### 当前方案：HTML + Alpine.js

采用轻量级方案，生成可直接在浏览器运行的代码：

- **HTML5** 语义化标签
- **Tailwind CSS** 通过 CDN 加载
- **Alpine.js** 轻量级交互框架

**优势**：
- 无需编译，生成即可运行
- 稳定性高，避免 React/Sandpack 编译问题
- 适合快速原型验证

### 备用方案：React + Sandpack

保留 `templates_v2.go` 中的 React 模板，供需要时切换。

## Prompt 模板（HTML 模式）

| 模板 | 用途 | 函数 |
|------|------|------|
| `SystemPromptHTML` | 系统角色设定 | `BuildSystemPromptHTML()` |
| `SingleImagePromptHTML` | 单图分析（4步） | `BuildSingleImagePromptHTML()` |
| `MultiImagePromptHTML` | 多图分析（5步） | `BuildMultiImagePromptHTML(count)` |
| `DiffAnalysisPromptHTML` | 双图差异分析 | `BuildDiffAnalysisPromptHTML()` |
| `ChatModifyPromptHTML` | 对话修改代码 | `BuildChatModifyPromptHTML(code, msg)` |

## 五步分析框架

Prompt 引导 LLM 按以下步骤执行：

```
Step 1: 布局识别 → 识别整体 UI 结构
Step 2: 组件识别 → 命名主要组件
Step 3: 差异检测 → 对比图片变化（多图模式）
Step 4: 交互推理 → 推断交互类型
Step 5: 代码生成 → 输出 HTML + Alpine.js 代码
```

### 输出格式要求

Prompt 要求 AI 按以下格式输出：

```
Step 1: 布局识别 整体布局采用...
Step 2: 组件识别 主要组件包括...
Step 3: 差异检测 图1到图2的变化...
Step 4: 交互推理 这是一个 Toggle 类型交互...

```html
<!DOCTYPE html>
<html>...</html>
```
```

## 交互类型分类

定义 6 种核心交互类型：

| 类型 | 识别特征 | Alpine.js 模式 |
|------|----------|----------------|
| `Toggle` | 元素显示/隐藏 | `x-show="visible"` |
| `Expand` | 区域展开/收起 | `x-show` + `x-transition` |
| `TabSwitch` | Tab 高亮切换 | `x-data="{ tab: 'a' }"` |
| `FormState` | 表单状态变化 | `x-data="{ state: 'idle' }"` |
| `Hover` | 悬停样式变化 | `:class` + `@mouseenter` |
| `Navigation` | 页面/视图跳转 | `x-show="page === 'home'"` |

## Alpine.js 语法参考

Prompt 中包含 Alpine.js 快速参考：

```javascript
x-data="{ state: 'idle' }"  // 定义组件状态
x-show="condition"           // 条件显示
@click="state = 'active'"    // 事件绑定
x-text="variable"            // 动态文本
:class="{ 'active': isActive }" // 动态类名
x-transition                 // 过渡动画
```

## 配置

通过环境变量配置 Few-Shot 示例：

```bash
# 启用 Few-Shot 示例（可选）
ENABLE_FEW_SHOT=true
```

## 使用示例

```go
import "multimodal-llm-frontend-generator/pkg/prompt"

// System Prompt（HTML 模式）
systemPrompt := prompt.BuildSystemPromptHTML()

// 多图分析 Prompt
multiPrompt := prompt.BuildMultiImagePromptHTML(3)

// 双图差异分析
diffPrompt := prompt.BuildDiffAnalysisPromptHTML()

// 对话修改
chatPrompt := prompt.BuildChatModifyPromptHTML(currentCode, userMessage)
```

## 测试

```bash
go test ./pkg/prompt/... -v
```

## 更新历史

- **2026-01-15**: 切换到 HTML + Alpine.js 方案，添加五步分析框架
- **2025-xx-xx**: 初始版本，使用 React + Sandpack

## 相关文档

- [Service 层文档](../../internal/service/CLAUDE.md) - PromptService 实现
- [前端 HtmlPreview 组件](../../../frontend/src/components/preview/HtmlPreview.tsx)
