# code-generation Specification Delta

## MODIFIED Requirements

### Requirement: Prompt 构建

系统 **MUST** 根据用户输入和配置的 Prompt 策略构建符合 LLM 输入格式的多模态 Prompt。

#### Scenario: 构建单图 Prompt

- **GIVEN** 用户选择了一张图片
- **WHEN** PromptService 构建生成 Prompt
- **THEN** Prompt 包含系统消息（角色设定）
- **AND** Prompt 包含用户消息（图片 + 生成指令）
- **AND** 图片以 Base64 data URL 格式嵌入

#### Scenario: 构建多图差异分析 Prompt

- **GIVEN** 用户选择了两张或更多图片
- **WHEN** PromptService 构建差异分析 Prompt
- **THEN** Prompt 包含所有图片
- **AND** Prompt 包含结构化的差异分析指令
- **AND** 指令引导模型按五步分析框架执行
- **AND** 指令包含交互类型分类指南

#### Scenario: 使用 V1 策略构建 Prompt

- **GIVEN** 配置 `PROMPT_VERSION=v1` 或未配置
- **WHEN** PromptService 构建 Prompt
- **THEN** 使用现有的 Prompt 模板
- **AND** 保持向后兼容

#### Scenario: 使用 V2 策略构建 Prompt

- **GIVEN** 配置 `PROMPT_VERSION=v2`
- **WHEN** PromptService 构建 Prompt
- **THEN** 使用优化版 Prompt 模板
- **AND** System Prompt 包含结构化输出格式要求
- **AND** User Prompt 包含五步分析框架

## ADDED Requirements

### Requirement: Prompt 版本配置

系统 **SHALL** 支持通过配置切换不同版本的 Prompt 策略。

#### Scenario: 默认使用 V1 策略

- **GIVEN** 未设置 `PROMPT_VERSION` 环境变量
- **WHEN** 系统初始化
- **THEN** 默认使用 V1 Prompt 策略

#### Scenario: 配置 V2 策略

- **GIVEN** 设置 `PROMPT_VERSION=v2` 环境变量
- **WHEN** 系统初始化
- **THEN** 使用 V2 优化版 Prompt 策略

#### Scenario: 无效配置回退

- **GIVEN** 设置了无效的 `PROMPT_VERSION` 值
- **WHEN** 系统初始化
- **THEN** 回退到 V1 策略
- **AND** 记录警告日志

### Requirement: 多图分析增强

系统 **SHALL** 增强多图分析能力，支持 2-5 张图片的序列分析。

#### Scenario: 三图及以上序列分析

- **GIVEN** 用户上传了 3-5 张 UI 设计稿图片
- **AND** session_id 有效
- **WHEN** 用户发送 POST /api/generate 请求
- **THEN** 系统依次分析相邻图片的差异
- **AND** 识别多步交互流程
- **AND** 生成包含多个状态变量的代码

#### Scenario: 图片数量超过限制

- **GIVEN** 用户上传了超过 5 张图片
- **AND** session_id 有效
- **WHEN** 用户发送 POST /api/generate 请求
- **THEN** 系统仅处理前 5 张图片
- **AND** SSE 流中包含提示信息说明限制
