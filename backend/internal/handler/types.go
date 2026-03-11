package handler

// UploadResponse is the response for image upload
type UploadResponse struct {
	SessionID string      `json:"session_id"`
	Images    []ImageInfo `json:"images"`
}

// ImageInfo contains basic image information
type ImageInfo struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Order    int    `json:"order"`
}

// GenerateRequest is the request for code generation
type GenerateRequest struct {
	SessionID string   `json:"session_id" binding:"required"`
	ImageIDs  []string `json:"image_ids" binding:"required,min=1"`
	Framework string   `json:"framework" binding:"required,oneof=react vue"`
}

// ChatRequest is the request for chat-based code modification
type ChatRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	Message   string `json:"message" binding:"required"`
}

// ErrorResponse is the standard error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error codes
const (
	ErrCodeInvalidImageFormat = "INVALID_IMAGE_FORMAT"
	ErrCodeImageTooLarge      = "IMAGE_TOO_LARGE"
	ErrCodeSessionNotFound    = "SESSION_NOT_FOUND"
	ErrCodeImageNotFound      = "IMAGE_NOT_FOUND"
	ErrCodeNoCodeGenerated    = "NO_CODE_GENERATED"
	ErrCodeGenerationFailed   = "GENERATION_FAILED"
	ErrCodeRateLimited        = "RATE_LIMITED"
	ErrCodeAgentUnavailable   = "AGENT_UNAVAILABLE"
	ErrCodeAgentTimeout       = "AGENT_TIMEOUT"
	ErrCodeAgentError         = "AGENT_ERROR"
	ErrCodeInvalidRequest     = "INVALID_REQUEST"
)
