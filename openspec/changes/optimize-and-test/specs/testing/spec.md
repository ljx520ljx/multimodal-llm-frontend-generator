# testing Specification

## Purpose

定义项目测试基础设施和覆盖率要求，确保代码质量和稳定性。

## ADDED Requirements

### Requirement: Frontend Unit Testing

前端 **SHALL** 配置单元测试框架并达到最低覆盖率要求。

#### Scenario: Vitest 配置

**Given** 前端项目使用 Next.js + React
**When** 运行 `npm run test`
**Then** Vitest 执行所有测试文件
**And** 生成覆盖率报告
**And** 测试通过率为 100%

#### Scenario: Store 测试覆盖

**Given** 项目使用 Zustand 管理状态
**When** 测试 Store 模块
**Then** `useProjectStore` 覆盖率 > 80%
**And** `useChatStore` 覆盖率 > 80%

#### Scenario: Hook 测试覆盖

**Given** 项目包含自定义 Hooks
**When** 测试 Hook 模块
**Then** `useGeneration` 覆盖率 > 70%
**And** `useChat` 覆盖率 > 70%

#### Scenario: 组件测试覆盖

**Given** 项目包含核心 UI 组件
**When** 测试组件模块
**Then** 核心组件（ImageDropzone, ImageList, ChatInput, GenerateButton）覆盖率 > 60%

### Requirement: Backend Unit Testing

后端 **SHALL** 达到最低测试覆盖率要求。

#### Scenario: 整体覆盖率要求

**Given** 后端使用 Go 标准测试框架
**When** 运行 `go test ./... -cover`
**Then** 整体覆盖率 > 60%

#### Scenario: 关键模块覆盖率

**Given** 后端包含多个模块
**When** 检查各模块覆盖率
**Then** `internal/handler` 覆盖率 > 75%
**And** `internal/service` 覆盖率 > 80%
**And** `internal/gateway` 覆盖率 > 70%
**And** `internal/config` 覆盖率 > 80%
**And** `internal/middleware` 覆盖率 > 70%

### Requirement: E2E Testing

系统 **SHALL** 配置 E2E 测试覆盖关键用户路径。

#### Scenario: E2E 框架配置

**Given** 项目需要端到端测试
**When** 运行 `npm run e2e`
**Then** Playwright 执行所有 E2E 测试
**And** 测试在 Chromium 浏览器运行
**And** 生成测试报告

#### Scenario: 上传流程 E2E

**Given** 用户访问应用首页
**When** 用户拖拽图片到上传区域
**Then** 图片显示在预览列表
**And** 用户可以重新排序图片
**And** 用户可以删除图片

#### Scenario: 生成流程 E2E

**Given** 用户已上传至少一张图片
**When** 用户点击"生成原型"按钮
**Then** 显示生成进度
**And** 流式输出代码内容
**And** 预览区域渲染生成的代码

#### Scenario: 聊天修正流程 E2E

**Given** 用户已生成代码
**When** 用户在聊天输入框输入修改指令并发送
**Then** 显示用户消息
**And** 显示 AI 响应（流式）
**And** 代码更新
**And** 预览刷新

### Requirement: Test Data Management

系统 **SHALL** 提供测试数据管理机制。

#### Scenario: 测试图片资源

**Given** E2E 测试需要测试图片
**When** 运行 E2E 测试
**Then** 使用 `e2e/fixtures/test-images/` 目录下的图片
**And** 测试图片包含不同尺寸和格式

#### Scenario: API Mock

**Given** 前端单元测试需要 Mock API
**When** 测试涉及 API 调用的组件
**Then** 使用 MSW 或手动 Mock 模拟 API 响应
