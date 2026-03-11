package handler

import (
	"net/http"

	"multimodal-llm-frontend-generator/internal/service"

	"github.com/gin-gonic/gin"
)

// FeedbackHandler handles user feedback endpoints
type FeedbackHandler struct {
	feedbackService *service.FeedbackService
}

// NewFeedbackHandler creates a new feedback handler
func NewFeedbackHandler(fs *service.FeedbackService) *FeedbackHandler {
	return &FeedbackHandler{feedbackService: fs}
}

type submitFeedbackRequest struct {
	SessionID     string  `json:"session_id" binding:"required"`
	CodeVersionID *string `json:"code_version_id"`
	Rating        int     `json:"rating" binding:"required,min=1,max=5"`
	FeedbackText  *string `json:"feedback_text"`
	FeedbackType  string  `json:"feedback_type"`
}

// Submit records user feedback
// POST /api/feedback
func (h *FeedbackHandler) Submit(c *gin.Context) {
	var req submitFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: rating (1-5) and session_id are required"})
		return
	}

	// Extract user_id from JWT context if available
	var userID *string
	if uid, exists := c.Get("user_id"); exists {
		s := uid.(string)
		userID = &s
	}

	fb := &service.UserFeedback{
		SessionID:     req.SessionID,
		UserID:        userID,
		CodeVersionID: req.CodeVersionID,
		Rating:        req.Rating,
		FeedbackText:  req.FeedbackText,
		FeedbackType:  req.FeedbackType,
	}

	result, err := h.feedbackService.Submit(c.Request.Context(), fb)
	if err != nil {
		if err == service.ErrInvalidRating {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to submit feedback"})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// ListBySession returns all feedback for a session
// GET /api/sessions/:session_id/feedback
func (h *FeedbackHandler) ListBySession(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	feedbacks, err := h.feedbackService.ListBySession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list feedback"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"feedback": feedbacks})
}

// Stats returns aggregate feedback stats for a session
// GET /api/sessions/:session_id/feedback/stats
func (h *FeedbackHandler) Stats(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	stats, err := h.feedbackService.GetStats(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
