# multi-image-analysis Specification Delta

## ADDED Requirements

### Requirement: 多图序列处理

系统 **SHALL** 支持处理 2-5 张 UI 设计稿序列，分析它们之间的状态流转。

#### Scenario: 两图对比分析

- **GIVEN** 用户上传了 2 张 UI 设计稿
- **WHEN** 用户请求生成代码
- **THEN** 系统分析两张图之间的视觉差异
- **AND** 推断单一交互事件触发的状态变化
- **AND** 生成对应的状态切换代码

#### Scenario: 多图序列分析

- **GIVEN** 用户上传了 3-5 张 UI 设计稿
- **WHEN** 用户请求生成代码
- **THEN** 系统依次分析相邻图片的差异
- **AND** 识别多步交互流程
- **AND** 生成包含多个状态变量的代码

#### Scenario: 图片数量超限

- **GIVEN** 用户上传了超过 5 张图片
- **WHEN** 用户请求生成代码
- **THEN** 系统仅分析前 5 张图片
- **AND** 返回提示信息说明图片数量限制

### Requirement: 视觉差异检测

系统 **MUST** 能够检测多图之间的视觉差异。

#### Scenario: 元素可见性变化

- **GIVEN** 图片 A 显示某元素，图片 B 隐藏该元素
- **WHEN** 系统分析差异
- **THEN** 检测到元素可见性变化
- **AND** 识别变化的具体元素

#### Scenario: 样式变化检测

- **GIVEN** 图片 A 和图片 B 中同一元素样式不同
- **WHEN** 系统分析差异
- **THEN** 检测到样式变化（颜色、大小、位置等）
- **AND** 描述具体的样式差异

#### Scenario: 内容变化检测

- **GIVEN** 图片 A 和图片 B 中文本或图标内容不同
- **WHEN** 系统分析差异
- **THEN** 检测到内容变化
- **AND** 描述具体的内容差异

### Requirement: 交互意图推理

系统 **SHALL** 基于视觉差异推断用户交互意图。

#### Scenario: 点击触发推理

- **GIVEN** 检测到按钮/链接样式变化或元素显示/隐藏
- **WHEN** 系统推理交互意图
- **THEN** 推断为 onClick 事件触发
- **AND** 识别触发元素的位置和特征

#### Scenario: 输入触发推理

- **GIVEN** 检测到表单字段内容变化
- **WHEN** 系统推理交互意图
- **THEN** 推断为 onChange 事件触发
- **AND** 识别相关的输入元素

#### Scenario: 悬停触发推理

- **GIVEN** 检测到轻微的样式变化（如颜色加深、阴影出现）
- **WHEN** 系统推理交互意图
- **THEN** 推断为 hover 状态变化
- **AND** 优先使用 CSS hover 而非 JS 事件

### Requirement: 状态机代码生成

系统 **SHALL** 生成包含完整状态管理的 React 代码。

#### Scenario: 单状态生成

- **GIVEN** 识别到单一交互（如 Toggle）
- **WHEN** 系统生成代码
- **THEN** 生成单个 useState 声明
- **AND** 生成对应的事件处理器
- **AND** 条件渲染实现状态切换

#### Scenario: 多状态生成

- **GIVEN** 识别到多个交互（如 Tab + Expand）
- **WHEN** 系统生成代码
- **THEN** 生成多个 useState 声明
- **AND** 每个状态有对应的事件处理器
- **AND** 状态之间相互独立

#### Scenario: 关联状态生成

- **GIVEN** 识别到关联的状态变化（如 Tab 切换影响内容显示）
- **WHEN** 系统生成代码
- **THEN** 正确处理状态之间的依赖关系
- **AND** 使用合适的状态结构（对象或多个变量）

### Requirement: 分析过程可见性

系统 **SHALL** 输出可读的分析过程，帮助用户理解生成逻辑。

#### Scenario: thinking 标签输出

- **WHEN** 系统进行多图分析
- **THEN** 在 `<thinking>` 标签中输出分析过程
- **AND** 包含布局识别、组件识别、差异检测、交互推理的描述

#### Scenario: 流式输出分析过程

- **WHEN** 系统以 SSE 流式方式返回
- **THEN** 分析过程（thinking）先于代码输出
- **AND** 用户可以实时看到分析进展
