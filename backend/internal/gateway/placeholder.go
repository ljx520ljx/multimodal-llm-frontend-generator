package gateway

import (
	"context"
	"fmt"

	"multimodal-llm-frontend-generator/internal/gateway/types"
)

// placeholderGateway is a placeholder implementation for providers not yet implemented
type placeholderGateway struct {
	provider string
}

// Provider returns the provider identifier
func (g *placeholderGateway) Provider() string {
	return g.provider
}

// ChatStream returns an error indicating the provider is not implemented
func (g *placeholderGateway) ChatStream(ctx context.Context, req *types.ChatRequest) (<-chan types.StreamChunk, error) {
	return nil, fmt.Errorf("provider %s is not yet implemented", g.provider)
}
