# Change: 优化多图交互分析架构与 Prompt

## Why

当前多图交互分析功能已实现基础版本，但存在以下问题：

1. **Prompt 不够精准**：现有 Prompt 较为通用，缺乏针对交互意图推理的结构化引导
2. **架构不够模块化**：图片分析、差异检测、交互推理耦合在一起，难以迭代优化
3. **成本敏感**：项目依赖外部 LLM API，需要用更精准的 Prompt 弥补模型能力差距

参考 Claude Code 的架构思想，将复杂任务拆解为明确的分析步骤，通过精准的 Prompt 工程提升生成质量。

## What Changes

### 1. Prompt 工程优化

- **结构化分析框架**：将多图分析拆解为 5 个明确步骤
  - Step 1: 单图布局识别
  - Step 2: 组件层级分析
  - Step 3: 视觉差异检测
  - Step 4: 交互意图推理
  - Step 5: 状态机代码生成

- **Few-Shot 示例库**：植入高质量的 "UI 对比 → 交互代码" 示例

- **交互类型分类**：明确定义支持的交互模式
  - 点击切换状态 (Toggle)
  - 展开/收起 (Expand/Collapse)
  - Tab 切换 (Tab Switch)
  - 表单状态变化 (Form State)
  - 悬停效果 (Hover)
  - 导航跳转 (Navigation)

### 2. 多图分析架构

- **分阶段处理**：
  - Phase A: 预处理 - 图片标注序号、尺寸归一化
  - Phase B: 单图分析 - 独立分析每张图的 UI 结构
  - Phase C: 差异检测 - 对比相邻图片的变化区域
  - Phase D: 意图推理 - 推断交互类型和触发条件
  - Phase E: 代码生成 - 生成包含状态管理的组件代码

- **分析结果结构化**：定义中间表示 JSON Schema
  ```json
  {
    "frames": [...],
    "transitions": [{
      "from": 0, "to": 1,
      "changes": [...],
      "interaction": { "type": "click", "target": "..." }
    }],
    "states": [...],
    "events": [...]
  }
  ```

### 3. Prompt 模板重构

- **System Prompt**：优化角色设定和输出格式约束
- **Multi-Image Prompt**：重构多图分析指令
- **Diff Analysis Prompt**：专门的差异分析 Prompt
- **State Machine Prompt**：状态机代码生成 Prompt

### 4. 后端服务增强

- 新增 `PromptStrategy` 接口支持多种 Prompt 策略
- 新增 `InteractionAnalyzer` 服务
- 优化流式输出的分段处理

## Impact

- **Affected specs**:
  - `code-generation` - MODIFIED (Prompt 构建逻辑)
  - `prompt-engineering` - ADDED (新能力)
  - `multi-image-analysis` - ADDED (新能力)

- **Affected code**:
  - `backend/pkg/prompt/templates.go` - 重构
  - `backend/internal/service/prompt_service.go` - 增强
  - `backend/internal/service/generate_service.go` - 优化

- **Breaking changes**: 无 - 向后兼容

## Success Metrics

| 指标 | 当前值 | 目标值 |
|------|--------|--------|
| 交互逻辑正确率 | ~70% | ≥85% |
| 状态机代码生成成功率 | ~60% | ≥80% |
| 平均生成时间 | ~45s | ≤40s (减少重试) |
