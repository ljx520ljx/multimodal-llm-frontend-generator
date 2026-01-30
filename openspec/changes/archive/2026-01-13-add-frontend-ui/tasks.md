# Tasks: Phase 3 前端基础 UI

> 核心定位：交互原型验证工具，让产品/设计师体验和验证 UI 交互流程

## 1. 项目依赖安装

- [x] 1.1 安装 Zustand 状态管理
- [x] 1.2 安装 @hello-pangea/dnd 拖拽库
- [x] 1.3 安装 @monaco-editor/react 代码编辑器
- [x] 1.4 安装 @codesandbox/sandpack-react 预览沙箱
- [x] 1.5 安装 react-dropzone 拖拽上传

## 2. 基础 UI 组件

- [x] 2.1 创建 `src/components/ui/Button.tsx` - 按钮组件
- [x] 2.2 创建 `src/components/ui/Card.tsx` - 卡片组件
- [x] 2.3 创建 `src/components/ui/Input.tsx` - 输入框组件
- [x] 2.4 创建 `src/components/ui/Skeleton.tsx` - 骨架屏组件
- [x] 2.5 创建 `src/components/ui/index.ts` - 组件导出

## 3. 布局组件（双模式）

- [x] 3.1 创建 `src/components/layout/Header.tsx` - 页头组件
- [x] 3.2 创建 `src/components/layout/ViewModeToggle.tsx` - 模式切换组件（体验/开发）
- [x] 3.3 创建 `src/components/layout/MainLayout.tsx` - 双模式主布局
- [x] 3.4 更新 `app/page.tsx` - 主页面

## 4. 状态管理

- [x] 4.1 创建 `src/types/index.ts` - 类型定义（含 ViewMode）
- [x] 4.2 创建 `src/stores/useProjectStore.ts` - Zustand store（含 viewMode 持久化）

## 5. API 客户端

- [x] 5.1 创建 `src/lib/api/client.ts` - API 基础封装
- [x] 5.2 创建 `src/lib/api/sse.ts` - SSE 流式处理工具

## 6. 图片上传功能

- [x] 6.1 创建 `src/components/upload/ImageDropzone.tsx` - 拖拽上传区
- [x] 6.2 创建 `src/components/upload/ImageList.tsx` - 可排序图片列表
- [x] 6.3 创建 `src/components/upload/UploadPanel.tsx` - 上传面板

## 7. 代码编辑器（可折叠）

- [x] 7.1 创建 `src/components/editor/CodeEditor.tsx` - Monaco 编辑器封装
- [x] 7.2 创建 `src/components/editor/CodePanel.tsx` - 可折叠代码面板（体验模式用）
- [x] 7.3 创建 `src/components/editor/EditorPanel.tsx` - 完整编辑器面板（开发模式用）

## 8. 交互预览（核心）

- [x] 8.1 创建 `src/components/preview/SandpackPreview.tsx` - Sandpack 预览封装
- [x] 8.2 创建 `src/components/preview/PreviewPanel.tsx` - 预览面板

## 9. 代码生成功能

- [x] 9.1 创建 `src/lib/hooks/useGeneration.ts` - 生成 Hook

## 10. 集成与联调

- [x] 10.1 在主页面集成所有组件
- [x] 10.2 实现体验模式布局（预览为主，代码折叠）
- [x] 10.3 实现开发模式布局（三栏）
- [x] 10.4 配置 tsconfig.json 路径别名
- [x] 10.5 TypeScript 类型检查通过
- [x] 10.6 开发服务器启动成功 (`npm run dev`)
- [ ] 10.7 配置环境变量 `.env.local`（后端就绪后）
- [ ] 10.8 端到端联调：上传 → 生成 → 交互体验（需后端支持）

## 依赖关系

```
1 (依赖安装)
     ↓
2 (基础 UI) → 3 (布局) → 主框架就绪
     ↓
4 (类型/状态) → 5 (API)
     ↓
6, 7, 8, 9 可并行开发
     ↓
10 (集成联调)
```

## 验收标准

### 体验模式（产品/设计师视角）
- [x] 默认进入体验模式，预览区占据主要空间 (~80%)
- [x] 可上传多张设计稿，支持拖拽排序
- [x] 代码面板默认折叠，可展开查看
- [ ] 点击生成后，预览区展示可交互的原型（需后端）
- [ ] 可在预览区点击、交互，验证流程（需后端）

### 开发模式（开发者视角）
- [x] 可切换到开发模式，显示三栏布局
- [x] Monaco Editor 动态加载，语法高亮正确
- [ ] 可手动编辑代码，预览实时更新（需完善联动）

### 通用
- [x] 所有依赖安装成功
- [x] `npm run dev` 无报错，页面正常访问
- [x] 模式切换流畅，状态通过 localStorage 持久化
- [x] TypeScript 严格模式检查通过
- [x] SSE 流式读取工具已实现
