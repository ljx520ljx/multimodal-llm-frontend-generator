# 开发路线图 | Development Roadmap

> 基于模块依赖关系和功能复杂度制定的增量开发计划

## 开发原则

1. **垂直切片优先**：每个阶段产出可运行、可验证的功能
2. **后端先行**：API 稳定后再开发对应前端
3. **核心路径优先**：先实现 MVP（单图生成），再扩展高级功能
4. **持续集成**：每个阶段结束进行联调测试

---

## 阶段总览

```
Phase 0: 基础设施 (2-3天)
    ↓
Phase 1: 后端核心 - LLM Gateway (3-4天)
    ↓
Phase 2: 后端核心 - 图片处理 & 代码生成 (3-4天)
    ↓
Phase 3: 前端核心 - 基础 UI 框架 (2-3天)
    ↓
Phase 4: 前端核心 - 编辑器 & 预览 (3-4天)
    ↓
Phase 5: MVP 联调 - 单图生成 (2-3天)
    ↓
Phase 6: 高级功能 - 多图交互分析 (4-5天)
    ↓
Phase 7: 高级功能 - 对话修正 (3-4天)
    ↓
Phase 8: 优化 & 测试 (3-4天)
```

---

## Phase 0: 基础设施搭建

### 目标
项目骨架搭建，开发环境就绪

### 任务清单

#### 0.1 后端项目初始化
```
backend/
├── cmd/server/main.go       # 入口
├── internal/
│   ├── config/config.go     # 配置加载
│   └── middleware/          # 中间件
├── go.mod
├── go.sum
├── Makefile
└── .env.example
```

- [ ] 初始化 Go module
- [ ] 集成 Gin 框架
- [ ] 配置 Viper 读取环境变量
- [ ] 实现 CORS、Logger、Recovery 中间件
- [ ] 健康检查端点 `GET /health`

#### 0.2 前端项目初始化
```
frontend/
├── src/
│   ├── app/
│   │   ├── layout.tsx
│   │   └── page.tsx
│   └── lib/utils.ts
├── package.json
├── tailwind.config.ts
├── tsconfig.json
└── .env.local.example
```

- [ ] 使用 `create-next-app` 初始化 Next.js 14 项目
- [ ] 配置 TypeScript 严格模式
- [ ] 配置 Tailwind CSS
- [ ] 配置 ESLint + Prettier
- [ ] 创建基础布局组件

#### 0.3 开发环境配置
- [ ] 创建 docker-compose.yml（可选，用于本地数据库）
- [ ] 配置 .gitignore
- [ ] 初始化 Git 仓库

### 验收标准
- 后端启动无报错，`/health` 返回 200
- 前端启动无报错，首页可访问
- 代码可提交到 Git

---

## Phase 1: 后端核心 - LLM Gateway

### 目标
封装 LLM API 调用，支持流式输出

### 任务清单

#### 1.1 Gateway 接口定义
```go
// internal/gateway/interface.go
type LLMGateway interface {
    ChatStream(ctx context.Context, prompt *LLMPrompt) <-chan StreamChunk
}
```

- [ ] 定义 `LLMPrompt` 结构（支持文本+图片）
- [ ] 定义 `StreamChunk` 结构
- [ ] 定义统一接口

#### 1.2 OpenAI 实现
- [ ] 集成 `go-openai` SDK
- [ ] 实现 `ChatStream` 方法
- [ ] 支持多模态输入（图片 Base64）
- [ ] 处理流式响应

#### 1.3 重试与错误处理
- [ ] 实现指数退避重试
- [ ] 处理 429 限流
- [ ] 处理超时

#### 1.4 Gateway 工厂
- [ ] 根据配置选择 Provider
- [ ] 支持 OpenAI / Gemini / Anthropic 切换

### 验收标准
- 单元测试通过
- 手动测试：发送文本+图片，收到流式响应

---

## Phase 2: 后端核心 - 图片处理 & 代码生成

### 目标
实现图片上传、处理和代码生成 API

### 任务清单

#### 2.1 图片处理服务
```go
// internal/service/image.go
type ImageService struct {}
func (s *ImageService) ProcessUpload(ctx, files) (*UploadResult, error)
func (s *ImageService) CompressAndEncode(data []byte) (string, error)
```

- [ ] 实现图片压缩（限制尺寸 2048x2048）
- [ ] 实现 Base64 编码
- [ ] 实现文件类型验证
- [ ] 实现临时存储（内存 Map 或 Redis）

#### 2.2 Prompt 服务
- [ ] 设计系统 Prompt 模板
- [ ] 设计单图分析 Prompt
- [ ] 实现 Prompt 构建逻辑

#### 2.3 代码后处理服务
- [ ] 实现代码块提取（从 Markdown）
- [ ] 集成 Prettier 格式化（可选）
- [ ] 实现基础语法校验

