# Design: optimize-and-test

## Overview

本设计文档描述 Phase 8 优化与测试阶段的技术方案。

## 1. 图片压缩优化

### 当前状态

```
User uploads image → FormData → Backend → Compress (Go imaging)
```

问题：大图片（如 5MB+）直接传输，浪费带宽和时间。

### 目标架构

```
User uploads image → Frontend compress → FormData → Backend → Validate & Store
```

### 技术方案

使用 `browser-image-compression` 库在前端压缩：

```typescript
// lib/utils/imageCompression.ts
import imageCompression from 'browser-image-compression';

export async function compressImage(file: File): Promise<File> {
  const options = {
    maxSizeMB: 2,           // 目标大小
    maxWidthOrHeight: 2048, // 最大尺寸
    useWebWorker: true,     // Web Worker 异步处理
    fileType: 'image/webp', // 转换为 WebP
  };

  if (file.size <= 2 * 1024 * 1024) {
    return file; // 小于 2MB 不压缩
  }

  return imageCompression(file, options);
}
```

### 压缩策略

| 原始大小 | 压缩行为 |
|----------|----------|
| < 2MB | 不压缩 |
| 2MB - 10MB | 压缩至 2MB，保持 80% 质量 |
| > 10MB | 压缩至 2MB，降低质量至 70% |

## 2. 前端代码分割

### 当前状态

所有组件在首屏加载，包括：
- Monaco Editor (~800KB)
- Sandpack (~500KB)

### 目标架构

使用 Next.js `dynamic` 进行代码分割：

```typescript
// 懒加载 Monaco Editor
const CodeEditor = dynamic(
  () => import('@/components/editor/CodeEditor'),
  {
    loading: () => <EditorSkeleton />,
    ssr: false
  }
);

// 懒加载 Sandpack Preview
const SandpackPreview = dynamic(
  () => import('@/components/preview/SandpackPreview'),
  {
    loading: () => <PreviewSkeleton />,
    ssr: false
  }
);
```

### 分割策略

| 组件 | 加载时机 | 预估节省 |
|------|----------|----------|
| Monaco Editor | 开发模式切换时 | ~800KB |
| Sandpack | 代码生成完成后 | ~500KB |
| ChatPanel | 首次打开聊天时 | ~50KB |

## 3. 测试架构

### 前端测试

```
vitest/
├── setup.ts              # 测试配置
├── mocks/
│   ├── zustand.ts       # Store mock
│   └── api.ts           # API mock
└── __tests__/
    ├── components/      # 组件测试
    ├── hooks/           # Hook 测试
    └── stores/          # Store 测试
```

**技术选型**：

| 工具 | 用途 |
|------|------|
| Vitest | 测试框架（Vite 生态，比 Jest 更快） |
| @testing-library/react | 组件渲染和交互 |
| @testing-library/user-event | 用户事件模拟 |
| MSW | API Mock |

### 后端测试

需要补充测试的模块：

| 模块 | 当前覆盖率 | 目标覆盖率 | 策略 |
|------|------------|------------|------|
| cmd/server | 0% | 跳过 | 入口点，集成测试覆盖 |
| internal/app | 0% | 60% | 依赖注入测试 |
| internal/config | 0% | 80% | 配置加载测试 |
| internal/gateway | 44.5% | 70% | Mock HTTP 客户端 |
| internal/middleware | 0% | 70% | 中间件单元测试 |

### E2E 测试

```
e2e/
├── playwright.config.ts
├── fixtures/
│   └── test-images/     # 测试用图片
└── tests/
    ├── upload.spec.ts   # 上传流程
    ├── generate.spec.ts # 生成流程
    └── chat.spec.ts     # 聊天修正流程
```

**关键路径覆盖**：

1. **上传流程**：拖拽上传 → 图片排序 → 删除图片
2. **生成流程**：点击生成 → 流式输出 → 预览渲染
3. **聊天流程**：发送消息 → 代码更新 → 预览刷新

## 4. 文档架构

### README 结构

```markdown
# Multimodal LLM Frontend Generator

## Features
## Quick Start
## Development
## API Reference
## Architecture
## Contributing
```

### API 文档

使用 OpenAPI 3.0 规范，生成交互式文档：

```yaml
# api/openapi.yaml
openapi: 3.0.3
info:
  title: Frontend Generator API
  version: 1.0.0
paths:
  /api/upload:
    post: ...
  /api/generate:
    post: ...
  /api/chat:
    post: ...
```

## 决策记录

### ADR-1: 前端测试框架选择

**决策**：选择 Vitest 而非 Jest

**原因**：
1. Vitest 与 Vite 生态兼容性更好
2. 执行速度更快（原生 ESM）
3. 配置更简单
4. 与 Next.js 14+ 兼容良好

### ADR-2: E2E 测试框架选择

**决策**：选择 Playwright 而非 Cypress

**原因**：
1. 支持多浏览器并行测试
2. 原生支持 TypeScript
3. 更好的 CI/CD 集成
4. 测试执行速度更快

### ADR-3: 图片压缩时机

**决策**：在前端上传前压缩，后端保留验证

**原因**：
1. 减少网络传输时间
2. 降低服务器负载
3. 用户体验更好（实时反馈压缩进度）
4. 后端仍验证格式和大小作为安全保障
