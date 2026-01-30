package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"multimodal-llm-frontend-generator/internal/service"

	"github.com/gin-gonic/gin"
)

// GenerateHandler handles code generation requests
type GenerateHandler struct {
	generateService service.GenerateService
}

// NewGenerateHandler creates a new GenerateHandler
func NewGenerateHandler(generateService service.GenerateService) *GenerateHandler {
	return &GenerateHandler{
		generateService: generateService,
	}
}

// Handle handles POST /api/generate
func (h *GenerateHandler) Handle(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err.Error())
		return
	}

	// Validate framework
	if !service.ValidFramework(req.Framework) {
		handleValidationError(c, "Invalid framework. Use 'react' or 'vue'")
		return
	}

	ctx := c.Request.Context()

	// Start generation
	eventChan, err := h.generateService.Generate(ctx, req.SessionID, req.ImageIDs, req.Framework)
	if err != nil {
		handleError(c, err)
		return
	}

	// Stream SSE response
	streamSSE(c, eventChan)
}

// streamSSE streams SSE events to the client
func streamSSE(c *gin.Context, eventChan <-chan service.SSEEvent) {
	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering

	c.Status(http.StatusOK)

	// Flush headers
	c.Writer.Flush()

	ctx := c.Request.Context()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-eventChan:
			if !ok {
				return
			}

			// Marshal event to JSON
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}

			// Write SSE event
			_, err = io.WriteString(c.Writer, "event: message\n")
			if err != nil {
				return
			}
			_, err = io.WriteString(c.Writer, "data: ")
			if err != nil {
				return
			}
			_, err = c.Writer.Write(data)
			if err != nil {
				return
			}
			_, err = io.WriteString(c.Writer, "\n\n")
			if err != nil {
				return
			}

			c.Writer.Flush()
		}
	}
}
