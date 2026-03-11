# Tasks: Production Architecture Upgrade

## Phase A: PostgreSQL 持久化（Week 1-2）

### A.1 数据库基础设施
- [ ] A.1.1 添加 `github.com/jackc/pgx/v5` 到 `backend/go.mod`
- [ ] A.1.2 添加 `github.com/golang-migrate/migrate/v4` 到 `backend/go.mod`（迁移工具）
- [ ] A.1.3 创建 `backend/migrations/001_create_sessions.sql`（sessions + session_images + session_history + shared_previews 四张表，含索引）
- [ ] A.1.4 创建 `backend/cmd/migrate/main.go`（迁移入口，支持 `up`/`down`/`status`）
- [ ] A.1.5 `docker-compose.yml` 添加 PostgreSQL 16 服务（端口 5432，卷映射 `pgdata`）

### A.2 PostgresStore 实现
- [ ] A.2.1 创建 `backend/internal/service/postgres_store.go`，定义 `PostgresStore struct { pool *pgxpool.Pool; historyLimit int }`
- [ ] A.2.2 实现 `Create(ctx)` — `INSERT INTO sessions DEFAULT VALUES RETURNING ...`
- [ ] A.2.3 实现 `Get(ctx, id)` — 查询 sessions + JOIN session_images (ORDER BY sort_order) + JOIN session_history (ORDER BY created_at DESC LIMIT N)，刷新 expires_at
- [ ] A.2.4 实现 `Update(ctx, session)` — `UPDATE sessions SET code=$1, framework=$2, updated_at=NOW()`
- [ ] A.2.5 实现 `Delete(ctx, id)` — `DELETE FROM sessions WHERE id=$1`（CASCADE 删除关联数据）
- [ ] A.2.6 实现 `AddImage(ctx, sid, image)` — `INSERT INTO session_images`
- [ ] A.2.7 实现 `GetImages(ctx, sid, imageIDs)` — `SELECT ... WHERE session_id=$1 AND id = ANY($2)`
- [ ] A.2.8 实现 `UpdateCode(ctx, sid, code)` — `UPDATE sessions SET code=$1`
- [ ] A.2.9 实现 `AddHistory(ctx, sid, entry)` — `INSERT INTO session_history` + 裁剪超出 historyLimit 的旧记录
- [ ] A.2.10 实现 `GetHistory(ctx, sid, limit)` — `SELECT ... ORDER BY created_at DESC LIMIT $2`（结果逆序返回）
- [ ] A.2.11 实现 `Close()` — `pool.Close()`

### A.3 集成与切换
- [ ] A.3.1 `backend/internal/config/config.go` 新增 `DatabaseURL string` 字段，`viper.SetDefault("DATABASE_URL", "")`
- [ ] A.3.2 `backend/internal/app/app.go:53-56` 修改 `initServices()`：根据 `DATABASE_URL` 是否为空选择 PostgresStore 或 MemoryStore
- [ ] A.3.3 `backend/internal/app/app.go` 修改 `Close()`：支持关闭 PostgresStore（需要类型断言或统一接口）
- [ ] A.3.4 创建 `backend/internal/service/postgres_store_test.go`，使用 testcontainers-go 编写集成测试（覆盖 Create → AddImage → Get → UpdateCode → AddHistory → GetHistory → Delete 完整流程）

### A.4 验证
- [ ] A.4.1 `go build ./...` 编译通过
- [ ] A.4.2 `go test ./...` 所有测试通过（包括 PostgresStore 集成测试）
- [ ] A.4.3 端到端验证：`docker-compose up` → upload → generate → 重启 backend → GET session 数据仍在
- [ ] A.4.4 验证不设置 `DATABASE_URL` 时自动 fallback 到 MemoryStore

## Phase B: Agent 插件化 + DesignState + Checkpoint（Week 3-5）

