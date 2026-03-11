package handler

import (
	"net/http"

	"multimodal-llm-frontend-generator/internal/service"

	"github.com/gin-gonic/gin"
)

// PreviewHandler handles share and preview endpoints
type PreviewHandler struct {
	previewService *service.PreviewService
}

// NewPreviewHandler creates a new preview handler
func NewPreviewHandler(previewService *service.PreviewService) *PreviewHandler {
	return &PreviewHandler{previewService: previewService}
}

type createShareRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

type shareResponse struct {
	ShortCode string `json:"short_code"`
	URL       string `json:"url"`
}

// CreateShare creates a new shareable link
// POST /api/share
func (h *PreviewHandler) CreateShare(c *gin.Context) {
	var req createShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	preview, err := h.previewService.CreateShare(c.Request.Context(), req.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build URL using the request host
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if fwd := c.GetHeader("X-Forwarded-Proto"); fwd != "" {
		scheme = fwd
	}
	url := scheme + "://" + c.Request.Host + "/p/" + preview.ShortCode

	c.JSON(http.StatusOK, shareResponse{
		ShortCode: preview.ShortCode,
		URL:       url,
	})
}

// ServePreview serves the shared HTML preview
// GET /p/:code
func (h *PreviewHandler) ServePreview(c *gin.Context) {
	code := c.Param("code")

	preview, err := h.previewService.GetByShortCode(c.Request.Context(), code)
	if err != nil {
		c.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte(`
<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>Not Found</title></head>
<body style="display:flex;align-items:center;justify-content:center;height:100vh;font-family:sans-serif;">
<div style="text-align:center;">
<h1 style="font-size:2em;margin-bottom:0.5em;">404</h1>
<p>该分享链接不存在或已过期</p>
</div></body></html>`))
		return
	}

	// CSP: 允许生成的 HTML 所需的 CDN（Tailwind CSS + Alpine.js）
	// Alpine.js requires 'unsafe-eval' for parsing x-data/x-show expressions
	c.Header("Content-Security-Policy", "default-src 'self'; "+
		"script-src 'unsafe-inline' 'unsafe-eval' https://cdn.tailwindcss.com https://cdn.jsdelivr.net https://unpkg.com; "+
		"style-src 'unsafe-inline' https://cdn.tailwindcss.com https://cdn.jsdelivr.net https://unpkg.com; "+
		"img-src 'self' data: https:; "+
		"font-src 'self' https://fonts.gstatic.com https://cdn.jsdelivr.net;")
	c.Header("X-Content-Type-Options", "nosniff")
	// SAMEORIGIN: 允许同域 iframe 嵌入预览，阻止跨域嵌入
	c.Header("X-Frame-Options", "SAMEORIGIN")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(preview.HTMLSnapshot))
}

type updateShareRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

// UpdateShare updates the HTML snapshot for an existing share
// PUT /api/share/:code
func (h *PreviewHandler) UpdateShare(c *gin.Context) {
	code := c.Param("code")

	var req updateShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	if err := h.previewService.UpdateShare(c.Request.Context(), req.SessionID, code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// DeleteShare deactivates a share link
// DELETE /api/share/:code
func (h *PreviewHandler) DeleteShare(c *gin.Context) {
	code := c.Param("code")

	if err := h.previewService.DeleteShare(c.Request.Context(), code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
