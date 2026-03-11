package handler

import (
	"context"
	"time"

	"multimodal-llm-frontend-generator/internal/service"

	"github.com/gin-gonic/gin"
)

// ChatHandler handles chat-based code modification requests
type ChatHandler struct {
	generateService service.GenerateService
	handlerTimeout  time.Duration
}

// NewChatHandler creates a new ChatHandler
func NewChatHandler(generateService service.GenerateService, handlerTimeout time.Duration) *ChatHandler {
	return &ChatHandler{
		generateService: generateService,
		handlerTimeout:  handlerTimeout,
	}
}

// Handle handles POST /api/chat
func (h *ChatHandler) Handle(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err.Error())
		return
	}

	if req.Message == "" {
		handleValidationError(c, "Message cannot be empty")
		return
	}

	// Apply handler-level timeout (sits between Agent timeout and Frontend SSE timeout)
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.handlerTimeout)
	defer cancel()

	// Start chat
	eventChan, err := h.generateService.Chat(ctx, req.SessionID, req.Message)
	if err != nil {
		handleError(c, err)
		return
	}

	// Stream SSE response (reuse from generate handler)
	streamSSE(c, eventChan)
}
