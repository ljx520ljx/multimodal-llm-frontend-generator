# Tasks: Add ChatAgent with Tool Calling

## 1. Tools Implementation ✅

- [x] 1.1 创建 `tools/html_validator.py`
  - 实现 `validate_html()` 工具函数
  - 使用 `@tool` 装饰器
  - 复用 CodeValidator.validate() 逻辑
- [x] 1.2 创建 `tools/interaction_checker.py`
  - 实现 `check_interaction()` 工具函数
  - 检查状态定义完整性
  - 检查转换可达性
- [x] 1.3 更新 `tools/__init__.py`
  - 导出 validate_html, check_interaction

## 2. Schemas Update ✅

- [x] 2.1 更新 `schemas/common.py`
  - 添加 `SSEEventType.TOOL_CALL`
  - 添加 `SSEEventType.TOOL_RESULT`
- [x] 2.2 创建 `schemas/chat.py`
  - ChatRequest 模型
  - ChatMessage 模型（历史记录）

## 3. Prompt Template ✅

- [x] 3.1 创建 `llm/prompts/chat.py`
  - CHAT_MODIFY_PROMPT 模板
  - CHAT_SYSTEM_PROMPT 模板
  - 包含当前代码、用户指令、工具说明
- [x] 3.2 更新 `llm/prompts/__init__.py`
  - 导出 CHAT_MODIFY_PROMPT, CHAT_SYSTEM_PROMPT

## 4. ChatAgent ✅

- [x] 4.1 创建 `agents/chat_agent.py`
  - ChatAgent 类
  - 支持 Tool Calling
  - 支持原图上下文
  - 支持标记修改消息
- [x] 4.2 实现 `run()` 方法（多轮工具调用循环）
  - 构建 prompt
  - 处理历史记录
  - 流式输出 SSE 事件
  - **多轮循环**：LLM 自主判定是否继续
  - MAX_TOOL_ITERATIONS = 5 防止无限循环
- [x] 4.3 实现工具执行逻辑
  - `_execute_tool()` 方法
  - 工具结果追加到 messages
  - 继续调用 LLM 让其根据结果决定下一步
- [x] 4.4 实现消息格式转换
  - AssistantMessage (包含 tool_calls)
  - ToolResultMessage (工具执行结果)

## 5. LLM Gateway Extension ✅

- [x] 5.1 使用 LangChain bind_tools 实现
  - ChatAgent 中直接使用 `self.llm.client.bind_tools(self.tools)`
  - 无需单独的 gateway 方法

## 6. API Route ✅

- [x] 6.1 创建 `app/routes/chat.py`
  - ChatRequest 验证
  - ChatAgent 调用
  - SSE 流式响应
- [x] 6.2 更新 `app/main.py`
  - 注册 `/api/v1/chat` 路由

## 7. Go Backend ✅

- [x] 7.1 `internal/handler/chat.go` 已存在
  - ChatHandler 结构
  - 请求验证
  - SSE 转发
- [x] 7.2 更新 `internal/service/agent_client.go`
  - 添加 `Chat()` 方法
  - 添加 `AgentChatRequest` 类型
  - 添加 `SSETypeToolCall`, `SSETypeToolResult` 事件类型
- [x] 7.3 更新 `internal/service/generate_service.go`
  - 添加 `NewGenerateServiceWithAgent()` 构造函数
  - `Chat()` 方法现在通过 Python Agent 实现
  - 添加 `chatViaAgent()` 方法
- [x] 7.4 更新 `internal/app/app.go`
  - 使用 `NewGenerateServiceWithAgent()` 构造函数

## 8. Testing ✅

- [x] 8.1 工具单元测试
  - test_chat_agent.py 包含 TestToolFunctions 类
  - validate_html, check_interaction 功能测试
- [x] 8.2 ChatAgent 测试
  - test_chat_agent.py (26 个测试)
  - Mock LLM 响应
  - 多轮工具调用测试
  - 超时测试
  - 错误处理测试
- [x] 8.3 API 集成测试
  - 端到端测试通过
  - SSE 流验证
- [x] 8.4 部署配置
  - docker-compose.yml 创建
  - .env.example 环境变量文档
  - README.md 更新
