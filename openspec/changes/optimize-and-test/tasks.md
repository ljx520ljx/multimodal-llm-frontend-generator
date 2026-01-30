# Tasks: optimize-and-test

## Task Groups

任务按依赖关系分组，同组内可并行执行。

---

## Group A: 基础设施（可并行）

### A1. 配置前端测试框架 (Vitest) ✅

**依赖**: 无

1. [x] 安装依赖：`vitest`, `@testing-library/react`, `@testing-library/user-event`, `jsdom`
2. [x] 创建 `vitest.config.ts` 配置文件
3. [x] 创建 `vitest/setup.ts` 全局设置
4. [x] 在 `package.json` 添加测试脚本
5. [x] 创建示例测试验证配置正确

**验收**: ✅ `npm run test` 执行成功

### A2. 配置 E2E 测试框架 (Playwright) ✅

**依赖**: 无

1. [x] 安装 Playwright：`@playwright/test`
2. [x] 创建 `playwright.config.ts`
3. [x] 创建 `e2e/fixtures/` 测试数据目录
4. [x] 添加测试用图片文件
5. [x] 在 `package.json` 添加 E2E 脚本

**验收**: ✅ `npm run e2e` 可执行

### A3. 安装图片压缩库 ✅

**依赖**: 无

1. [x] 安装 `browser-image-compression`
2. [x] 创建 `lib/utils/imageCompression.ts` 工具函数
3. [x] 添加压缩配置常量

**验收**: ✅ 压缩函数可正常调用

---

## Group B: 前端优化

### B1. 实现前端图片压缩 ✅

**依赖**: A3

1. [x] 在 `ImageDropzone` 组件集成压缩逻辑
2. [x] 添加压缩进度显示
3. [x] 处理压缩失败场景
4. [x] 更新 `image-upload` spec

**验收**: ✅ 上传 > 2MB 图片时自动压缩

### B2. 实现代码分割 ✅

**依赖**: 无

1. [x] 将 `CodeEditor` 改为动态导入
2. [x] 将 `SandpackPreview` 改为动态导入
3. [x] 添加加载骨架屏组件
4. [x] 验证首屏包大小减少

**验收**: ✅ 使用 Next.js dynamic imports 实现懒加载

---

## Group C: 前端测试（依赖 A1）

### C1. Store 测试 ✅

**依赖**: A1

1. [x] 测试 `useProjectStore` 状态变更
2. [x] 测试 `useChatStore` 消息管理
3. [x] 测试持久化逻辑

**验收**: ✅ Store 测试 34 个用例通过

### C2. Hook 测试

**依赖**: A1

1. [ ] 测试 `useGeneration` hook
2. [ ] 测试 `useChat` hook
3. [ ] Mock API 调用

**验收**: Hook 测试覆盖率 > 70%

### C3. 组件测试

**依赖**: A1

1. [ ] 测试 `ImageDropzone` 拖拽上传
2. [ ] 测试 `ImageList` 排序删除
3. [ ] 测试 `ChatInput` 输入发送
4. [ ] 测试 `GenerateButton` 状态切换

**验收**: 核心组件测试覆盖率 > 60%

---

## Group D: 后端测试

### D1. config 模块测试 ✅

**依赖**: 无

1. [x] 测试配置文件加载
2. [x] 测试环境变量覆盖
3. [x] 测试默认值处理

**验收**: ✅ `internal/config` 覆盖率 100%

### D2. middleware 模块测试 ✅

**依赖**: 无

1. [x] 测试 CORS 中间件
2. [x] 测试日志中间件
3. [x] 测试 Recovery 中间件

**验收**: ✅ `internal/middleware` 覆盖率 100%

### D3. gateway 模块测试补充

**依赖**: 无

1. [ ] 补充 `openai_compatible.go` 测试
2. [ ] Mock HTTP 响应
3. [ ] 测试错误处理路径

**验收**: `internal/gateway` 覆盖率 > 70% (当前 44.5%)

### D4. app 模块测试

**依赖**: 无

1. [ ] 测试依赖注入
2. [ ] 测试服务初始化
3. [ ] Mock 外部依赖

**验收**: `internal/app` 覆盖率 > 60%

---

## Group E: E2E 测试（依赖 A2） ✅

### E1. 上传流程 E2E ✅

**依赖**: A2

1. [x] 测试拖拽上传图片
2. [x] 测试图片排序
3. [x] 测试图片删除

**验收**: ✅ 上传流程 E2E 测试文件创建

### E2. 生成流程 E2E ✅

**依赖**: A2, E1

1. [x] 测试点击生成按钮
2. [x] 测试流式输出显示
3. [x] 测试预览渲染

**验收**: ✅ 生成流程 E2E 测试文件创建

### E3. 聊天流程 E2E ✅

**依赖**: A2, E2

1. [x] 测试发送聊天消息
2. [x] 测试代码更新
3. [x] 测试预览刷新

**验收**: ✅ 聊天流程 E2E 测试文件创建

---

## Group F: 文档

### F1. 更新 README ✅

**依赖**: 无

1. [x] 更新项目描述
2. [x] 添加安装步骤
3. [x] 添加开发指南
4. [x] 添加 API 端点说明
5. [x] 添加架构图

**验收**: ✅ README 完整覆盖使用场景

### F2. 创建 API 文档 ✅

**依赖**: 无

1. [x] 创建/更新 `api/openapi.yaml`
2. [x] 文档化所有 API 端点
3. [x] 添加请求/响应示例
4. [x] 添加错误码说明

**验收**: ✅ OpenAPI 规范完整

---

## Summary

| Group | Tasks | 完成状态 |
|-------|-------|----------|
| A | 3 | ✅ 全部完成 |
| B | 2 | ✅ 全部完成 |
| C | 3 | 1/3 完成 |
| D | 4 | 2/4 完成 |
| E | 3 | ✅ 全部完成 |
| F | 2 | ✅ 全部完成 |

**总计**: 17 个任务，已完成 13 个
