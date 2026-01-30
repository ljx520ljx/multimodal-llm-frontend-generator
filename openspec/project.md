# Project Context

## Purpose

基于多模态LLM的**交互原型验证平台**。

**核心目标**：利用多模态大语言模型（如 GPT-4o、Gemini 3 Pro）的视觉理解能力，将连续的 UI 设计稿序列自动转化为**可交互的前端原型**，让产品/设计师快速体验和验证交互流程是否合理。

**核心定位**：交互体验优先，代码生成为辅。

**目标用户**：
- **主要用户**：产品经理、UI 设计师 —— 体验和验证交互流程
- **次要用户**：前端开发者 —— 查看和修改生成的代码

**解决的问题**：
- 传统"设计稿到代码"转换效率低下
- 现有工具仅能还原单张静态页面，无法理解页面间的交互逻辑
- 赋能非技术人员（如产品经理）独立验证交互原型是否符合预期

## Tech Stack

### Frontend
- React 18+
- Next.js 14+
- TypeScript
- Tailwind CSS
- Monaco Editor（代码编辑器）
- Sandpack / CodeSandbox（实时代码预览沙箱）
- React-Beautiful-DND（拖拽排序）

### Backend
- Go (Golang)
- Gin / Echo（Web 框架）
- Goroutine（并发处理）

### AI Layer
- 多模态 LLM API（GPT-4o / Gemini 1.5 Pro / Claude）
- Prompt Engineering（CoT、Few-Shot、Role Play）

## Project Conventions

### Code Style
- 前端使用 TypeScript 严格模式
- 使用 ESLint + Prettier 进行代码格式化
- React 组件使用函数式组件 + Hooks
- 后端 Go 代码遵循 Go 官方代码规范

### Architecture Patterns
- 前后端分离的 B/S 架构
- 前端：组件化开发，状态管理使用 React Context 或 Zustand
- 后端：分层架构（Handler -> Service -> Repository）
- AI 调用：封装为独立的 Gateway 模块

### Testing Strategy
- 前端：Jest + React Testing Library
- 后端：Go 标准 testing 包
- 代码生成准确率测试：建立评估集对比不同 Prompt 策略

### Git Workflow
- main 分支保持稳定
- feature/* 分支用于功能开发
- 提交信息遵循 Conventional Commits 规范

## Domain Context

### 核心模块

1. **多模态交互意图解析模块**
   - 图像预处理（压缩、裁剪、Base64 编码）
   - 视觉差异分析（对比前后两帧推断交互事件）
   - 输出标准化 JSON Schema 中间表示

2. **动态代码生成引擎**
   - 组件化代码生成（React/Vue + Tailwind CSS）
   - 状态逻辑注入（useState、onClick 等）
   - 代码后处理（Prettier 格式化 + 语法校验）

3. **交互原型验证平台**
   - **简洁布局**：交互面板（上传+对话）+ 预览面板
   - 多图拖拽上传与排序
   - HTML + Alpine.js 实时交互预览（核心）
   - 对话式代码修改

4. **自然语言迭代修正模块**
   - 多轮对话交互
   - 增量代码更新

### Prompt 策略
- **Role Play**：设定模型为"资深前端架构师"
- **CoT (Chain of Thought)**：先描述视觉变化，再生成代码
- **Few-Shot Learning**：植入高质量"UI图对-代码"范例

## Important Constraints

### 性能指标
- 代码生成响应时间 ≤ 60秒
- 编译成功率 ≥ 90%
- 视觉还原度 ≥ 80%
- 交互逻辑正确率 ≥ 85%
- 支持 50+ 并发用户

### 限制条件
- 依赖外部 LLM API，需处理 API 限流和错误重试
- 高分辨率图片需压缩以适配 API 输入限制
- 生成的代码需通过后处理确保可编译性

## External Dependencies

### LLM APIs
- OpenAI GPT-4o API
- Google Gemini 1.5 Pro API
- Anthropic Claude API

### 前端沙箱
- Sandpack (CodeSandbox)

### 代码处理
- Prettier（代码格式化）
- ESLint（语法校验）
