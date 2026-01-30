# API 集成测试 Skill

> 版本: 1.0.0 | 最后优化: 初始版本

## 触发条件

当用户需要以下场景时激活：
- 开发或修改 API 端点
- 验证 API 功能正确性
- 测试与外部服务的集成

## 能力定义

### 1. 端点测试
验证 API HTTP 请求/响应的正确性：
- 请求参数验证
- 响应格式检查
- 状态码验证
- 错误处理测试

### 2. 认证测试
验证安全机制：
- Token 生成与验证
- JWT 签名与过期
- 权限控制（RBAC）
- 会话管理

### 3. 持久化测试
验证数据存储：
- 数据库 CRUD 操作
- 数据一致性
- 事务处理
- 并发写入

### 4. 集成测试
验证外部服务集成：
- Redis 缓存操作
- MongoDB 文档存储
- 消息队列交互
- 第三方 API 调用

### 5. 测试脚本生成
自动生成可复用的测试：
- curl 命令
- HTTP 文件（VS Code REST Client）
- Go/TypeScript 测试代码

## 执行流程

```
1. API 分析
   ├── 解析 API 定义（OpenAPI/路由）
   ├── 识别端点列表
   └── 提取请求/响应模型

2. 测试用例生成
   ├── 正常路径测试
   ├── 边界条件测试
   ├── 错误处理测试
   └── 安全测试

3. 测试执行
   ├── 发送请求
   ├── 验证响应
   ├── 检查副作用
   └── 记录结果

4. 报告生成
   ├── 覆盖率统计
   ├── 失败用例详情
   └── 性能指标
```

## 测试模板

### HTTP 请求测试

```http
### {test_name}
# @name {request_name}
{METHOD} {{baseUrl}}/{endpoint}
Content-Type: application/json
Authorization: Bearer {{token}}

{request_body}

### 预期响应
# Status: {expected_status}
# Body: {expected_body_pattern}
```

### Go 测试代码

```go
func Test{EndpointName}(t *testing.T) {
    // Arrange
    req := httptest.NewRequest("{METHOD}", "/{endpoint}", {body})
    rec := httptest.NewRecorder()

    // Act
    handler.ServeHTTP(rec, req)

    // Assert
    assert.Equal(t, {expected_status}, rec.Code)
    // 验证响应体
}
```

## 测试检查清单

### 功能测试
- [ ] 正常请求返回正确响应
- [ ] 参数验证生效
- [ ] 错误情况正确处理
- [ ] 分页/过滤/排序正常

### 安全测试
- [ ] 未认证请求被拒绝
- [ ] Token 过期处理正确
- [ ] 权限不足返回 403
- [ ] 无 SQL 注入风险

### 性能测试
- [ ] 响应时间 < 阈值
- [ ] 并发请求处理正常
- [ ] 无资源泄漏

## 输出格式

```markdown
## API 测试报告

### 测试概览
- **测试时间**: {timestamp}
- **端点数量**: {endpoint_count}
- **用例总数**: {total_cases}
- **通过率**: {pass_rate}%

### 端点覆盖

| 端点 | 方法 | 测试数 | 通过 | 失败 |
|------|------|--------|------|------|
| {endpoint} | {method} | {total} | {pass} | {fail} |

### 失败用例详情

| 用例 | 端点 | 预期 | 实际 | 原因 |
|------|------|------|------|------|
| {case} | {endpoint} | {expected} | {actual} | {reason} |

### 生成的测试脚本
- `tests/api/{endpoint}.http`
- `tests/api/{endpoint}_test.go`

### 建议
1. {recommendation_1}
2. {recommendation_2}
```

## 自我优化记录

| 日期 | 版本 | 优化内容 | 触发原因 |
|------|------|----------|----------|
| - | 1.0.0 | 初始版本 | 项目初始化 |

## 优化触发条件

当出现以下情况时，应更新此 Skill：
- 新增测试类型
- 发现测试盲区
- 测试脚本模板优化
- 集成新的外部服务
