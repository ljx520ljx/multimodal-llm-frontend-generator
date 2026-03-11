package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"multimodal-llm-frontend-generator/internal/service"

	"github.com/gin-gonic/gin"
)

// AgentGenerateRequest is the request for agent-based code generation.
// Either image_ids or description must be provided.
type AgentGenerateRequest struct {
	SessionID   string   `json:"session_id" binding:"required"`
	ImageIDs    []string `json:"image_ids"`
	Description string   `json:"description"`
}

// AgentGenerateHandler handles code generation via Python Agent
type AgentGenerateHandler struct {
	agentGenerateService service.AgentGenerateService
	handlerTimeout       time.Duration
}

// NewAgentGenerateHandler creates a new AgentGenerateHandler
func NewAgentGenerateHandler(agentGenerateService service.AgentGenerateService, handlerTimeout time.Duration) *AgentGenerateHandler {
	return &AgentGenerateHandler{
		agentGenerateService: agentGenerateService,
		handlerTimeout:       handlerTimeout,
	}
}

// Handle handles POST /api/agent/generate
func (h *AgentGenerateHandler) Handle(c *gin.Context) {
	var req AgentGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err.Error())
		return
	}

	req.Description = strings.TrimSpace(req.Description)
	if len(req.ImageIDs) == 0 && req.Description == "" {
		handleValidationError(c, "Either image_ids or description is required")
		return
	}

	// Apply handler-level timeout (sits between Agent timeout and Frontend SSE timeout)
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.handlerTimeout)
	defer cancel()

	// Start generation via Python Agent
	eventChan, err := h.agentGenerateService.Generate(ctx, req.SessionID, req.ImageIDs, req.Description)
	if err != nil {
		handleError(c, err)
		return
	}

	// Stream SSE response
	streamAgentSSE(c, eventChan)
}

// streamAgentSSE streams SSE events to the client from Python Agent
func streamAgentSSE(c *gin.Context, eventChan <-chan service.SSEEvent) {
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
			var eventType string
			switch event.Type {
			case service.SSETypeDone:
				eventType = "done"
			case service.SSETypeError:
				eventType = "error"
			case service.SSETypeCode:
				eventType = "code"
			case service.SSETypeThinking:
				eventType = "thinking"
			case service.SSETypeAgentStart:
				eventType = "agent_start"
			case service.SSETypeAgentResult:
				eventType = "agent_result"
			default:
				eventType = "message"
			}

			// Write SSE event with proper event type
			_, err := io.WriteString(c.Writer, "event: "+eventType+"\n")
			if err != nil {
				return
			}

			// Marshal data (include agent field only when present)
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
