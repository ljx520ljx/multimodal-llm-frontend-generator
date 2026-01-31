# commit-as-prompt

生成结构化的 commit message，包含 WHAT、WHY、HOW 描述，供 AI 查询和理解。

## 执行步骤

1. 运行 `git status` 和 `git diff --staged` 查看当前变更
2. 如果没有暂存的变更，先运行 `git add .` 暂存所有变更
3. 分析变更内容，生成结构化的 commit message：

```
WHAT: [一句话描述做了什么]

WHY: [为什么要做这个改动，背景和原因]

HOW:
- [关键实现点1]
- [关键实现点2]
- [关键实现点3]
```

4. 使用生成的 message 执行 `git commit`
5. 显示提交结果

## 注意事项

- WHAT 应该简洁明了，一句话概括
- WHY 解释业务背景和技术原因
- HOW 列出关键的实现细节，便于后续 AI 查询理解
- 不要包含敏感信息（密钥、密码等）
- 遵循项目现有的 commit message 风格

## 示例输出

```
WHAT: 初始化 vlogs-receiver 项目骨架

WHY: 作为日志接收服务，需要建立基础项目结构，包括 Go 模块配置、HTTP 服务入口和环境变量管理

HOW:
- 使用 Hertz 框架初始化 HTTP 服务器，支持优雅关闭
- 实现 /ping 健康检查端点，返回 {"status": "ok"}
- 通过 godotenv 加载环境变量，校验必填配置
- 创建 DocOps 三级文档体系和 Skills 技能定义
```
