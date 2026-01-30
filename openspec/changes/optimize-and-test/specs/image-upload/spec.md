# image-upload Spec Delta

## MODIFIED Requirements

### Requirement: Frontend Image Compression

系统 **SHALL** 在前端上传图片前进行压缩，以减少网络传输时间和服务器负载。

#### Scenario: 压缩大于 2MB 的图片

**Given** 用户选择上传一张 5MB 的 PNG 图片
**When** 图片被添加到上传列表
**Then** 系统自动压缩图片至 2MB 以下
**And** 压缩后的图片格式为 WebP
**And** 显示压缩进度指示器

#### Scenario: 小图片不压缩

**Given** 用户选择上传一张 1MB 的 JPEG 图片
**When** 图片被添加到上传列表
**Then** 图片保持原始大小不压缩
**And** 不显示压缩进度指示器

#### Scenario: 压缩过程中显示进度

**Given** 用户选择上传一张 8MB 的图片
**When** 压缩过程进行中
**Then** 显示压缩进度百分比
**And** 用户可以取消上传

#### Scenario: 压缩失败处理

**Given** 图片压缩过程中发生错误
**When** 压缩失败
**Then** 显示错误提示信息
**And** 提供"使用原图上传"选项
**And** 提供"取消上传"选项

### Requirement: Image Compression Configuration

系统 **SHALL** 提供图片压缩配置选项。

#### Scenario: 默认压缩配置

**Given** 用户未自定义压缩配置
**When** 压缩图片
**Then** 使用默认配置：
  - 目标大小: 2MB
  - 最大尺寸: 2048px
  - 质量: 80%
  - 输出格式: WebP

#### Scenario: 自定义压缩质量

**Given** 系统配置了自定义压缩质量为 70%
**When** 压缩图片
**Then** 使用 70% 质量进行压缩
