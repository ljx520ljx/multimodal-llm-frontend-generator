package handler

import (
	"log"
	"net/http"

	"multimodal-llm-frontend-generator/internal/service"

	"github.com/gin-gonic/gin"
)

// EchoHandler handles echo requests for testing Go ↔ Python communication
type EchoHandler struct {
	agentClient service.AgentClient
}

// NewEchoHandler creates a new EchoHandler
func NewEchoHandler(agentClient service.AgentClient) *EchoHandler {
	return &EchoHandler{
		agentClient: agentClient,
	}
}

// EchoRequest represents the echo request body
type EchoRequest struct {
	Message string  `json:"message" binding:"required"`
	Count   int     `json:"count"`
	Delay   float64 `json:"delay"`
}

// Handle handles POST /api/echo
func (h *EchoHandler) Handle(c *gin.Context) {
	var req EchoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err.Error())
		return
	}

	// Default values
	if req.Count == 0 {
		req.Count = 5
	}
	if req.Delay == 0 {
		req.Delay = 0.5
	}

	ctx := c.Request.Context()

	// Call Python agent
	eventChan, err := h.agentClient.Echo(ctx, &service.EchoRequest{
		Message: req.Message,
		Count:   req.Count,
		Delay:   req.Delay,
	})
	if err != nil {
		log.Printf("[Echo] Agent service error: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Agent service unavailable",
		})
		return
	}

	// Stream SSE response
	streamSSE(c, eventChan)
}

// HealthCheck handles GET /api/agent-health
func (h *EchoHandler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()

	if err := h.agentClient.Health(ctx); err != nil {
		log.Printf("[Echo] Agent health check failed: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "Agent service health check failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}
