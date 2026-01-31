package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"multimodal-llm-frontend-generator/internal/service"

	"github.com/gin-gonic/gin"
)

// AgentGenerateRequest is the request for agent-based code generation
type AgentGenerateRequest struct {
	SessionID string   `json:"session_id" binding:"required"`
	ImageIDs  []string `json:"image_ids" binding:"required,min=1"`
}

// AgentGenerateHandler handles code generation via Python Agent
type AgentGenerateHandler struct {
	agentGenerateService service.AgentGenerateService
}

// NewAgentGenerateHandler creates a new AgentGenerateHandler
func NewAgentGenerateHandler(agentGenerateService service.AgentGenerateService) *AgentGenerateHandler {
	return &AgentGenerateHandler{
		agentGenerateService: agentGenerateService,
	}
}

// Handle handles POST /api/agent/generate
func (h *AgentGenerateHandler) Handle(c *gin.Context) {
	var req AgentGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err.Error())
		return
	}

	ctx := c.Request.Context()

	// Start generation via Python Agent
	eventChan, err := h.agentGenerateService.Generate(ctx, req.SessionID, req.ImageIDs)
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
			default:
				eventType = "message"
			}

			// Write SSE event with proper event type
			_, err := io.WriteString(c.Writer, "event: "+eventType+"\n")
			if err != nil {
				return
			}

			// Marshal data
			data, err := json.Marshal(map[string]interface{}{
				"type":    event.Type,
				"content": event.Content,
			})
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
