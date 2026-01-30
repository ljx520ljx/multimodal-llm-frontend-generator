# Design: 多图交互分析架构与 Prompt 工程

## Context

### 背景

本项目是毕业设计，目标是将连续 UI 设计稿自动转化为可交互的前端原型。当前 Phase 5 MVP 已完成，需要优化多图交互分析能力（Phase 6）。

### 核心约束

1. **成本敏感**：依赖外部 LLM API，调用成本较高
2. **模型能力有限**：不同 Provider（OpenAI/DeepSeek/Doubao）能力差异大
3. **生成质量要求高**：需要达到 85%+ 的交互逻辑正确率

### 设计原则

参考 Claude Code 架构思想：
- **Prompt 精准化**：用精确的指令弥补模型能力差距
- **任务分解**：将复杂任务拆解为明确的分析步骤
- **结构化输出**：引导模型输出结构化的中间结果

## Goals / Non-Goals

### Goals

1. 设计结构化的多图分析 Prompt 框架
2. 定义交互类型分类体系
3. 优化 Prompt 模板提升生成质量
4. 保持架构简单，易于迭代

### Non-Goals

1. 不引入复杂的多 Agent 系统
2. 不增加额外的 LLM 调用次数（避免成本增加）
3. 不改变现有的 API 接口

## Decisions

### Decision 1: 采用单次调用 + 结构化 Prompt 策略

**选择**：优化单次 LLM 调用的 Prompt，而非多次调用

**理由**：
- 多次调用会增加成本和延迟
- 精心设计的 Prompt 可以引导模型完成复杂推理
- 保持架构简单，易于调试和维护

**备选优化方案（当前阶段不实现，效果不理想时考虑）**：
- 多 Agent 协作：分析 Agent + 生成 Agent 分工，提升专注度
- 两阶段分析：先输出结构化分析结果，再基于分析生成代码

### Decision 2: 五步分析框架

将多图分析拆解为 5 个明确步骤，在 Prompt 中引导模型依次执行：

```
Step 1: 布局识别 → 识别每张图的 UI 结构
Step 2: 组件分析 → 识别组件层级和命名
Step 3: 差异检测 → 对比相邻图片的变化
Step 4: 意图推理 → 推断交互类型和触发条件
Step 5: 代码生成 → 生成状态机代码
```

**理由**：
- 分步引导降低模型推理难度
- 每一步的输出可以作为下一步的输入
- 便于定位生成错误的阶段

### Decision 3: 交互类型分类体系

定义 6 种核心交互类型：

| 类型 | 识别特征 | 代码模式 |
|------|----------|----------|
| Toggle | 元素显示/隐藏 | `useState<boolean>` |
| Expand | 区域展开/收起 | `useState<boolean>` + height transition |
| TabSwitch | Tab 高亮切换 | `useState<string>` |
| FormState | 表单状态变化 | `useState<object>` |
| Hover | 悬停样式变化 | `hover:` 或 `onMouseEnter/Leave` |
| Navigation | 页面/视图跳转 | `useState<string>` 或 Router |

**理由**：
- 覆盖 90%+ 的常见 UI 交互场景
- 每种类型有明确的代码生成模式
- 便于在 Prompt 中给出针对性示例

### Decision 4: Prompt 模板结构

采用三层 Prompt 结构：

```
┌─────────────────────────────────────┐
│  System Prompt                      │
│  - 角色设定：资深前端架构师          │
│  - 技术约束：React + Tailwind       │
│  - 输出格式：<thinking> + ```code   │
└─────────────────────────────────────┘
          ↓
┌─────────────────────────────────────┐
│  Analysis Prompt (多图分析指令)      │
│  - 五步分析框架                      │
│  - 交互类型分类指南                  │
│  - 输出结构要求                      │
└─────────────────────────────────────┘
          ↓
