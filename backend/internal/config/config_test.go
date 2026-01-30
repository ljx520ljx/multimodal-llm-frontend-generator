package config

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func resetViper() {
	viper.Reset()
}

func TestLoad_Defaults(t *testing.T) {
	resetViper()

	cfg := Load()

	if cfg.ServerPort != "8080" {
		t.Errorf("expected ServerPort '8080', got '%s'", cfg.ServerPort)
	}

	if cfg.ServerMode != "development" {
		t.Errorf("expected ServerMode 'development', got '%s'", cfg.ServerMode)
	}

	if cfg.LLMProvider != "openai" {
		t.Errorf("expected LLMProvider 'openai', got '%s'", cfg.LLMProvider)
	}

	if cfg.OpenAIModel != "gpt-4o" {
		t.Errorf("expected OpenAIModel 'gpt-4o', got '%s'", cfg.OpenAIModel)
	}

	if cfg.ImageMaxSize != 10*1024*1024 {
		t.Errorf("expected ImageMaxSize 10MB, got %d", cfg.ImageMaxSize)
	}

	if cfg.ImageQuality != 80 {
		t.Errorf("expected ImageQuality 80, got %d", cfg.ImageQuality)
	}

	if cfg.ImageMaxDimension != 2048 {
		t.Errorf("expected ImageMaxDimension 2048, got %d", cfg.ImageMaxDimension)
	}

	if cfg.SessionTTL != 30*time.Minute {
		t.Errorf("expected SessionTTL 30m, got %v", cfg.SessionTTL)
	}

	if cfg.EnableFewShot != false {
		t.Error("expected EnableFewShot to be false by default")
	}
}

func TestLoad_EnvironmentOverrides(t *testing.T) {
	resetViper()

	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("SERVER_MODE", "production")
	os.Setenv("LLM_PROVIDER", "deepseek")
	os.Setenv("OPENAI_API_KEY", "test-key-123")
	os.Setenv("IMAGE_QUALITY", "70")
	os.Setenv("ENABLE_FEW_SHOT", "true")
	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("SERVER_MODE")
		os.Unsetenv("LLM_PROVIDER")
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("IMAGE_QUALITY")
		os.Unsetenv("ENABLE_FEW_SHOT")
	}()

	cfg := Load()

	if cfg.ServerPort != "9090" {
		t.Errorf("expected ServerPort '9090', got '%s'", cfg.ServerPort)
	}

	if cfg.ServerMode != "production" {
		t.Errorf("expected ServerMode 'production', got '%s'", cfg.ServerMode)
	}

	if cfg.LLMProvider != "deepseek" {
		t.Errorf("expected LLMProvider 'deepseek', got '%s'", cfg.LLMProvider)
	}

	if cfg.OpenAIAPIKey != "test-key-123" {
		t.Errorf("expected OpenAIAPIKey 'test-key-123', got '%s'", cfg.OpenAIAPIKey)
	}

	if cfg.ImageQuality != 70 {
		t.Errorf("expected ImageQuality 70, got %d", cfg.ImageQuality)
	}

	if cfg.EnableFewShot != true {
		t.Error("expected EnableFewShot to be true")
	}
}

func TestLoad_AllowedOrigins(t *testing.T) {
	resetViper()

	os.Setenv("ALLOWED_ORIGINS", "http://localhost:3000, http://example.com, https://app.example.com")
	defer os.Unsetenv("ALLOWED_ORIGINS")

	cfg := Load()

	if len(cfg.AllowedOrigins) != 3 {
		t.Fatalf("expected 3 allowed origins, got %d", len(cfg.AllowedOrigins))
	}

	expected := []string{"http://localhost:3000", "http://example.com", "https://app.example.com"}
	for i, origin := range expected {
		if cfg.AllowedOrigins[i] != origin {
			t.Errorf("expected origin '%s', got '%s'", origin, cfg.AllowedOrigins[i])
		}
	}
}

func TestLoad_LLMTimeout(t *testing.T) {
	resetViper()

	os.Setenv("LLM_TIMEOUT", "10m")
	defer os.Unsetenv("LLM_TIMEOUT")

	cfg := Load()

	if cfg.LLMTimeout != 10*time.Minute {
		t.Errorf("expected LLMTimeout 10m, got %v", cfg.LLMTimeout)
	}
}

func TestLoad_InvalidLLMTimeout(t *testing.T) {
	resetViper()

	os.Setenv("LLM_TIMEOUT", "invalid")
	defer os.Unsetenv("LLM_TIMEOUT")

	cfg := Load()

	// Should fall back to default
	if cfg.LLMTimeout != 5*time.Minute {
		t.Errorf("expected default LLMTimeout 5m, got %v", cfg.LLMTimeout)
	}
}