### B.1 Agent YAML Schema
- [ ] B.1.1 创建 `agent-core/registry/__init__.py`
- [ ] B.1.2 创建 `agent-core/registry/agent_registry.py`，定义 `AgentConfig(BaseModel)` 含 name/version/description/tags/llm/input/output/prompt/retry/dependencies 字段
- [ ] B.1.3 实现 `AgentRegistry.__init__(configs_dir)` — 扫描目录加载所有 `.yaml` 文件
- [ ] B.1.4 实现 `AgentRegistry.get(name)` / `list_by_tag(tag)` / `list_all()`
- [ ] B.1.5 实现 `AgentRegistry.get_pipeline_order()` — 基于 `dependencies` 字段的 Kahn 拓扑排序，检测循环依赖
- [ ] B.1.6 编写 `agent-core/tests/test_agent_registry.py`（测试加载、查询、排序、循环依赖检测）

### B.2 Agent YAML 配置迁移
- [ ] B.2.1 创建 `agent-core/agents/configs/` 目录
- [ ] B.2.2 创建 `layout_analyzer.yaml`（从 `agents/layout_analyzer.py` 提取 name/description/prompt，设 temperature=0.2，dependencies=[]）
- [ ] B.2.3 创建 `component_detector.yaml`（temperature=0.2，dependencies=[layout_analyzer]）
- [ ] B.2.4 创建 `interaction_infer.yaml`（temperature=0.5，dependencies=[layout_analyzer, component_detector]）
- [ ] B.2.5 创建 `code_generator.yaml`（temperature=0.3，max_tokens=8192，dependencies=[layout_analyzer, component_detector, interaction_infer]）
- [ ] B.2.6 创建 `code_validator.yaml`（非 LLM Agent，tags=["pipeline","validation"]，dependencies=[code_generator]）
- [ ] B.2.7 创建 `chat_agent.yaml`（temperature=0.7，tags=["chat"]，dependencies=[]）

### B.3 DesignState 实现
- [ ] B.3.1 重写 `agent-core/graph/state.py`：PipelineState (TypedDict) → DesignState (Pydantic BaseModel)
  - 保留所有现有字段：session_id, images, layout_info, component_info, interaction_info, generated_code, validation_errors, retry_count, final_html, success, error
  - 新增字段：current_agent, completed_agents, checkpoints (dict[str,str]), errors (list[AgentError])
  - 新增方法：mark_agent_completed(name), get_next_agent(order), set_agent_output(name, output)
- [ ] B.3.2 新增 `AgentError(BaseModel)` 类（agent_name, error, recoverable, timestamp）
- [ ] B.3.3 编写 `agent-core/tests/test_design_state.py`（测试 mark_completed、get_next_agent、set_agent_output、序列化/反序列化）

### B.4 CheckpointManager 实现
- [ ] B.4.1 创建 `agent-core/checkpoint/__init__.py`
- [ ] B.4.2 创建 `agent-core/checkpoint/manager.py`，实现 `save(state)` / `load(session_id)` / `close()`
  - save: PUT `/api/internal/checkpoint/{session_id}`，序列化 DesignState（排除 images），保存到 Go 后端
  - load: GET `/api/internal/checkpoint/{session_id}`，反序列化为 DesignState
- [ ] B.4.3 Go 后端新增内部 API：
  - `PUT /api/internal/checkpoint/:session_id` — `UPDATE sessions SET design_state = $1 WHERE id = $2`
  - `GET /api/internal/checkpoint/:session_id` — `SELECT design_state FROM sessions WHERE id = $1`
  - 注意：这两个路由不加 RateLimiter 中间件，仅限内部调用
- [ ] B.4.4 `backend/internal/service/postgres_store.go` 新增方法：`SaveDesignState(ctx, sid, json)` / `GetDesignState(ctx, sid) (json, error)`
  - 需在 sessions 表新增 `design_state JSONB` 列（migration 002）
- [ ] B.4.5 编写 Checkpoint 集成测试（save → load → 验证字段一致）

