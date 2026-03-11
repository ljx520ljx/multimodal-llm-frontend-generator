# Design: Production Architecture Upgrade

## Context

系统已完成 3 轮代码审查和修复（SSE 批处理、iframe 安全、graceful shutdown、rate limiting、zod 验证、结构化错误、per-agent temperature 等），核心功能稳定。

现在需要解决 6 个结构性缺陷（内存存储、Agent 硬编码、Pipeline 不可恢复、无分享、无用户体系），将系统从"技术验证"升级为"可用产品"。

**约束条件**：
- 毕业设计项目，截止日期 2026-04-20
- 单人开发，需严格控制改动范围
- 必须保持现有功能（upload → generate → chat）完全兼容
- SessionStore 接口已完善（`session_store.go:6-36`，9 个方法），新实现只需实现该接口

**参考架构**：NebulaForge AI (`/Users/ljx/datanebula/nebulaforge-ai`) 的 Skill YAML 注册、WorkingMemory 累积、SessionStore checkpoint 模式。

## Goals / Non-Goals

**Goals:**
- 数据持久化：PostgreSQL 实现 SessionStore 接口，服务重启不丢数据
- Agent 插件化：YAML 声明式配置，新增 Agent 无需修改 Python 代码
- DesignState 类型安全：Pydantic 模型替代 TypedDict，支持序列化和验证
- Pipeline 断点续跑：每个 Agent 完成后 checkpoint，失败可从断点恢复
- 分享功能：生成 nanoid 8 位短链接，访客打开即可体验 HTML + Alpine.js 原型
- 用户系统（基础版）：JWT 认证 + GitHub OAuth2 + 项目 CRUD

**Non-Goals:**
- 实时协作编辑（Google Docs 式多人同时编辑）
- React/Vue 代码生成（保持 HTML + Alpine.js + Tailwind 技术栈）
- 付费/订阅系统
- Kubernetes 编排、微服务拆分
- 完整的 Figma 插件开发（仅评估 Figma API 导入可行性）
- 将 Go 后端替换为 Python（Go 继续作为 HTTP 网关和会话管理层）

## Architecture

### 分层变更

```
现有 3 层（不变）:
  Frontend (Next.js :3000) → Go Backend (Gin :8080) → Agent Core (FastAPI :8081)

新增 1 层:
  Go Backend → PostgreSQL (Session/Preview/User 持久化)

新增 1 路由:
  Go Backend → GET /p/:code (Preview Host，直接返回 HTML)
```

### 核心变更 1: PostgresStore（替代 MemoryStore）

```
现有接口（不变）:                    新增实现:
session_store.go:6-36               postgres_store.go
┌──────────────────────────┐        ┌──────────────────────────┐
│ SessionStore interface   │        │ PostgresStore struct      │
│                          │        │                          │
│ Create() (*Session)      │◄───────│ CREATE TABLE sessions    │
│ Get(id) (*Session)       │        │ + session_images JOIN    │
│ Update(*Session)         │        │ + session_history JOIN   │
│ Delete(id)               │        │                          │
│ AddImage(sid, *ImageData)│        │ INSERT session_images    │
│ GetImages(sid, ids)      │        │ SELECT WHERE session_id  │
│ UpdateCode(sid, code)    │        │ UPDATE sessions.code     │
│ AddHistory(sid, entry)   │        │ INSERT session_history   │
│ GetHistory(sid, limit)   │        │ SELECT ORDER BY LIMIT    │
│ Close()                  │        │ pool.Close()             │
└──────────────────────────┘        └──────────────────────────┘

数据库表设计（图片和历史拆为独立表）:
sessions ──1:N──► session_images (id, session_id, base64_data, sort_order)
         ──1:N──► session_history (id, session_id, role, content, entry_type)
         ──1:N──► shared_previews (id, session_id, short_code, html_snapshot)
```

### 核心变更 2: Agent 插件化

```
现有（硬编码）:                        目标（YAML 驱动）:
generate_workflow.py                   agents/configs/
├── from agents.layout_analyzer        ├── layout_analyzer.yaml
│   import LayoutAnalyzerAgent         ├── component_detector.yaml
├── from agents.component_detector     ├── interaction_infer.yaml
│   import ComponentDetectorAgent      ├── code_generator.yaml
├── from agents.interaction_infer      └── chat_agent.yaml
│   import InteractionInferAgent
├── from agents.code_generator                 ↓ 启动时加载
│   import CodeGeneratorAgent          AgentRegistry
│                                        .get(name) → AgentConfig
│ self.layout_agent = Layout...(llm)     .list_by_tag("pipeline") → [...]
│ self.component_agent = Comp...(llm)    .get_pipeline_order() → ["layout_analyzer", ...]
│ self.interaction_agent = Int...(llm)   .create_agent(name, llm) → BaseAgent
│ self.code_agent = Code...(llm)
```

### 核心变更 3: DesignState + Checkpoint

