package gateway

import (
	"fmt"

	"multimodal-llm-frontend-generator/internal/gateway/types"
)

// Provider constants
const (
	ProviderOpenAI    = "openai"
	ProviderDeepSeek  = "deepseek"
	ProviderDoubao    = "doubao"
	ProviderGemini    = "gemini"
	ProviderAnthropic = "anthropic"
)

// Default API endpoints
const (
	OpenAIBaseURL   = "https://api.openai.com/v1"
	DeepSeekBaseURL = "https://api.deepseek.com"
	DoubaoBaseURL   = "https://ark.cn-beijing.volces.com/api/v3"
)

// Default models
const (
	DefaultOpenAIModel   = "gpt-4o"
	DefaultDeepSeekModel = "deepseek-chat"
)

// GatewayConfig holds all configuration needed to create a gateway
type GatewayConfig struct {
	Provider string

	// OpenAI
	OpenAIAPIKey  string
	OpenAIModel   string
	OpenAIBaseURL string // Custom base URL for OpenAI-compatible APIs

	// DeepSeek
	DeepSeekAPIKey string
	DeepSeekModel  string

	// Doubao (Volcengine)
	DoubaoAPIKey string
	DoubaoModel  string // ep-xxx-xxx format

	// Gemini
	GeminiAPIKey string

	// Anthropic
	AnthropicAPIKey string

	// Retry config
	RetryConfig *RetryConfig
}

// NewGateway creates a new LLM gateway based on the configuration
func NewGateway(cfg GatewayConfig) (LLMGateway, error) {
	var gateway LLMGateway
	var err error

	switch cfg.Provider {
	case ProviderOpenAI:
		gateway, err = newOpenAIGateway(cfg)
	case ProviderDeepSeek:
		gateway, err = newDeepSeekGateway(cfg)
	case ProviderDoubao:
		gateway, err = newDoubaoGateway(cfg)
	case ProviderGemini:
		gateway, err = newGeminiGateway(cfg)
	case ProviderAnthropic:
		gateway, err = newAnthropicGateway(cfg)
	default:
		return nil, fmt.Errorf("%w: %s (supported: %s, %s, %s, %s, %s)",
			types.ErrUnsupportedProvider,
			cfg.Provider,
			ProviderOpenAI, ProviderDeepSeek, ProviderDoubao, ProviderGemini, ProviderAnthropic,
		)
	}

	if err != nil {
		return nil, err
	}

	// Wrap with retry if configured
	if cfg.RetryConfig != nil {
		gateway = NewRetryableGateway(gateway, *cfg.RetryConfig)
	}

	return gateway, nil
}

func newOpenAIGateway(cfg GatewayConfig) (LLMGateway, error) {
	model := cfg.OpenAIModel
	if model == "" {
		model = DefaultOpenAIModel
	}

	baseURL := cfg.OpenAIBaseURL
	if baseURL == "" {
		baseURL = OpenAIBaseURL
	}

	return NewOpenAICompatibleGateway(OpenAICompatibleConfig{
		APIKey:   cfg.OpenAIAPIKey,
		Model:    model,
		BaseURL:  baseURL,
		Provider: ProviderOpenAI,
	})
}

func newDeepSeekGateway(cfg GatewayConfig) (LLMGateway, error) {
	model := cfg.DeepSeekModel
	if model == "" {
		model = DefaultDeepSeekModel
	}

	return NewOpenAICompatibleGateway(OpenAICompatibleConfig{
		APIKey:   cfg.DeepSeekAPIKey,
		Model:    model,
		BaseURL:  DeepSeekBaseURL,
		Provider: ProviderDeepSeek,
	})
}

func newDoubaoGateway(cfg GatewayConfig) (LLMGateway, error) {
	if cfg.DoubaoModel == "" {
		return nil, fmt.Errorf("%w: DOUBAO_MODEL is required (format: ep-xxx-xxx)", types.ErrInvalidModel)
	}

	return NewOpenAICompatibleGateway(OpenAICompatibleConfig{
		APIKey:   cfg.DoubaoAPIKey,
		Model:    cfg.DoubaoModel,
		BaseURL:  DoubaoBaseURL,
		Provider: ProviderDoubao,
	})
}

func newGeminiGateway(cfg GatewayConfig) (LLMGateway, error) {
	if cfg.GeminiAPIKey == "" {
		return nil, types.ErrInvalidAPIKey
	}
	// TODO: Implement Gemini gateway
	return &placeholderGateway{provider: ProviderGemini}, nil
}

func newAnthropicGateway(cfg GatewayConfig) (LLMGateway, error) {
	if cfg.AnthropicAPIKey == "" {
		return nil, types.ErrInvalidAPIKey
	}
	// TODO: Implement Anthropic gateway
	return &placeholderGateway{provider: ProviderAnthropic}, nil
}
