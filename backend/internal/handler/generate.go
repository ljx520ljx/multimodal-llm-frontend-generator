package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"multimodal-llm-frontend-generator/internal/service"

	"github.com/gin-gonic/gin"
)

// GenerateHandler handles code generation requests
type GenerateHandler struct {
	generateService service.GenerateService
	handlerTimeout  time.Duration
}

// NewGenerateHandler creates a new GenerateHandler
func NewGenerateHandler(generateService service.GenerateService, handlerTimeout time.Duration) *GenerateHandler {
	return &GenerateHandler{
		generateService: generateService,
		handlerTimeout:  handlerTimeout,
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

	// Apply handler-level timeout (sits between Agent timeout and Frontend SSE timeout)
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.handlerTimeout)
	defer cancel()

	// Start generation
	eventChan, err := h.generateService.Generate(ctx, req.SessionID, req.ImageIDs, req.Framework)
	if err != nil {
		handleError(c, err)
		return
	}

	// Stream SSE response
	streamSSE(c, eventChan)
}

// streamSSE streams SSE events to the client.
// Uses the same payload format as streamAgentSSE for consistency:
//
//	{"type": "<event_type>", "content": "<content>", "agent": "<agent>"}
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

			// Determine event type for SSE
			eventType := event.Type
			if eventType == "" {
				eventType = "message"
			}

			// Build payload with explicit fields (same format as streamAgentSSE)
			payload := map[string]any{
				"type":    event.Type,
				"content": event.Content,
			}
			if event.Agent != "" {
				payload["agent"] = event.Agent
			}
			data, err := json.Marshal(payload)
			if err != nil {
				continue
			}

			_, err = io.WriteString(c.Writer, "event: "+eventType+"\n")
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