```
现有 PipelineState (TypedDict):           目标 DesignState (Pydantic):
state.py (44行)                           state.py (~80行)
┌──────────────────────────┐              ┌──────────────────────────────────┐
│ session_id: str          │              │ session_id: str                  │
│ images: list[dict]       │              │ images: list[dict]               │
│ layout_info: Optional    │              │ layout_info: Optional[LayoutInfo]│
│ component_info: Optional │              │ component_info: ...              │
│ interaction_info: Optional│             │ interaction_info: ...            │
│ generated_code: Optional │              │ generated_code: ...              │
│ validation_errors: list  │              │ validation: Optional[Valid...]   │
│ retry_count: int         │              │ retry_count: int                 │
│ final_html: Optional     │              │ final_html: Optional[str]        │
│ success: bool            │              │ success: bool                    │
│                          │              │                                  │
│                          │              │ # 新增 Checkpoint 元数据         │
│                          │              │ current_agent: Optional[str]     │
│                          │              │ completed_agents: list[str]      │
│                          │              │ checkpoints: dict[str, str]      │
│                          │              │ errors: list[AgentError]         │
│                          │              │                                  │
│                          │              │ mark_agent_completed(name)       │
│                          │              │ get_next_agent(order) → str|None │
│                          │              │ set_agent_output(name, output)   │
└──────────────────────────┘              └──────────────────────────────────┘

Checkpoint 流程:
  Agent 完成 → state.mark_agent_completed(name)
             → CheckpointManager.save(state)
             → PUT /api/internal/checkpoint/{session_id}  (Go API)
             → UPDATE sessions SET design_state = $1

  Pipeline 恢复 → CheckpointManager.load(session_id)
                → GET /api/internal/checkpoint/{session_id}  (Go API)
                → SELECT design_state FROM sessions
                → state.get_next_agent(pipeline_order) → 跳过已完成的 Agent
```

### 核心变更 4: Preview Host（分享功能）

```
创建分享:
  用户点击"分享" → POST /api/share {session_id}
    → PreviewService.CreateShare()
      → SELECT code FROM sessions WHERE id = $1
      → wrapHTML(code)  // 确保完整 HTML 文档
      → nanoid.New(8)   // 生成短码 "xK4m9qZb"
      → INSERT shared_previews (session_id, short_code, html_snapshot)
    → 返回 {url: "https://domain.com/p/xK4m9qZb"}

访客访问:
  浏览器打开 https://domain.com/p/xK4m9qZb
    → GET /p/xK4m9qZb
    → PreviewService.GetByShortCode("xK4m9qZb")
      → SELECT html_snapshot FROM shared_previews WHERE short_code = $1
      → UPDATE view_count + 1 (异步)
    → c.Data(200, "text/html", html)
    → 浏览器直接渲染可交互原型（Alpine.js + Tailwind CDN 自包含）
```

## Decisions

### 1. 数据库选型

**Decision:** PostgreSQL 16+ (pgx 驱动)

**Why:**
- 我们的数据有明确的关系模型（users → projects → sessions → images/history/previews）
- JSONB 支持灵活的 DesignState checkpoint 存储
- 并发量 < 100 QPS，PostgreSQL 性能足够
- 部署在 docker-compose 中，一个服务搞定

**Alternatives:**
- Redis: 适合缓存但不适合主存储，关系查询弱
- SQLite: 不支持并发写入（多个 SSE goroutine 同时写 session）
- MongoDB: 关系查询弱，额外学习成本

### 2. 图片/历史拆表 vs JSONB 字段

**Decision:** 独立表（`session_images` + `session_history`）

**Why:**
- `session_images.base64_data` 单条可达 1-2MB，JSONB UPDATE 会重写整个字段
- `AddImage()` 和 `AddHistory()` 是高频操作，独立 INSERT 比 JSONB APPEND 高效
- 历史需要 ORDER BY + LIMIT 分页
- 现有 SessionStore 接口的 `AddImage`/`AddHistory` 方法天然对应 INSERT

**Alternative:** 全部存 JSONB → 代价是每次 AddImage 都要 UPDATE 整个 JSON 数组

### 3. Agent 配置格式

**Decision:** YAML 文件，启动时加载到内存

**Why:**
- Prompt 内容需要多行字符串 → YAML 原生支持（`|` 块标量）
- 支持注释 → 便于 Prompt 调试和维护
- NebulaForge 验证了此模式的可行性（Skill YAML 注册）

**Alternatives:**
- 数据库存储: 过度工程
- Python 装饰器: 当前方案，不够灵活（修改 Prompt 要改代码）
- JSON: 不支持多行字符串和注释

### 4. Checkpoint 存储位置

**Decision:** 通过 Go 后端 API 存到 PostgreSQL（`sessions.design_state` JSONB 字段）

**Why:**
- 已有 PostgreSQL，不增加基础设施
- Checkpoint 排除 images 后约 10KB（LayoutInfo + ComponentList + InteractionSpec + 元数据）
- 与 session 过期清理联动（session 过期 → checkpoint 自然清理）
- Python → Go API → PostgreSQL 的跨服务调用增加约 5ms 延迟，完全可接受

