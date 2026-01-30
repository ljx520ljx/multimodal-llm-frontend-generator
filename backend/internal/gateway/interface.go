package gateway

import (
	"context"

	"multimodal-llm-frontend-generator/internal/gateway/types"
)

// LLMGateway defines the interface for LLM API interactions
type LLMGateway interface {
	// ChatStream sends a chat request and returns a channel for streaming responses
	ChatStream(ctx context.Context, req *types.ChatRequest) (<-chan types.StreamChunk, error)

	// Provider returns the provider identifier (e.g., "openai", "deepseek", "doubao")
	Provider() string
}
