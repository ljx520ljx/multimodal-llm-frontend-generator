package service

import (
	"fmt"
	"time"
)

// Session represents a user session containing uploaded images and generated code
type Session struct {
	ID        string         // UUID
	Images    []ImageData    // Uploaded images
	Code      string         // Latest generated code
	Framework string         // Target framework (react/vue)
	History   []HistoryEntry // Conversation history
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ImageData represents a processed image
type ImageData struct {
	ID       string // Image ID (UUID)
	Filename string // Original filename
	MimeType string // MIME type (image/png, image/jpeg, image/webp)
	Base64   string // Base64 encoded data (data:image/...;base64,...)
	Order    int    // Sort order
}

// HistoryEntry represents a single conversation entry
type HistoryEntry struct {
	Role    string // user | assistant
	Content string // Message content
	Type    string // text | code
}

// SSEEvent represents a Server-Sent Event
type SSEEvent struct {
	Type    string `json:"type"`              // thinking | code | error | done | agent_start | agent_result
	Content string `json:"content"`           // Event content
	Agent   string `json:"agent,omitempty"`   // Agent name (for agent_start/agent_result)
}

// SSE event types
const (
	SSETypeThinking    = "thinking"
	SSETypeCode        = "code"
	SSETypeError       = "error"
	SSETypeDone        = "done"
	SSETypeToolCall    = "tool_call"
	SSETypeToolResult  = "tool_result"
	SSETypeAgentStart  = "agent_start"
	SSETypeAgentResult = "agent_result"
)

// Agent error types

// ErrAgentRateLimited is returned when the agent service returns 429
type ErrAgentRateLimited struct{}

func (e *ErrAgentRateLimited) Error() string {
	return "agent service rate limited"
}

// ErrAgentUnavailable is returned when the agent service returns 503
type ErrAgentUnavailable struct{}

func (e *ErrAgentUnavailable) Error() string {
	return "agent service unavailable"
}

// ErrAgentTimeout is returned when the agent service returns 504 or the request times out
type ErrAgentTimeout struct{}

func (e *ErrAgentTimeout) Error() string {
	return "agent service timeout"
}

// ErrAgentError is returned for other agent service errors
type ErrAgentError struct {
	StatusCode int
}

func (e *ErrAgentError) Error() string {
	return fmt.Sprintf("agent service error: status %d", e.StatusCode)
}

// Supported frameworks
const (
	FrameworkReact = "react"
	FrameworkVue   = "vue"
)

// ValidFramework checks if the framework is supported
func ValidFramework(framework string) bool {
	return framework == FrameworkReact || framework == FrameworkVue
}
