# Proposal: Phase 2 后端核心服务

## Summary

实现后端核心服务层，包括图片处理服务、Prompt 构建服务、代码生成服务，以及对应的 HTTP Handler 层。这是连接前端 Web IDE 和 LLM Gateway 的核心业务逻辑层。

## Motivation

Phase 0 已完成基础设施搭建，Phase 1 已完成 LLM Gateway 开发。现在需要实现：

1. **图片上传与处理** - 接收用户上传的 UI 设计稿，进行压缩、格式转换、Base64 编码
2. **Prompt 构建** - 将图片和用户意图构建为符合 LLM 输入格式的多模态 Prompt
3. **代码生成** - 调用 LLM Gateway 生成代码，并进行流式输出
4. **自然语言迭代** - 支持用户通过自然语言修改生成的代码

## Scope

### In Scope

- **Service 层**
  - ImageService: 图片压缩、格式转换、Base64 编码、临时存储
  - PromptService: 构建系统 Prompt、用户 Prompt、管理对话上下文
  - GenerateService: 协调图片处理、Prompt 构建、LLM 调用的主流程

- **Handler 层**
  - UploadHandler: POST /api/upload - 多图上传
  - GenerateHandler: POST /api/generate - 代码生成 (SSE)
  - ChatHandler: POST /api/chat - 自然语言修改 (SSE)

- **Prompt 模板**
  - SystemPrompt: 角色设定和输出规范
  - GeneratePrompt: 代码生成指令
  - DiffAnalysisPrompt: 视觉差异分析

- **会话管理**
  - 基于内存的会话存储 (Session Store)
  - 会话数据结构: 图片、生成的代码、对话历史

### Out of Scope

- 持久化存储 (Redis/PostgreSQL) - 留待后续阶段
- 用户认证与授权
- 代码后处理 (Prettier 格式化) - 可在后续迭代添加
- 前端集成

## Design

见 [design.md](./design.md)

## Tasks

见 [tasks.md](./tasks.md)

## Risks & Mitigations

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 大图片上传导致内存溢出 | 服务崩溃 | 限制单张图片最大 10MB，总上传 50MB |
| 会话内存占用过大 | 内存不足 | 设置会话过期时间 (30 分钟)，定期清理 |
| LLM 响应超时 | 用户等待过久 | 可配置超时 (默认 5 分钟)，支持取消请求 |
| 多并发请求竞争 | 数据不一致 | 使用 sync.RWMutex 保护共享状态 |

## Success Criteria

1. 图片上传接口可正常接收并处理多张图片
2. 代码生成接口可流式返回 LLM 生成的代码
3. 自然语言修改接口可基于上下文迭代修改代码
4. 所有接口有完整的单元测试覆盖
5. 集成测试可端到端验证核心流程
