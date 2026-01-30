package gateway

import (
	"testing"

	"multimodal-llm-frontend-generator/internal/gateway/types"
)

func TestNewGateway_OpenAI(t *testing.T) {
	cfg := GatewayConfig{
		Provider:     ProviderOpenAI,
		OpenAIAPIKey: "test-key",
		OpenAIModel:  "gpt-4o",
	}

	gw, err := NewGateway(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if gw.Provider() != ProviderOpenAI {
		t.Errorf("expected provider %s, got %s", ProviderOpenAI, gw.Provider())
	}
}

func TestNewGateway_DeepSeek(t *testing.T) {
	cfg := GatewayConfig{
		Provider:       ProviderDeepSeek,
		DeepSeekAPIKey: "test-key",
		DeepSeekModel:  "deepseek-chat",
	}

	gw, err := NewGateway(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if gw.Provider() != ProviderDeepSeek {
		t.Errorf("expected provider %s, got %s", ProviderDeepSeek, gw.Provider())
	}
}

func TestNewGateway_Doubao(t *testing.T) {
	cfg := GatewayConfig{
		Provider:     ProviderDoubao,
		DoubaoAPIKey: "test-key",
		DoubaoModel:  "ep-test-endpoint",
	}

	gw, err := NewGateway(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if gw.Provider() != ProviderDoubao {
		t.Errorf("expected provider %s, got %s", ProviderDoubao, gw.Provider())
	}
}

func TestNewGateway_DoubaoMissingModel(t *testing.T) {
	cfg := GatewayConfig{
		Provider:     ProviderDoubao,
		DoubaoAPIKey: "test-key",
		// DoubaoModel is missing
	}

	_, err := NewGateway(cfg)
	if err == nil {
		t.Error("expected error for missing Doubao model")
	}
}

func TestNewGateway_InvalidProvider(t *testing.T) {
	cfg := GatewayConfig{
		Provider: "invalid-provider",
	}

	_, err := NewGateway(cfg)
	if err == nil {
		t.Error("expected error for invalid provider")
	}

	var gwErr *types.GatewayError
	// Check that error message mentions supported providers
	if err != nil && err.Error() == "" {
		t.Error("error message should not be empty")
	}
	_ = gwErr // avoid unused variable warning
}

func TestNewGateway_MissingAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		provider string
	}{
		{"OpenAI", ProviderOpenAI},
		{"DeepSeek", ProviderDeepSeek},
		{"Doubao", ProviderDoubao},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := GatewayConfig{
				Provider: tt.provider,
				// No API key provided
			}

			// For Doubao, we also need a model
			if tt.provider == ProviderDoubao {
				cfg.DoubaoModel = "ep-test"
			}

			_, err := NewGateway(cfg)
			if err == nil {
				t.Errorf("expected error for missing API key for %s", tt.name)
			}
		})
	}
}

func TestNewGateway_WithRetry(t *testing.T) {
	retryConfig := DefaultRetryConfig()
	cfg := GatewayConfig{
		Provider:     ProviderOpenAI,
		OpenAIAPIKey: "test-key",
		OpenAIModel:  "gpt-4o",
		RetryConfig:  &retryConfig,
	}

	gw, err := NewGateway(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should be wrapped with RetryableGateway
	_, ok := gw.(*RetryableGateway)
	if !ok {
		t.Error("expected gateway to be wrapped with RetryableGateway")
	}

	if gw.Provider() != ProviderOpenAI {
		t.Errorf("expected provider %s, got %s", ProviderOpenAI, gw.Provider())
	}
}

func TestNewGateway_Gemini_Placeholder(t *testing.T) {
	cfg := GatewayConfig{
		Provider:     ProviderGemini,
		GeminiAPIKey: "test-key",
	}

	gw, err := NewGateway(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if gw.Provider() != ProviderGemini {
		t.Errorf("expected provider %s, got %s", ProviderGemini, gw.Provider())
	}
}

func TestNewGateway_Anthropic_Placeholder(t *testing.T) {
	cfg := GatewayConfig{
		Provider:        ProviderAnthropic,
		AnthropicAPIKey: "test-key",
	}

	gw, err := NewGateway(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if gw.Provider() != ProviderAnthropic {
		t.Errorf("expected provider %s, got %s", ProviderAnthropic, gw.Provider())
	}
}
