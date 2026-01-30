# prompt-engineering Specification Delta

## ADDED Requirements

### Requirement: Prompt 策略接口

系统 **SHALL** 提供可切换的 Prompt 策略接口，支持不同版本的 Prompt 模板。

#### Scenario: V1 策略（默认）

- **WHEN** 配置 `PROMPT_VERSION=v1` 或未配置
- **THEN** 系统使用现有的 Prompt 模板
- **AND** 生成行为与当前版本一致

#### Scenario: V2 策略（优化版）

- **WHEN** 配置 `PROMPT_VERSION=v2`
- **THEN** 系统使用优化版 Prompt 模板
- **AND** 包含结构化五步分析框架
- **AND** 包含交互类型分类指南

#### Scenario: 策略运行时切换

- **WHEN** 管理员修改环境变量 `PROMPT_VERSION`
- **THEN** 下一次 API 请求使用新策略
- **AND** 无需重启服务

### Requirement: 结构化分析框架

V2 Prompt **MUST** 引导 LLM 按以下五步执行多图分析：

#### Scenario: Step 1 布局识别

- **WHEN** LLM 接收到多图分析请求
- **THEN** 首先识别每张图的整体布局结构
- **AND** 在 `<thinking>` 中输出布局描述

#### Scenario: Step 2 组件识别

- **WHEN** 布局识别完成后
- **THEN** 识别主要 UI 组件并命名
- **AND** 使用语义化命名（Header, Sidebar, Card 等）

#### Scenario: Step 3 差异检测

- **WHEN** 组件识别完成后
- **THEN** 对比相邻图片的视觉变化
- **AND** 列出具体的变化项（显示/隐藏、样式、内容）

#### Scenario: Step 4 交互推理

- **WHEN** 差异检测完成后
- **THEN** 根据差异推断交互类型
- **AND** 识别触发元素和事件类型

#### Scenario: Step 5 代码生成

- **WHEN** 交互推理完成后
- **THEN** 生成包含状态管理的 React 代码
- **AND** 代码包含所有识别到的状态和事件处理器

### Requirement: 交互类型分类

系统 **SHALL** 支持识别以下 6 种交互类型：

#### Scenario: Toggle 交互

- **WHEN** 检测到元素显示/隐藏切换
- **THEN** 识别为 Toggle 类型
- **AND** 生成 `useState<boolean>` 状态
- **AND** 绑定 `onClick` 事件处理器

#### Scenario: Expand 交互

- **WHEN** 检测到区域展开/收起
- **THEN** 识别为 Expand 类型
- **AND** 生成带过渡动画的展开/收起代码

#### Scenario: TabSwitch 交互

- **WHEN** 检测到 Tab 或标签页高亮切换
- **THEN** 识别为 TabSwitch 类型
- **AND** 生成 `useState<string>` 状态
- **AND** 支持多个 Tab 之间切换

#### Scenario: FormState 交互

- **WHEN** 检测到表单输入状态变化
- **THEN** 识别为 FormState 类型
- **AND** 生成表单状态对象
- **AND** 绑定 `onChange` 事件处理器

#### Scenario: Hover 交互

- **WHEN** 检测到悬停样式变化
- **THEN** 识别为 Hover 类型
- **AND** 优先使用 Tailwind `hover:` 类
- **AND** 或使用 `onMouseEnter/onMouseLeave`

#### Scenario: Navigation 交互

- **WHEN** 检测到页面/视图跳转
- **THEN** 识别为 Navigation 类型
- **AND** 生成视图状态管理代码

### Requirement: Few-Shot 示例

V2 Prompt **SHALL** 支持可选的高质量示例来引导生成：

#### Scenario: 示例格式

- **WHEN** 启用 Few-Shot 模式
- **THEN** 示例包含输入描述、分析过程和输出代码
- **AND** 示例展示期望的分析深度和代码风格

#### Scenario: 示例数量控制

- **WHEN** 添加 Few-Shot 示例
- **THEN** 最多包含 2 个示例
- **AND** 总 token 数控制在合理范围内
