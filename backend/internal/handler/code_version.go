package handler

import (
	"net/http"

	"multimodal-llm-frontend-generator/internal/service"

	"github.com/gin-gonic/gin"
)

// CodeVersionHandler handles code version history endpoints
type CodeVersionHandler struct {
	codeVersionService *service.CodeVersionService
}

// NewCodeVersionHandler creates a new code version handler
func NewCodeVersionHandler(cvs *service.CodeVersionService) *CodeVersionHandler {
	return &CodeVersionHandler{codeVersionService: cvs}
}

// List returns all code versions for a session (without full HTML code)
// GET /api/sessions/:session_id/versions
func (h *CodeVersionHandler) List(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	versions, err := h.codeVersionService.ListBySession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list versions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"versions": versions})
}

// Get returns a specific code version with full HTML
// GET /api/versions/:id
func (h *CodeVersionHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "version id is required"})
		return
	}

	version, err := h.codeVersionService.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrCodeVersionNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get version"})
		return
	}

	c.JSON(http.StatusOK, version)
}