**Alternative:** Python 直连 PostgreSQL → 增加 Python 端 DB 依赖，违反"Go 管会话，Python 管 Agent"的职责分离

### 5. 分享方案

**Decision:** 静态 HTML 快照 + nanoid 8 位短链接

**Why:**
- HTML + Alpine.js + Tailwind CDN = 自包含零依赖，任何浏览器直接打开
- nanoid 8 位字符空间 64^8 ≈ 281 万亿，百万级数据无碰撞风险
- 不需要前端应用渲染，直接 `c.Data(200, "text/html", html)` → 延迟 < 50ms

**Alternatives:**
- iframe embed（需要运行完整前端应用）: 成本高，依赖前端部署
- CodeSandbox/StackBlitz 链接: 第三方依赖，不可控

### 6. 用户认证方案

**Decision:** JWT + GitHub OAuth2

**Why:**
- JWT 无状态，Go 后端处理简单（Gin 中间件 ~30 行）
- GitHub OAuth2 是目标用户群体（产品/设计/开发者）最熟悉的方式
- 后续可扩展微信/Google 登录

### 7. MVP 策略

**Decision:** Phase A + C（数据持久化 + 分享功能）为 MVP，约 6-7 周

**Why:**
- 持久化解决"数据丢失"这个最大痛点
- 分享功能是产品核心差异化——"不只是代码生成，而是可分享的交互 Demo"
- Agent 插件化（Phase B）和用户系统（Phase D）可后续迭代
- 毕设截止日 4/20 前有足够缓冲

## Risks / Trade-offs

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| PostgresStore 实现不兼容 SessionStore 接口 | 低 | 高 | 接口已定义（9 个方法），编译时即可发现不兼容 |
| 图片 Base64 存 DB 导致表膨胀 | 中 | 中 | 单张图片压缩后 < 1MB，sessions 定期过期清理 |
| YAML 配置解析错误影响 Pipeline | 低 | 中 | 启动时校验所有 YAML（Pydantic 验证），启动失败则快速失败 |
| Checkpoint 跨服务调用失败 | 低 | 低 | Checkpoint 保存失败只 warning 日志，不阻塞 Pipeline 执行 |
| 分享链接 XSS/安全风险 | 中 | 高 | 1) HTML 内容来自 LLM 生成，非用户注入；2) 可添加 CSP header 限制 |
| 毕设截止前未完成所有 Phase | 中 | 高 | MVP = Phase A + C（6-7 周），Phase D 基础版（+2 周），总计 8-9 周在截止日前完成 |

## Migration Plan

### Step 1: 接口验证（无功能变更）

确认现有 `SessionStore` 接口的 9 个方法已覆盖所有调用方需求。

调用方清单（通过 `findReferences` 确认）：
- `app.go:55` — `NewMemoryStore()` 初始化
- `upload.go:35-42` — `Create()`, `Get()`, `AddImage()`
- `generate_service.go:60-80` — `Get()`, `GetImages()`, `UpdateCode()`, `AddHistory()`
- `agent_generate_service.go:40-55` — `Get()`, `GetImages()`, `UpdateCode()`
- `chat.go:25` — 通过 `GenerateService.Chat()` 间接使用

**结论**：接口完整，无需新增方法。

### Step 2: 实现 PostgresStore

- 编写 SQL schema（4 张表 + 索引）
- 实现 PostgresStore 的 9 个接口方法
- 使用 testcontainers-go 编写集成测试
- 集成测试覆盖：Create → AddImage → Get → UpdateCode → AddHistory → GetHistory

### Step 3: 条件切换

```go
// app.go 修改 initServices()
if a.Config.DatabaseURL != "" {
    pool, err := pgxpool.New(ctx, a.Config.DatabaseURL)
    a.SessionStore = service.NewPostgresStore(pool, a.Config.SessionHistoryLimit)
} else {
    a.SessionStore = service.NewMemoryStore(a.Config.SessionTTL, a.Config.SessionHistoryLimit)
}
```

- 设置 `DATABASE_URL` → PostgresStore
- 不设置 → MemoryStore（向后兼容）

### Step 4: 端到端验证

Upload → Generate → Chat → 重启 → 数据仍在

### Step 5: Agent 逐个迁移

1. 先迁移 `layout_analyzer`（最简单），验证 YAML → AgentRegistry → BaseAgent 流程
2. 确认无回归后批量迁移其余 4 个 Pipeline Agent + ChatAgent

## Open Questions

- [ ] PostgreSQL 部署方式？docker-compose 内嵌 vs 云托管（Supabase/Neon）
- [ ] 分享页面是否需要添加"由 XX 平台生成"的水印/品牌标识？
- [ ] 是否需要私密分享（密码/登录才能查看）？
- [ ] Agent YAML 是否需要版本管理（同一 Agent 多版本 A/B 测试不同 Prompt）？
- [ ] sessions 表的 `design_state` JSONB 字段最大允许多大？（排除 images 后预计 < 50KB）
