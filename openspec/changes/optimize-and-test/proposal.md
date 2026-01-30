# Change Proposal: optimize-and-test

## Summary

Phase 8 优化与测试阶段，涵盖性能优化、测试覆盖率提升和文档完善。

## Motivation

当前项目已完成核心功能开发，但存在以下问题：

1. **性能问题**
   - 图片上传未进行前端压缩，大图直接传输浪费带宽
   - 前端未进行代码分割，首屏加载包含所有模块

2. **测试覆盖不足**
   - 前端无测试配置（缺少 Jest/Vitest）
   - 后端测试覆盖率分布不均：
     - `cmd/server`: 0%
     - `internal/app`: 0%
     - `internal/config`: 0%
     - `internal/gateway`: 44.5%
     - `internal/handler`: 79.6%
     - `internal/service`: 85.7%
   - 无 E2E 测试

3. **文档缺失**
   - README 未更新至最新状态
   - API 文档不完整

## Scope

### In Scope

1. **图片压缩优化** - 前端上传前压缩
2. **前端代码分割** - Next.js dynamic imports
3. **前端测试框架** - Vitest + React Testing Library
4. **后端测试补充** - 提升整体覆盖率至 60%+
5. **E2E 测试** - Playwright 关键路径测试
6. **文档更新** - README + API 文档

### Out of Scope

- 功能新增或重大重构
- 部署流程优化
- 监控/日志系统

## Affected Capabilities

| Capability | Change Type | Description |
|------------|-------------|-------------|
| image-upload | MODIFIED | 添加前端压缩需求 |
| testing | ADDED | 新增测试基础设施规范 |
| documentation | ADDED | 新增文档规范 |

## Dependencies

- 无外部依赖阻塞
- 可并行执行多个任务

## Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| 前端压缩影响图片质量 | Medium | Medium | 提供压缩质量配置，默认 80% |
| Vitest 与 Next.js 集成问题 | Low | Medium | 使用官方推荐配置 |
| E2E 测试不稳定 | Medium | Low | 添加重试机制和显式等待 |

## Success Criteria

- [ ] 图片上传前端压缩 > 2MB 图片
- [ ] 首屏 JS 包减少 30%+
- [ ] 前端测试覆盖率 > 60%
- [ ] 后端整体测试覆盖率 > 60%
- [ ] E2E 测试覆盖核心流程（上传 → 生成 → 预览）
- [ ] README 包含完整安装/使用说明
- [ ] API 文档覆盖所有端点
