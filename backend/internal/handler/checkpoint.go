package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"multimodal-llm-frontend-generator/internal/service"

	"github.com/gin-gonic/gin"
)

// DesignStatePersister extends SessionStore with checkpoint capabilities
type DesignStatePersister interface {
	SaveDesignState(ctx context.Context, sessionID string, stateJSON []byte) error
	GetDesignState(ctx context.Context, sessionID string) ([]byte, error)
}

// CheckpointHandler handles internal checkpoint API for Python Agent
type CheckpointHandler struct {
	sessionStore service.SessionStore
}

// NewCheckpointHandler creates a new checkpoint handler
func NewCheckpointHandler(store service.SessionStore) *CheckpointHandler {
	return &CheckpointHandler{sessionStore: store}
}

// SaveCheckpoint saves the design state for a session
// PUT /api/internal/checkpoint/:session_id
func (h *CheckpointHandler) SaveCheckpoint(c *gin.Context) {
	sessionID := c.Param("session_id")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	// Validate that body is valid JSON before saving
	if !json.Valid(body) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON in request body"})
		return
	}

	persister, ok := h.sessionStore.(DesignStatePersister)
	if !ok {
		// MemoryStore fallback - checkpoint not supported
		c.JSON(http.StatusOK, gin.H{"status": "skipped", "reason": "checkpoint not supported in memory mode"})
		return
	}

	if err := persister.SaveDesignState(c.Request.Context(), sessionID, body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GetCheckpoint retrieves the design state for a session
// GET /api/internal/checkpoint/:session_id
func (h *CheckpointHandler) GetCheckpoint(c *gin.Context) {
	sessionID := c.Param("session_id")

	persister, ok := h.sessionStore.(DesignStatePersister)
	if !ok {
		c.JSON(http.StatusOK, gin.H{"design_state": nil})
		return
	}

	stateJSON, err := persister.GetDesignState(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if stateJSON == nil {
		c.JSON(http.StatusOK, gin.H{"design_state": nil})
		return
	}

	// Validate stateJSON is valid JSON before wrapping
	if !json.Valid(stateJSON) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "corrupted design state"})
		return
	}

	// Use json.RawMessage to safely embed the state JSON
	c.JSON(http.StatusOK, gin.H{"design_state": json.RawMessage(stateJSON)})
}