### B.5 GenerateWorkflow 重构
- [ ] B.5.1 修改 `agent-core/graph/generate_workflow.py` 的 `__init__`：接受 `AgentRegistry` 参数，从 registry 获取 pipeline_order，替代硬编码 import
- [ ] B.5.2 修改 `_stream_workflow()` 方法：使用 `DesignState` 替代 `PipelineState dict`，循环 `pipeline_order` 执行 Agent
- [ ] B.5.3 在每个 Agent 完成后调用 `state.mark_agent_completed()` + `checkpoint_mgr.save(state)`
- [ ] B.5.4 `run()` 方法新增 `resume: bool = False` 参数：为 True 时先 `checkpoint_mgr.load()`，跳过 `state.completed_agents` 中已完成的 Agent
- [ ] B.5.5 修改 `app/routes/generate.py`：GenerateRequest 新增 `resume: bool = False` 可选字段
- [ ] B.5.6 端到端验证：Pipeline 在 CodeGenerator 阶段手动失败 → 重新调用 `resume=True` → 从 InteractionInfer 后继续（跳过 Layout/Component/Interaction）

## Phase C: 分享功能（Week 5-7）

### C.1 后端实现
- [ ] C.1.1 添加 `github.com/matoous/go-nanoid/v2` 到 `backend/go.mod`
- [ ] C.1.2 创建 `backend/internal/service/preview_service.go`
  - `PreviewService struct { db *pgxpool.Pool }`
  - `CreateShare(ctx, sessionID) → (*SharedPreview, error)` — 读取 session.code → wrapHTML() → nanoid.New(8) → INSERT shared_previews（碰撞重试 3 次）
  - `GetByShortCode(ctx, code) → (*SharedPreview, error)` — SELECT WHERE short_code + is_active + 未过期
  - `UpdateShare(ctx, sessionID, shortCode) → error` — UPDATE html_snapshot
  - `DeleteShare(ctx, shortCode) → error` — UPDATE is_active = false
  - `wrapHTML(code) → string` — 如果不是完整 HTML 文档则包装 `<!DOCTYPE>` + Alpine.js CDN + Tailwind CDN
- [ ] C.1.3 创建 `backend/internal/handler/preview.go`
  - `POST /api/share` — 调用 PreviewService.CreateShare()，返回 `{short_code, url}`
  - `GET /p/:code` — 调用 PreviewService.GetByShortCode()，返回 `text/html` 内容
  - `PUT /api/share/:code` — 更新快照（重新从 session.code 获取最新代码）
  - `DELETE /api/share/:code` — 停用分享链接
- [ ] C.1.4 `backend/internal/app/app.go` 注册路由：
  - `r.GET("/p/:code", previewHandler.ServePreview)` — 不需要 rate limiting
  - `api.POST("/share", previewHandler.CreateShare)`
  - `api.PUT("/share/:code", previewHandler.UpdateShare)`
  - `api.DELETE("/share/:code", previewHandler.DeleteShare)`
- [ ] C.1.5 异步 view_count 递增：`GetByShortCode` 中用 `go pool.Exec(...)` 异步更新

### C.2 前端实现
- [ ] C.2.1 `frontend/src/lib/api/client.ts` 新增 `api.share(sessionId)` 方法（POST /api/share，返回 `{short_code, url}`）
- [ ] C.2.2 `frontend/src/lib/api/schemas.ts` 新增 `ShareResponseSchema` (zod: `z.object({ short_code: z.string(), url: z.string() })`)
- [ ] C.2.3 创建 `frontend/src/components/preview/ShareButton.tsx`
  - 未分享状态：显示"分享"按钮
  - 已分享状态：显示链接输入框 + "复制"按钮 + "更新"按钮
  - 分享中状态：显示 loading
  - 复制使用 `navigator.clipboard.writeText()`，成功后 2s 内显示"已复制"
- [ ] C.2.4 修改 `frontend/src/components/preview/PreviewPanel.tsx`：在操作栏（tabs 旁边）集成 ShareButton
- [ ] C.2.5 `frontend/src/stores/useProjectStore.ts` 新增 `shareUrl: string | null` 状态

