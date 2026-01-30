# Tasks: 优化多图交互分析

## 1. Prompt 模板重构

- [x] 1.1 创建 `backend/pkg/prompt/templates_v2.go`
  - 新增 `SystemPromptV2` 优化版系统 Prompt
  - 新增 `MultiImageAnalysisPromptV2` 多图分析 Prompt
  - 新增 `DiffAnalysisPromptV2` 差异分析 Prompt
  - 验证: 单元测试确保模板字符串格式正确

- [x] 1.2 创建 `backend/pkg/prompt/interaction_types.go`
  - 定义 `InteractionType` 枚举（6 种类型）
  - 定义 `InteractionPattern` 结构体
  - 定义 `InteractionPatterns` 映射表
  - 验证: 单元测试确保类型定义完整

- [x] 1.3 创建 `backend/pkg/prompt/few_shot.go`
  - 设计 1-2 个高质量 Few-Shot 示例
  - 示例包含：输入图片描述 → 分析过程 → 输出代码
  - 验证: 人工 review 示例质量

## 2. Prompt 构建服务

- [x] 2.1 创建 `backend/internal/service/prompt_strategy.go`
  - 实现 `PromptBuilder` 直接使用优化版模板
  - 验证: 单元测试确保 Prompt 构建正常

- [x] 2.2 修改 `backend/internal/service/prompt_service.go`
  - 使用 PromptBuilder 构建 Prompt
  - 保持现有方法签名不变（向后兼容）
  - 验证: 现有测试用例仍然通过

## 3. 配置支持

- [x] 3.1 修改 `backend/internal/config/config.go`
  - 添加 `ENABLE_FEW_SHOT` 配置项（默认 false）
  - 验证: 配置正确加载

## 4. 生成服务优化

- [x] 4.1 修改 `backend/internal/app/app.go`
  - 集成 Prompt 配置
  - 启动时打印 Few-Shot 配置
  - 验证: 服务启动正常

## 5. 测试与验证

- [x] 5.1 编写单元测试
  - `prompt_strategy_test.go` 测试 PromptBuilder
  - `templates_v2_test.go` 测试模板构建
  - `interaction_types_test.go` 测试交互类型
  - 验证: 所有测试通过，覆盖率 80%+

- [x] 5.2 编写集成测试
  - 验证 Go 代码编译成功
  - 验证现有测试继续通过

- [x] 5.3 手动验收测试
  - 需要部署后使用真实 UI 设计稿测试
  - 验证 6 种交互类型的识别准确性
  - 记录失败案例用于后续优化

## 6. 文档更新

- [x] 6.1 创建 `backend/pkg/prompt/CLAUDE.md`
  - 补充模板设计说明
  - 说明交互类型分类

- [x] 6.2 更新 `backend/internal/service/CLAUDE.md`
  - 补充 PromptBuilder 说明

## Implementation Summary

### Current Files
- `backend/pkg/prompt/templates_v2.go` - 优化版 Prompt 模板 + GetFrameworkDisplayName
- `backend/pkg/prompt/interaction_types.go` - 6 种交互类型定义
- `backend/pkg/prompt/few_shot.go` - Few-Shot 示例（toggle_modal, tab_switch）
- `backend/internal/service/prompt_strategy.go` - PromptBuilder 实现
- `backend/pkg/prompt/CLAUDE.md` - Prompt 模块文档

### Test Files
- `backend/pkg/prompt/templates_v2_test.go`
- `backend/pkg/prompt/interaction_types_test.go`
- `backend/pkg/prompt/few_shot_test.go`
- `backend/internal/service/prompt_strategy_test.go`

### Configuration
```bash
# 启用 Few-Shot 示例（可选）
ENABLE_FEW_SHOT=true
```

### Architecture
```
PromptService
    └── PromptBuilder
        └── Templates (templates_v2.go)
            ├── SystemPromptV2
            ├── SingleImagePromptV2
            ├── MultiImagePromptV2
            ├── DiffAnalysisPromptV2
            └── ChatModifyPromptV2
```
