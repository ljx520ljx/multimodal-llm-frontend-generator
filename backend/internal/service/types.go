package service

import "time"

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
	Type    string `json:"type"`    // thinking | code | error | done
	Content string `json:"content"` // Event content
}

// SSE event types
const (
	SSETypeThinking   = "thinking"
	SSETypeCode       = "code"
	SSETypeError      = "error"
	SSETypeDone       = "done"
	SSETypeToolCall   = "tool_call"
	SSETypeToolResult = "tool_result"
)

// Supported frameworks
const (
	FrameworkReact = "react"
	FrameworkVue   = "vue"
)

// ValidFramework checks if the framework is supported
func ValidFramework(framework string) bool {
	return framework == FrameworkReact || framework == FrameworkVue
}