#### 2.4 生成服务编排
```go
// internal/service/generate.go
func (s *GenerateService) GenerateStream(ctx, req) <-chan GenerateEvent
```

- [ ] 编排 Image → Prompt → LLM → PostProcess 流程
- [ ] 实现事件分类（thinking / code / error / done）

#### 2.5 HTTP Handler
- [ ] `POST /api/upload` - 图片上传
- [ ] `POST /api/generate` - 代码生成（SSE）

### 验收标准
- 上传图片返回 session_id
- 调用 generate 返回流式代码
- 使用 curl 或 Postman 手动验证

---

## Phase 3: 前端核心 - 交互原型验证平台

### 目标
搭建以**交互体验为核心**的前端页面框架，支持双模式布局

### 核心定位
- **交互体验优先**：重点是让用户体验交互流程，而非查看代码
- **主要用户**：产品经理、设计师
- **次要用户**：开发者

### 任务清单

#### 3.1 基础 UI 组件
```
src/components/ui/
├── Button.tsx
├── Card.tsx
├── Input.tsx
├── Skeleton.tsx
└── index.ts
```

- [ ] 实现 Button 组件（variant, size, loading）
- [ ] 实现 Card 组件
- [ ] 实现 Input 组件
- [ ] 实现 Skeleton 组件
- [ ] 配置 cn() 工具函数

#### 3.2 双模式布局组件
```
src/components/layout/
├── Header.tsx
├── ViewModeToggle.tsx    # 模式切换（体验/开发）
└── MainLayout.tsx        # 双模式布局容器
```

- [ ] 实现 Header 组件
- [ ] 实现模式切换组件（体验模式 / 开发模式）
- [ ] 实现体验模式布局（预览 80% + 上传 20%）
- [ ] 实现开发模式布局（三栏：上传 | 编辑器 | 预览）
- [ ] 实现模式切换动画
- [ ] 实现响应式设计

#### 3.3 状态管理
```typescript
// src/stores/useProjectStore.ts
interface ProjectState {
  // 视图状态
  viewMode: 'experience' | 'developer';  // 双模式
  codeExpanded: boolean;                  // 代码面板展开状态
  // 项目状态
  images: ImageFile[];
  generatedCode: string;
  status: 'idle' | 'generating' | 'completed';
}
```

- [ ] 集成 Zustand
- [ ] 定义全局状态结构（含 viewMode）
- [ ] 实现 Actions
- [ ] 实现 viewMode 持久化（localStorage）

#### 3.4 API 客户端
- [ ] 实现 fetch 封装
- [ ] 实现 SSE 流式读取工具

### 验收标准
- [ ] 默认进入体验模式，预览区占据主要空间
- [ ] 可切换到开发模式，显示三栏布局
- [ ] 模式切换流畅，状态持久化正常
- [ ] 基础 UI 组件可用
- [ ] Zustand 状态可读写

---

## Phase 4: 前端核心 - 交互预览 & 编辑器

### 目标
实现**交互预览**（核心）和代码编辑（辅助）功能

### 任务清单

#### 4.1 图片上传组件
```
src/features/upload/
├── components/
│   ├── ImageUploader.tsx
│   ├── ImagePreviewCard.tsx
│   └── ImageSortableList.tsx
└── hooks/useImageUpload.ts
```

- [ ] 实现拖拽上传（react-dropzone）
- [ ] 实现图片预览卡片
- [ ] 实现拖拽排序（@hello-pangea/dnd）
- [ ] 实现上传到后端

#### 4.2 交互预览（核心）
```
src/components/preview/
├── SandpackPreview.tsx
└── PreviewError.tsx
```

- [ ] 集成 Sandpack
- [ ] 配置 React + Tailwind 模板
- [ ] **确保支持点击、状态切换等交互**
- [ ] 实现错误边界处理
- [ ] 实现友好的错误提示

#### 4.3 代码编辑器（可折叠）
```
src/components/editor/
├── CodeEditor.tsx
├── CodePanel.tsx       # 可折叠面板
└── EditorSkeleton.tsx
```

- [ ] 集成 Monaco Editor（动态导入）
- [ ] 配置 TypeScript/JSX 语法高亮
- [ ] 实现代码只读/可编辑切换
- [ ] 实现可折叠代码面板（体验模式用）

### 验收标准
- [ ] 可上传图片并显示预览
- [ ] 可拖拽调整图片顺序
- [ ] Sandpack 可渲染 React 代码，**支持交互**
- [ ] Monaco Editor 正常显示代码
- [ ] 体验模式下代码面板默认折叠

---

## Phase 5: MVP 联调 - 交互原型体验

### 目标
打通完整链路：上传 → 生成 → **体验交互**

### 任务清单

#### 5.1 生成功能集成
```
src/features/generation/
├── components/
│   ├── GenerateButton.tsx
│   └── GenerationStatus.tsx
└── hooks/useCodeGeneration.ts
```

