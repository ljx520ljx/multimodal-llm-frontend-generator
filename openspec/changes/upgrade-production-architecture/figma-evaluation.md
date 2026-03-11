# E.3 Figma API 集成评估

## 可行性评估

### Figma API 导出能力

Figma REST API (v1) 提供 `GET /v1/images/:file_key` 端点，支持：
- 导出指定 node 为 PNG/JPG/SVG/PDF
- 自定义 scale (1x-4x)
- 需要 Personal Access Token 或 OAuth2

### 集成方案

```
用户输入 Figma URL
  → 解析 file_key 和 node_id
  → Figma API /v1/files/:file_key/nodes?ids=:node_ids (获取页面列表)
  → Figma API /v1/images/:file_key?ids=:node_ids&format=png&scale=2
  → 下载 PNG 到内存
  → 复用现有 upload → generate Pipeline
```

### 限制

1. **认证要求**：需要用户提供 Personal Access Token 或 OAuth2 授权
2. **API 限制**：Free plan 有 rate limit (30 req/min)
3. **导出质量**：PNG 导出丢失矢量信息，但对我们的视觉分析 Pipeline 足够
4. **页面发现**：需要额外请求获取文件结构，找到所有页面/Frame
5. **实时同步**：Figma API 不支持 webhook 推送变更

### 结论

**可行但非 MVP 必需**。
- 技术上完全可行，约需 2-3 天开发
- 但需要用户配置 Token，增加使用门槛
- 建议作为 post-MVP 功能，在用户系统完善后（可以安全存储 Token）再实现
- 放入 roadmap: Phase F

### 未来实现路径

1. `backend/internal/service/figma_service.go` — URL 解析、API 调用、PNG 下载
2. `frontend/src/components/upload/FigmaImport.tsx` — URL 输入、Token 配置、页面选择
3. 复用现有 `api.upload()` + `api.generate()` 流程
