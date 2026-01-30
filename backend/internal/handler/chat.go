package handler

import (
	"multimodal-llm-frontend-generator/internal/service"

	"github.com/gin-gonic/gin"
)

// ChatHandler handles chat-based code modification requests
type ChatHandler struct {
	generateService service.GenerateService
}

// NewChatHandler creates a new ChatHandler
func NewChatHandler(generateService service.GenerateService) *ChatHandler {
	return &ChatHandler{
		generateService: generateService,
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

	ctx := c.Request.Context()

	// Start chat
	eventChan, err := h.generateService.Chat(ctx, req.SessionID, req.Message)
	if err != nil {
		handleError(c, err)
		return
	}

	// Stream SSE response (reuse from generate handler)
	streamSSE(c, eventChan)
}
