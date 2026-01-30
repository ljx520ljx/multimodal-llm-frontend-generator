# Capability: 图片上传

## ADDED Requirements

### Requirement: 多图上传

系统 **SHALL** 支持用户同时上传多张 UI 设计稿图片，进行验证、压缩和存储。

#### Scenario: 上传单张有效图片

**Given** 用户准备上传一张 PNG 格式的图片（小于 10MB）
**When** 用户发送 POST /api/upload 请求，包含图片文件
**Then** 系统返回 200 OK
**And** 响应包含 session_id 和图片信息列表
**And** 图片被压缩并转换为 Base64 存储

#### Scenario: 上传多张图片

**Given** 用户准备上传 3 张 JPEG 格式的图片
**When** 用户发送 POST /api/upload 请求，包含所有图片文件
**Then** 系统返回 200 OK
**And** 响应包含 session_id
**And** 响应的 images 数组包含 3 个元素，每个有 id, filename, order

#### Scenario: 上传不支持的图片格式

**Given** 用户准备上传一张 GIF 格式的图片
**When** 用户发送 POST /api/upload 请求
**Then** 系统返回 400 Bad Request
**And** 响应包含错误信息 "Invalid image format"

#### Scenario: 上传过大的图片

**Given** 用户准备上传一张 15MB 的图片（超过 10MB 限制）
**When** 用户发送 POST /api/upload 请求
**Then** 系统返回 400 Bad Request
**And** 响应包含错误信息 "Image too large"

### Requirement: 图片处理

系统 **MUST** 对上传的图片进行预处理，包括压缩、缩放和格式转换。

#### Scenario: 压缩大尺寸图片

**Given** 用户上传一张 4000x3000 像素的图片
**When** 图片处理服务处理该图片
**Then** 图片被缩放至最大 2048 像素（保持宽高比）
**And** 图片被压缩至 80% 质量
**And** 图片被转换为 Base64 data URL 格式

#### Scenario: 保持小图片原尺寸

**Given** 用户上传一张 800x600 像素的图片
**When** 图片处理服务处理该图片
**Then** 图片尺寸保持不变
**And** 图片被压缩至 80% 质量

### Requirement: 会话管理

系统 **SHALL** 为每次上传创建或复用会话，存储图片和相关数据。

#### Scenario: 首次上传创建新会话

**Given** 用户首次访问系统
**When** 用户上传图片
**Then** 系统创建新的会话
**And** 返回新生成的 session_id (UUID 格式)

#### Scenario: 会话自动过期

**Given** 一个会话创建后超过 30 分钟未被访问
**When** 后台清理任务运行
**Then** 该会话及其关联的图片数据被删除