┌─────────────────────────────────────┐
│  Few-Shot Examples (可选)           │
│  - 1-2 个高质量示例                  │
│  - 展示期望的分析和输出格式          │
└─────────────────────────────────────┘
```

### Decision 5: 不改变现有 API 接口

保持现有的 `/api/generate` 接口不变，优化仅限于内部 Prompt 构建逻辑。

**理由**：
- 前端代码无需修改
- 向后兼容
- 减少测试范围

## Technical Design

### Prompt 模板设计

#### System Prompt (优化版)

```go
const SystemPromptV2 = `你是一位资深前端架构师，专精于 UI 设计稿转代码。

## 核心能力
- 精准识别 UI 布局结构和组件层级
- 分析多图差异，推断交互逻辑
- 生成高质量的 React + Tailwind 代码

## 技术约束
- 使用 React 18+ 函数式组件
- 使用 Tailwind CSS（不使用 @apply）
- 代码必须在 Sandpack 沙箱中可直接运行
- 不使用外部图片 URL，使用占位符或内联 SVG
- 所有文本使用中文

## 输出格式
1. 先用 <thinking> 标签描述分析过程
2. 然后输出完整的 React 代码块
3. 代码必须包含所有必要的 import 语句
4. 导出默认组件名为 App`
```

#### Multi-Image Analysis Prompt

```go
const MultiImageAnalysisPromptV2 = `## 多图交互分析任务

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
| 类型 | 特征 |
|------|------|
| Toggle | 元素显示/隐藏切换 |
| Expand | 区域展开/收起 |
| TabSwitch | Tab/标签页切换 |
| FormState | 表单输入状态变化 |
| Hover | 悬停样式变化 |
| Navigation | 页面/视图跳转 |

**Step 5: 代码生成**
生成包含以下内容的 React 代码：
- useState 管理所有识别到的状态
- onClick/onChange 处理器绑定到触发元素
- 条件渲染实现所有图片中展示的状态

### 输出要求
- <thinking> 中完成 Step 1-4 的分析
- 代码块中输出 Step 5 的完整代码
- 确保代码能实现所有图片展示的状态切换`
```

### 代码结构

```
backend/
├── pkg/prompt/
│   ├── templates.go       # 基础模板（保留）
│   ├── templates_v2.go    # 优化版模板（新增）
│   └── interaction_types.go  # 交互类型定义（新增）
│
├── internal/service/
│   ├── prompt_service.go  # 增加策略选择
│   └── prompt_strategy.go # Prompt 策略接口（新增）
```

### 新增类型定义

```go
// pkg/prompt/interaction_types.go

// InteractionType 定义支持的交互类型
type InteractionType string

const (
    InteractionToggle     InteractionType = "toggle"
    InteractionExpand     InteractionType = "expand"
    InteractionTabSwitch  InteractionType = "tab_switch"
    InteractionFormState  InteractionType = "form_state"
    InteractionHover      InteractionType = "hover"
    InteractionNavigation InteractionType = "navigation"
)

// InteractionPattern 描述交互模式的代码生成规则
type InteractionPattern struct {
    Type         InteractionType
    StateType    string // "boolean" | "string" | "object"
    DefaultValue string
    EventHandler string // "onClick" | "onChange" | "onMouseEnter"
}

var InteractionPatterns = map[InteractionType]InteractionPattern{
    InteractionToggle: {
        Type:         InteractionToggle,
        StateType:    "boolean",
        DefaultValue: "false",
        EventHandler: "onClick",
    },
    // ... 其他类型
}
```

### Prompt Strategy 接口

```go
// internal/service/prompt_strategy.go

// PromptStrategy 定义 Prompt 构建策略
type PromptStrategy interface {
    // BuildSystemPrompt 构建系统 Prompt
    BuildSystemPrompt(framework string) string

    // BuildUserPrompt 构建用户 Prompt
    BuildUserPrompt(images []ImageData, framework string) []types.Message
}

// V1Strategy 现有策略（保留向后兼容）
type V1Strategy struct{}

// V2Strategy 优化版策略
type V2Strategy struct {
    includeFewShot bool
}
```

## Risks / Trade-offs

### Risk 1: Prompt 过长导致性能下降

**风险**：结构化 Prompt 较长，可能增加 token 消耗和响应时间

**缓解措施**：
- 控制 System Prompt 在 500 tokens 以内
- Few-Shot 示例仅在必要时添加
- 监控 token 使用量

### Risk 2: 不同模型适配问题

**风险**：优化的 Prompt 可能在 DeepSeek/Doubao 上效果不一致

**缓解措施**：
- 针对不同 Provider 调整 Prompt 细节
- 建立评估集测试各模型效果
- 保留 V1 策略作为回退方案

### Risk 3: 复杂交互场景支持不足

**风险**：6 种交互类型可能无法覆盖所有场景

**缓解措施**：
- 先覆盖 80% 常见场景
- 收集失败案例，迭代补充类型
- 提供 "其他" 类型的通用处理

## Migration Plan

### Phase 1: 添加新 Prompt 模板（不影响现有功能）
- 新增 `templates_v2.go`
- 新增 `PromptStrategy` 接口
- 默认使用 V1 策略

### Phase 2: 小范围验证
- 通过环境变量切换到 V2 策略
- 收集测试数据，对比生成质量

### Phase 3: 全量切换
- 默认使用 V2 策略
- 监控指标，如有问题可快速回滚

### Rollback Plan
- 保留 V1 策略代码
- 通过配置快速切换回 V1

## Open Questions

1. **Few-Shot 示例数量**：1 个还是 2 个示例效果更好？需要 A/B 测试验证
2. **交互类型扩展**：是否需要支持 "动画过渡" 类型？
3. **多语言支持**：是否需要支持英文 UI 的分析？
