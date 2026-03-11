# Change: Upgrade to Production Architecture

## Why

系统已完成多 Agent Pipeline 架构升级和 3 轮代码审查修复，代码质量达到较高水平，但存在 6 个阻碍产品化的结构性缺陷：

1. **数据全部存在内存中**（`memory_store.go:30` `sessions map[string]*Session`），服务重启全部丢失。当前 MemoryStore 实现了正确的 SessionStore 接口（`session_store.go:6-36`），但后端没有持久化方案。
2. **Agent 硬编码**（`generate_workflow.py:25-40` 直接 import 各 Agent 类），新增/修改 Agent 必须改 Python 代码、重新部署。
3. **Pipeline 状态用 TypedDict 松散传递**（`state.py:1-44`），Agent 间数据依赖靠字符串 key 约定，无类型安全保障。
4. **Pipeline 无断点续跑能力**（`generate_workflow.py:90-130`），第 4 步 CodeGenerator 失败需从第 1 步 LayoutAnalyzer 重跑，浪费 4 次 LLM 调用费用。
5. **生成的交互原型无法分享**——用户只能在平台内预览，无法把原型给老板/同事体验。这是**产品核心差异化特性的缺失**。
6. **无用户体系**——无法保存项目历史、无法多设备使用。

## What Changes

### Go Backend (`backend/`)

- **ADDED** `internal/service/postgres_store.go` — PostgresStore 实现 SessionStore 接口，使用 pgx 连接池，图片和历史拆为独立表（避免 JSONB 大字段更新）
- **ADDED** `internal/service/preview_service.go` — 分享快照 CRUD，nanoid 8 位短码生成
- **ADDED** `internal/handler/preview.go` — `POST /api/share`（创建分享）+ `GET /p/:code`（Preview Host 静态 HTML 服务）
- **ADDED** `internal/middleware/auth.go` — JWT 认证中间件
- **ADDED** `internal/handler/auth.go` — 注册/登录/GitHub OAuth2 回调
- **ADDED** `internal/handler/project.go` — 项目 CRUD
- **ADDED** `cmd/migrate/main.go` — 数据库迁移工具
- **ADDED** `migrations/*.sql` — 数据库 schema（sessions/session_images/session_history/shared_previews/users/projects）
- **MODIFIED** `internal/config/config.go:11-62` — 新增 `DatabaseURL`、`JWTSecret` 配置项
- **MODIFIED** `internal/app/app.go:53-56` — 根据 `DATABASE_URL` 环境变量选择 PostgresStore 或 MemoryStore
- **MODIFIED** `docker-compose.yml` — 添加 PostgreSQL 服务

### Python Agent Core (`agent-core/`)

- **ADDED** `registry/agent_registry.py` — AgentRegistry，启动时从 YAML 加载 Agent 配置，提供 `get(name)`、`list_by_tag(tag)`、`get_pipeline_order()` 方法
- **ADDED** `agents/configs/*.yaml` — 6 个 Agent YAML 配置文件（layout_analyzer、component_detector、interaction_infer、code_generator、code_validator、chat_agent）
- **ADDED** `checkpoint/manager.py` — CheckpointManager，通过 Go 后端 API 保存/恢复 DesignState 到 PostgreSQL
- **MODIFIED** `graph/state.py` — PipelineState (TypedDict 44行) 升级为 DesignState (Pydantic ~80行)，增加 `completed_agents`、`checkpoints`、`errors` 字段，增加 `mark_agent_completed()`、`get_next_agent()`、`set_agent_output()` 方法
- **MODIFIED** `graph/generate_workflow.py` — GenerateWorkflow 重构：从 AgentRegistry 加载 Agent，使用 DesignState 管理状态，支持 `resume=True` 参数断点续跑
- **MODIFIED** `app/routes/generate.py:95-136` — `/api/v1/generate` 新增可选 `resume` 请求参数

### Frontend (`frontend/`)

- **ADDED** `src/components/preview/ShareButton.tsx` — 分享按钮 + 链接复制弹窗
- **ADDED** `src/app/dashboard/page.tsx` — 项目列表页（Phase D）
- **ADDED** `src/app/login/page.tsx` — 登录/注册页（Phase D）
- **MODIFIED** `src/lib/api/client.ts` — 新增 `api.share(sessionId)` 方法
- **MODIFIED** `src/components/preview/PreviewPanel.tsx` — 在操作栏集成 ShareButton
- **MODIFIED** `src/components/upload/UnifiedInput.tsx` — 添加剪贴板粘贴截图支持（Phase E）

## Impact

- Affected specs: `agent-pipeline`（Agent 配置和编排方式变更）、`session-management`（存储后端变更）
- New specs: `preview-sharing`（分享功能）
- Affected code: 见上方详细列表
- **BREAKING**: `SessionStore` 实现替换（接口不变，MemoryStore 仍保留用于测试和开发模式）
- **BREAKING**: `PipelineState` TypedDict → `DesignState` Pydantic（`generate_workflow.py` 内部使用，不影响外部 API）
- 新增外部依赖:
  - Go: `github.com/jackc/pgx/v5`（PostgreSQL 驱动）、`github.com/matoous/go-nanoid/v2`（短链接）、`github.com/golang-jwt/jwt/v5`（JWT）
  - Python: `pyyaml`（YAML 解析）、`httpx`（已有，Checkpoint 跨服务调用）
  - 基础设施: PostgreSQL 16+
