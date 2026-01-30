<!-- OPENSPEC:START -->
# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:
- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:
- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->

---

# L1 - 项目级文档 | Multimodal LLM Frontend Generator

> DocOps 层级: **L1 (Project Level)**
> 文档即代码，代码是文档的实现

## 项目概述

**基于多模态LLM的交互原型验证平台**

将连续 UI 设计稿序列自动转化为可交互的前端原型，让产品/设计师快速体验和验证交互流程。

## 核心价值主张

```
设计稿序列 → 视觉差异分析 → 交互意图推理 → 可交互原型
```

- **交互体验优先**：重点是让用户体验交互流程，而非查看代码
- **赋能非技术人员**：产品经理、设计师可独立验证交互原型是否符合预期
- **对话式交互**：支持自然语言描述修改需求，AI 增量更新代码
- 突破"单图生成"限制，理解页面间的状态流转

## 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                  交互原型验证平台 (Frontend)                  │
│  ┌──────────────────────┬──────────────────────────────────┐ │
│  │   交互面板 (320px)    │         预览面板 (flex-1)        │ │
│  │  ┌────────────────┐  │  ┌────────────────────────────┐  │ │
│  │  │ 图片上传       │  │  │                            │  │ │
│  │  │ 对话输入       │  │  │    HTML/Alpine.js 预览     │  │ │
│  │  │ 对话历史       │  │  │    (iframe 沙箱)           │  │ │
│  │  └────────────────┘  │  └────────────────────────────┘  │ │
│  └──────────────────────┴──────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Backend (Go + Gin)                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐   │
│  │ Image        │  │ Prompt       │  │ Code             │   │
│  │ Processor    │  │ Engine       │  │ Post-Processor   │   │
│  └──────┬───────┘  └──────┬───────┘  └────────┬─────────┘   │
└─────────┼─────────────────┼───────────────────┼─────────────┘
          │                 │                   │
          ▼                 ▼                   ▼
┌─────────────────────────────────────────────────────────────┐
│                    AI Layer (LLM Gateway)                    │
│         GPT-4o  |  Gemini 3 Pro  |  Claude                  │
│                                                              │
│    Prompt Strategies: CoT | Few-Shot | Role Play            │
└─────────────────────────────────────────────────────────────┘
```

## 目标用户

| 用户类型 | 使用场景 |
|----------|----------|
| **产品经理** | 验证交互流程是否合理 |
| **UI 设计师** | 快速生成可交互原型 |
| **前端开发者** | 查看生成的代码、通过对话修改 |

## 技术栈速查

| 层级 | 技术选型 | 版本要求 |
|------|----------|----------|
| Frontend | Next.js + React + TypeScript | Next 14+, React 18+ |
| Styling | Tailwind CSS | 3.x |
| Editor | Monaco Editor | latest |
| Preview | Sandpack (CodeSandbox) | latest |
| Backend | Go + Gin | Go 1.21+ |
| AI | OpenAI / Google / Anthropic API | - |

## 目录结构

```
multimodal-llm-frontend-generator/
├── CLAUDE.md                 # L1 项目级文档 (本文件)
├── openspec/                 # OpenSpec 规范目录
│   ├── project.md           # 项目上下文
│   ├── specs/               # 规范文档
│   └── changes/             # 变更提案
├── .claude/
│   ├── skills/              # Agent Skills
│   └── commands/            # Claude Commands
├── frontend/                 # 前端代码 (Next.js)
│   ├── CLAUDE.md            # L2 前端模块文档
│   ├── src/
│   │   ├── components/      # L3 组件级文档
│   │   ├── features/        # L3 功能级文档
│   │   └── lib/             # L3 工具库文档
│   └── ...
├── backend/                  # 后端代码 (Go)
│   ├── CLAUDE.md            # L2 后端模块文档
│   ├── internal/
│   │   ├── handler/         # L3 处理器文档
│   │   ├── service/         # L3 服务层文档
│   │   └── gateway/         # L3 网关文档
│   └── ...
└── docs/                     # 额外文档
```

## DocOps 三级文档体系

| 级别 | 位置 | 职责 | 更新频率 |
|------|------|------|----------|
| **L1** | `/CLAUDE.md` | 项目愿景、架构、技术栈、开发规范 | 低 |
| **L2** | `/{module}/CLAUDE.md` | 模块设计、接口定义、依赖关系 | 中 |
| **L3** | `/{module}/{feature}/CLAUDE.md` | 实现细节、代码约定、测试策略 | 高 |

## 开发规范

### 代码风格
- **TypeScript**: 严格模式，禁止 `any`
- **React**: 函数式组件 + Hooks
- **Go**: 官方 gofmt 规范
- **命名**: 组件 PascalCase，函数/变量 camelCase

### Git 工作流
```bash
main          # 稳定版本
├── develop   # 开发分支
└── feature/* # 功能分支
```

### 提交规范
```
<type>(<scope>): <subject>

feat(frontend): add image upload component
fix(backend): resolve token expiration issue
docs(l2): update frontend module documentation
```

## Skills 使用指南

| Skill | 触发场景 | 文件位置 |
|-------|----------|----------|
| 项目理解 | 首次接触、需求分析 | `.claude/skills/project-understanding.md` |
| 质量保障 | 代码审查、提交前 | `.claude/skills/quality-assurance.md` |
| 代码演进 | 重构、技术债务 | `.claude/skills/code-evolution.md` |
| API测试 | 接口开发、集成测试 | `.claude/skills/api-integration-testing.md` |

## 快速开始

```bash
# 1. 安装依赖
cd frontend && npm install
cd backend && go mod download

# 2. 配置环境变量
cp .env.example .env
# 填写 LLM API Keys

# 3. 启动开发服务
npm run dev          # 前端 (localhost:3000)
go run cmd/server/main.go  # 后端 (localhost:8080)
```

## 关键性能指标

| 指标 | 目标值 | 测量方法 |
|------|--------|----------|
| 代码生成响应时间 | ≤ 60s | API 响应时间 |
| 编译成功率 | ≥ 90% | Sandpack 编译结果 |
| 视觉还原度 | ≥ 80% | DOM 结构对比 |
| 交互逻辑正确率 | ≥ 85% | 人工验证 |
| 并发支持 | 50+ | 压力测试 |

## 相关文档

- [OpenSpec 工作流程](./openspec/AGENTS.md)
- [项目上下文](./openspec/project.md)
- [前端模块文档](./frontend/CLAUDE.md) (L2)
- [后端模块文档](./backend/CLAUDE.md) (L2)