### C.3 验证
- [ ] C.3.1 `go build ./...` + `go test ./...` 通过
- [ ] C.3.2 `cd frontend && npm run build` 通过
- [ ] C.3.3 端到端验证：上传图片 → 生成代码 → 点击分享 → 复制链接 → 无痕浏览器打开 → 可体验交互原型
- [ ] C.3.4 验证分享链接的 HTML 包含 Alpine.js + Tailwind CDN，交互功能正常

## Phase D: 用户系统（Week 7-9）

### D.1 数据库
- [ ] D.1.1 创建 `backend/migrations/003_create_users.sql`（users + projects 表，sessions 表添加 user_id/project_id 可选外键）
- [ ] D.1.2 运行迁移验证

### D.2 认证
- [ ] D.2.1 添加 `github.com/golang-jwt/jwt/v5` 到 `backend/go.mod`
- [ ] D.2.2 添加 `golang.org/x/crypto/bcrypt` 到 `backend/go.mod`
- [ ] D.2.3 `backend/internal/config/config.go` 新增 `JWTSecret`、`GitHubClientID`、`GitHubClientSecret` 配置
- [ ] D.2.4 创建 `backend/internal/service/auth_service.go`（Register/Login/GitHubCallback/RefreshToken）
- [ ] D.2.5 创建 `backend/internal/handler/auth.go`（POST /api/auth/register, /login, /refresh, GET /api/auth/github）
- [ ] D.2.6 创建 `backend/internal/middleware/auth.go`（JWTAuth 中间件：解析 Bearer token，设置 `c.Set("user_id", ...)`)
- [ ] D.2.7 编写认证中间件测试

### D.3 项目管理
- [ ] D.3.1 创建 `backend/internal/service/project_service.go`（CRUD + 关联 session）
- [ ] D.3.2 创建 `backend/internal/handler/project.go`（GET/POST/PUT/DELETE /api/projects）
- [ ] D.3.3 修改 upload/generate 流程：如果用户已登录，session 关联到当前 project
- [ ] D.3.4 编写项目 CRUD 测试

### D.4 前端
- [ ] D.4.1 创建 `frontend/src/app/login/page.tsx`（登录/注册表单 + GitHub OAuth 按钮）
- [ ] D.4.2 创建 `frontend/src/app/dashboard/page.tsx`（项目列表 + 创建项目入口）
- [ ] D.4.3 `frontend/src/lib/api/client.ts` 添加 `Authorization: Bearer ${token}` header
- [ ] D.4.4 `frontend/src/stores/useProjectStore.ts` 新增 `user` / `currentProject` / `projects` 状态
- [ ] D.4.5 端到端验证：注册 → 登录 → 创建项目 → 上传 → 生成 → 项目列表可见

## Phase E: 多输入源（Week 9-11）

### E.1 剪贴板粘贴
- [ ] E.1.1 `frontend/src/components/upload/UnifiedInput.tsx` 添加 `document.addEventListener('paste', ...)` 监听
- [ ] E.1.2 检测 `clipboardData.items` 中的 image 类型 → 转为 File → 复用现有 `handleFileUpload()` 流程
- [ ] E.1.3 手动验证：截图 → Ctrl+V → 图片出现在上传面板

### E.2 文字描述生成
- [ ] E.2.1 创建 `agent-core/agents/configs/text_to_ui.yaml`（根据文字描述生成 UI 规格，输出兼容 LayoutInfo + ComponentList + InteractionSpec）
- [ ] E.2.2 `agent-core/app/routes/generate.py` 支持无 images 请求 + `description` 字段
- [ ] E.2.3 前端 UnifiedInput 添加"文字描述"tab（textarea 输入，调用新的 generate 流程）
- [ ] E.2.4 端到端验证：输入"一个电商首页，有搜索栏、商品列表、底部导航" → 生成可交互原型

### E.3 Figma 集成评估
- [ ] E.3.1 评估 Figma API 导出 PNG 的可行性和限制（需要 Personal Access Token）
- [ ] E.3.2 如果可行：实现 URL 解析 → Figma API → 逐页导出 PNG → 现有 Pipeline
- [ ] E.3.3 如果不可行：记录原因，作为未来 roadmap 项
