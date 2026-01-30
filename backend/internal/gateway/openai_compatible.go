package gateway

import (
	"context"
	"errors"
	"io"

	"multimodal-llm-frontend-generator/internal/gateway/types"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAICompatibleConfig holds configuration for OpenAI-compatible gateways
type OpenAICompatibleConfig struct {
	APIKey   string
	Model    string
	BaseURL  string
	Provider string
}

// OpenAICompatibleGateway implements LLMGateway for OpenAI-compatible APIs
// This includes OpenAI, DeepSeek, and Doubao (via Volcengine)
type OpenAICompatibleGateway struct {
	client   *openai.Client
	model    string
	provider string
}

// NewOpenAICompatibleGateway creates a new OpenAI-compatible gateway
func NewOpenAICompatibleGateway(cfg OpenAICompatibleConfig) (*OpenAICompatibleGateway, error) {
	if cfg.APIKey == "" {
		return nil, types.ErrInvalidAPIKey
	}
	if cfg.Model == "" {
		return nil, types.ErrInvalidModel
	}

	config := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		config.BaseURL = cfg.BaseURL
	}

	return &OpenAICompatibleGateway{
		client:   openai.NewClientWithConfig(config),
		model:    cfg.Model,
		provider: cfg.Provider,
	}, nil
}

// Provider returns the provider identifier
func (g *OpenAICompatibleGateway) Provider() string {
	return g.provider
}

// ChatStream sends a chat request and returns a channel for streaming responses
func (g *OpenAICompatibleGateway) ChatStream(ctx context.Context, req *types.ChatRequest) (<-chan types.StreamChunk, error) {
	openaiReq := g.convertRequest(req)

	stream, err := g.client.CreateChatCompletionStream(ctx, openaiReq)
	if err != nil {
		return nil, g.wrapError(err)
	}

	ch := make(chan types.StreamChunk)

	go func() {
		defer close(ch)
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				ch <- types.NewDoneChunk()
				return
			}
			if err != nil {
				select {
				case ch <- types.NewErrorChunk(g.wrapError(err)):
				case <-ctx.Done():
				}
				return
			}

			if len(response.Choices) > 0 {
				content := response.Choices[0].Delta.Content
				if content != "" {
					select {
					case ch <- types.NewContentChunk(content):
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return ch, nil
}

// convertRequest converts our ChatRequest to OpenAI's format
func (g *OpenAICompatibleGateway) convertRequest(req *types.ChatRequest) openai.ChatCompletionRequest {
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))

	for i, msg := range req.Messages {
		if len(msg.Content) == 1 && msg.Content[0].Type == types.ContentTypeText {
			// Simple text message
			messages[i] = openai.ChatCompletionMessage{
				Role:    msg.Role,
				Content: msg.Content[0].Text,
			}
		} else {
			// Multi-part message (with images)
			parts := make([]openai.ChatMessagePart, len(msg.Content))
			for j, part := range msg.Content {
				switch part.Type {
				case types.ContentTypeText:
					parts[j] = openai.ChatMessagePart{
						Type: openai.ChatMessagePartTypeText,
						Text: part.Text,
					}
				case types.ContentTypeImageURL:
					if part.ImageURL != nil {
						parts[j] = openai.ChatMessagePart{
							Type: openai.ChatMessagePartTypeImageURL,
							ImageURL: &openai.ChatMessageImageURL{
								URL:    part.ImageURL.URL,
								Detail: openai.ImageURLDetail(part.ImageURL.Detail),
							},
						}
					}
				}
			}
			messages[i] = openai.ChatCompletionMessage{
				Role:         msg.Role,
				MultiContent: parts,
			}
		}
	}

	model := req.Model
	if model == "" {
		model = g.model
	}

	openaiReq := openai.ChatCompletionRequest{
		Model:    model,
		Messages: messages,
		Stream:   true,
	}

	if req.Options != nil {
		if req.Options.Temperature > 0 {
			openaiReq.Temperature = req.Options.Temperature
		}
		if req.Options.MaxTokens > 0 {
			openaiReq.MaxTokens = req.Options.MaxTokens
		}
		if req.Options.TopP > 0 {
			openaiReq.TopP = req.Options.TopP
		}
		if req.Options.FrequencyPenalty != 0 {
			openaiReq.FrequencyPenalty = req.Options.FrequencyPenalty
		}
		if req.Options.PresencePenalty != 0 {
			openaiReq.PresencePenalty = req.Options.PresencePenalty
		}
	}

	return openaiReq
}

// wrapError converts OpenAI errors to our error types
func (g *OpenAICompatibleGateway) wrapError(err error) error {
	if err == nil {
		return nil
	}

	var apiErr *openai.APIError
	if errors.As(err, &apiErr) {
		retryable := types.IsRetryableStatusCode(apiErr.HTTPStatusCode)
		return types.NewGatewayError(
			g.provider,
			apiErr.HTTPStatusCode,
			apiErr.Message,
			retryable,
			err,
		)
	}

	// Check for context errors
	if errors.Is(err, context.Canceled) {
		return types.NewGatewayError(g.provider, 0, "request canceled", false, types.ErrContextCanceled)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return types.NewGatewayError(g.provider, 0, "request timeout", true, types.ErrTimeout)
	}

	// Generic error
	return types.NewGatewayError(g.provider, 0, err.Error(), false, err)
}