- [ ] 实现生成按钮
- [ ] 实现生成状态展示
- [ ] 实现 SSE 流式接收
- [ ] 实现思考过程展示

#### 5.2 端到端联调
- [ ] 上传单张图片
- [ ] 点击生成，观察思考过程
- [ ] 预览区实时渲染
- [ ] **验证预览区可交互**（点击按钮等）

#### 5.3 用户体验优化
- [ ] 体验模式下预览全屏显示
- [ ] 代码面板可折叠查看
- [ ] 模式切换流畅

#### 5.4 错误处理
- [ ] 网络错误提示
- [ ] 生成失败重试
- [ ] 预览编译错误展示（友好提示）

### 验收标准
- [ ] 完整流程可运行
- [ ] 生成的代码可在 Sandpack 中渲染
- [ ] **预览区可交互**（点击、状态切换）
- [ ] 体验模式下用户可专注于交互体验
- [ ] 错误有友好提示

---

## Phase 6: 高级功能 - 多图交互分析

### 目标
支持多图上传，自动推断交互逻辑

### 任务清单

#### 6.1 多图 Prompt 优化
- [ ] 设计多图对比分析 Prompt
- [ ] 实现视觉差异分析指令
- [ ] 优化 CoT 思考链

#### 6.2 状态机代码生成
- [ ] 引导模型生成 useState
- [ ] 引导模型生成事件处理（onClick）
- [ ] 引导模型实现页面切换逻辑

#### 6.3 前端适配
- [ ] 支持选择多张图片顺序
- [ ] 展示图片序列关系

#### 6.4 Prompt 调优
- [ ] 建立测试用例集
- [ ] A/B 测试不同 Prompt 策略
- [ ] 记录最佳 Prompt 模板

### 验收标准
- 上传 2-3 张设计稿
- 生成包含状态切换的代码
- 预览中可交互（点击切换状态）

---

## Phase 7: 高级功能 - 对话修正

### 目标
通过自然语言修改生成的代码

### 任务清单

#### 7.1 后端对话服务
```go
// internal/service/chat.go
func (s *ChatService) ProcessMessage(ctx, req) <-chan ChatEvent
```

- [ ] 实现对话上下文管理
- [ ] 实现增量修改 Prompt
- [ ] 实现代码 diff 生成（可选）

#### 7.2 后端 API
- [ ] `POST /api/chat` - 对话交互（SSE）

#### 7.3 前端对话组件
```
src/features/chat/
├── components/
│   ├── ChatPanel.tsx
│   ├── ChatInput.tsx
│   └── ChatMessage.tsx
└── hooks/useChat.ts
```

- [ ] 实现对话面板 UI
- [ ] 实现消息输入
- [ ] 实现消息历史展示
- [ ] 实现代码增量更新

### 验收标准
- 可发送修改指令（如"把按钮改成蓝色"）
- 代码实时更新
- 预览反映修改结果

---

## Phase 8: 优化 & 测试

### 目标
性能优化、稳定性提升、文档完善

### 任务清单

#### 8.1 性能优化
- [ ] 图片压缩优化
- [ ] 前端代码分割
- [ ] SSE 连接优化
- [ ] 后端并发处理优化

#### 8.2 稳定性
- [ ] 添加单元测试（覆盖率 > 60%）
- [ ] 添加 E2E 测试（关键路径）
- [ ] 错误监控（Sentry 可选）

#### 8.3 用户体验
- [ ] 加载状态优化
- [ ] 错误提示优化
- [ ] 响应式适配
- [ ] 快捷键支持

#### 8.4 文档
- [ ] 更新 README
- [ ] API 文档（OpenAPI）
- [ ] 部署文档

### 验收标准
- 性能指标达标（见 L1 文档）
- 测试覆盖核心功能
- 文档完整可用

---

## 里程碑总结

| 里程碑 | 阶段 | 核心产出 |
|--------|------|----------|
| **M1: 基础就绪** | Phase 0-1 | 项目可运行，LLM 可调用 |
| **M2: 后端完成** | Phase 2 | API 可用，单图可生成代码 |
| **M3: 前端完成** | Phase 3-4 | UI 完整，编辑器/预览可用 |
| **M4: MVP** | Phase 5 | 单图生成端到端可用 |
| **M5: 核心功能** | Phase 6-7 | 多图分析 + 对话修正 |
| **M6: 发布就绪** | Phase 8 | 优化完成，可演示 |

---

## 风险与应对

| 风险 | 可能性 | 影响 | 应对措施 |
|------|--------|------|----------|
| LLM 生成代码不可运行 | 高 | 中 | 增加后处理校验，自动修复常见错误 |
| Prompt 效果不稳定 | 中 | 高 | 建立评估集，持续调优，Few-Shot |
| Sandpack 渲染失败 | 中 | 中 | 错误边界，友好提示，降级方案 |
| API 限流 | 低 | 低 | 重试机制，用户提示 |
