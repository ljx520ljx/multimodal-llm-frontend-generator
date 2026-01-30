# image-upload Specification

## Purpose
TBD - created by archiving change add-backend-services. Update Purpose after archive.
## Requirements
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

### Requirement: Drag and Drop Upload

The system SHALL provide a drag-and-drop zone for uploading UI design images.

#### Scenario: User drags images into upload zone
- **WHEN** user drags one or more image files into the upload zone
- **THEN** the zone SHALL display a visual highlight indicating drop target
- **AND** upon drop, images SHALL be added to the upload list

#### Scenario: User clicks to select files
- **WHEN** user clicks on the upload zone
- **THEN** a file picker dialog SHALL open
- **AND** selected images SHALL be added to the upload list

#### Scenario: Invalid file type rejected
- **WHEN** user attempts to upload a non-image file
- **THEN** the system SHALL reject the file
- **AND** display an error message indicating valid formats (PNG, JPG, WebP)

### Requirement: Image Preview List

The system SHALL display uploaded images in a preview list with thumbnails.

#### Scenario: Images displayed as cards
- **WHEN** images are added to the upload list
- **THEN** each image SHALL be displayed as a card with thumbnail preview
- **AND** the card SHALL show the image filename

#### Scenario: Remove image from list
- **WHEN** user clicks the delete button on an image card
- **THEN** the image SHALL be removed from the upload list

### Requirement: Image Reordering

The system SHALL allow users to reorder uploaded images via drag-and-drop.

#### Scenario: Drag to reorder
- **WHEN** user drags an image card to a new position
- **THEN** the image order SHALL update accordingly
- **AND** visual feedback SHALL indicate the new position during drag

#### Scenario: Order persisted
- **WHEN** images are reordered
- **THEN** the new order SHALL be reflected when generating code

### Requirement: Upload to Backend

The system SHALL upload images to the backend API and obtain a session ID.

#### Scenario: Successful upload
- **WHEN** user triggers code generation with images in the list
- **THEN** images SHALL be uploaded to `POST /api/upload`
- **AND** the returned session_id SHALL be stored for subsequent API calls

#### Scenario: Upload failure handling
- **WHEN** the upload API returns an error
- **THEN** the system SHALL display an error message
- **AND** allow the user to retry

### Requirement: Backend Upload Integration

The system SHALL upload images to the backend API.

#### Scenario: Upload images to backend
- **GIVEN** the user has added images to the upload panel
- **WHEN** the user clicks "Generate Prototype"
- **THEN** images SHALL be uploaded to POST /api/upload
- **AND** the response imageIds SHALL be stored

#### Scenario: Upload progress display
- **WHEN** images are being uploaded
- **THEN** a progress indicator SHALL be displayed
- **AND** the upload status SHALL be updated in real-time

### Requirement: Upload Error Handling

The system SHALL handle upload failures gracefully.

#### Scenario: Network error during upload
- **WHEN** the upload request fails due to network issues
- **THEN** an error message SHALL be displayed
- **AND** a retry option SHALL be available

#### Scenario: File too large
- **GIVEN** an image file exceeds 10MB
- **WHEN** the user attempts to upload
- **THEN** the image SHALL be compressed before upload
- **OR** an error message SHALL inform the user about the size limit