func TestLoad_SessionTTL(t *testing.T) {
	resetViper()

	os.Setenv("SESSION_TTL", "1h")
	defer os.Unsetenv("SESSION_TTL")

	cfg := Load()

	if cfg.SessionTTL != time.Hour {
		t.Errorf("expected SessionTTL 1h, got %v", cfg.SessionTTL)
	}
}

func TestLoad_InvalidSessionTTL(t *testing.T) {
	resetViper()

	os.Setenv("SESSION_TTL", "invalid")
	defer os.Unsetenv("SESSION_TTL")

	cfg := Load()

	// Should fall back to default
	if cfg.SessionTTL != 30*time.Minute {
		t.Errorf("expected default SessionTTL 30m, got %v", cfg.SessionTTL)
	}
}

func TestLoad_ImageAllowedTypes(t *testing.T) {
	resetViper()

	os.Setenv("IMAGE_ALLOWED_TYPES", "image/png, image/jpeg, image/gif")
	defer os.Unsetenv("IMAGE_ALLOWED_TYPES")

	cfg := Load()

	if len(cfg.ImageAllowedTypes) != 3 {
		t.Fatalf("expected 3 allowed types, got %d", len(cfg.ImageAllowedTypes))
	}

	expected := []string{"image/png", "image/jpeg", "image/gif"}
	for i, mimeType := range expected {
		if cfg.ImageAllowedTypes[i] != mimeType {
			t.Errorf("expected type '%s', got '%s'", mimeType, cfg.ImageAllowedTypes[i])
		}
	}
}

func TestGetGatewayConfig(t *testing.T) {
	cfg := &Config{
		LLMProvider:     "openai",
		OpenAIAPIKey:    "openai-key",
		OpenAIModel:     "gpt-4o",
		DeepSeekAPIKey:  "deepseek-key",
		DeepSeekModel:   "deepseek-chat",
		DoubaoAPIKey:    "doubao-key",
		DoubaoModel:     "doubao-model",
		GeminiAPIKey:    "gemini-key",
		AnthropicAPIKey: "anthropic-key",
	}

	gatewayConfig := cfg.GetGatewayConfig()

	if gatewayConfig["provider"] != "openai" {
		t.Errorf("expected provider 'openai', got '%s'", gatewayConfig["provider"])
	}

	if gatewayConfig["openai_api_key"] != "openai-key" {
		t.Errorf("expected openai_api_key 'openai-key', got '%s'", gatewayConfig["openai_api_key"])
	}

	if gatewayConfig["deepseek_api_key"] != "deepseek-key" {
		t.Errorf("expected deepseek_api_key 'deepseek-key', got '%s'", gatewayConfig["deepseek_api_key"])
	}
}

func TestLoad_AllProviderKeys(t *testing.T) {
	resetViper()

	os.Setenv("OPENAI_API_KEY", "openai-key")
	os.Setenv("DEEPSEEK_API_KEY", "deepseek-key")
	os.Setenv("DOUBAO_API_KEY", "doubao-key")
	os.Setenv("DOUBAO_MODEL", "doubao-pro")
	os.Setenv("GEMINI_API_KEY", "gemini-key")
	os.Setenv("ANTHROPIC_API_KEY", "anthropic-key")
	defer func() {
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("DEEPSEEK_API_KEY")
		os.Unsetenv("DOUBAO_API_KEY")
		os.Unsetenv("DOUBAO_MODEL")
		os.Unsetenv("GEMINI_API_KEY")
		os.Unsetenv("ANTHROPIC_API_KEY")
	}()

	cfg := Load()

	if cfg.OpenAIAPIKey != "openai-key" {
		t.Errorf("expected OpenAIAPIKey 'openai-key', got '%s'", cfg.OpenAIAPIKey)
	}

	if cfg.DeepSeekAPIKey != "deepseek-key" {
		t.Errorf("expected DeepSeekAPIKey 'deepseek-key', got '%s'", cfg.DeepSeekAPIKey)
	}

	if cfg.DoubaoAPIKey != "doubao-key" {
		t.Errorf("expected DoubaoAPIKey 'doubao-key', got '%s'", cfg.DoubaoAPIKey)
	}

	if cfg.DoubaoModel != "doubao-pro" {
		t.Errorf("expected DoubaoModel 'doubao-pro', got '%s'", cfg.DoubaoModel)
	}

	if cfg.GeminiAPIKey != "gemini-key" {
		t.Errorf("expected GeminiAPIKey 'gemini-key', got '%s'", cfg.GeminiAPIKey)
	}

	if cfg.AnthropicAPIKey != "anthropic-key" {
		t.Errorf("expected AnthropicAPIKey 'anthropic-key', got '%s'", cfg.AnthropicAPIKey)
	}
}
