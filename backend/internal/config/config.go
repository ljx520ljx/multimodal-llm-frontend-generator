package config

import (
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	ServerPort     string
	ServerMode     string
	AllowedOrigins []string

	// LLM Provider
	LLMProvider string
	LLMTimeout  time.Duration

	// Prompt Configuration
	EnableFewShot bool // Enable few-shot examples in prompts

	// OpenAI
	OpenAIAPIKey  string
	OpenAIModel   string
	OpenAIBaseURL string // Custom base URL for OpenAI-compatible APIs

	// DeepSeek
	DeepSeekAPIKey string
	DeepSeekModel  string

	// Doubao (Volcengine)
	DoubaoAPIKey string
	DoubaoModel  string

	// Gemini
	GeminiAPIKey string

	// Anthropic
	AnthropicAPIKey string

	// Image processing
	ImageMaxSize      int64    // Max size per image in bytes (default 10MB)
	ImageMaxTotal     int64    // Max total upload size in bytes (default 50MB)
	ImageQuality      int      // Compression quality 1-100 (default 80)
	ImageMaxDimension int      // Max width/height in pixels (default 2048)
	ImageAllowedTypes []string // Allowed MIME types

	// Session
	SessionTTL time.Duration // Session expiration time (default 30 minutes)
}

func Load() *Config {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Server defaults
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("SERVER_MODE", "development")
	viper.SetDefault("ALLOWED_ORIGINS", "http://localhost:3000")

	// LLM defaults
	viper.SetDefault("LLM_PROVIDER", "openai")
	viper.SetDefault("LLM_TIMEOUT", "300s")
	viper.SetDefault("OPENAI_MODEL", "gpt-4o")
	viper.SetDefault("DEEPSEEK_MODEL", "deepseek-chat")

	// Prompt defaults
	viper.SetDefault("ENABLE_FEW_SHOT", false)

	// Image defaults
	viper.SetDefault("IMAGE_MAX_SIZE", 10*1024*1024)       // 10MB
	viper.SetDefault("IMAGE_MAX_TOTAL", 50*1024*1024)      // 50MB
	viper.SetDefault("IMAGE_QUALITY", 80)                  // 80%
	viper.SetDefault("IMAGE_MAX_DIMENSION", 2048)          // 2048px
	viper.SetDefault("IMAGE_ALLOWED_TYPES", "image/png,image/jpeg,image/webp")

	// Session defaults
	viper.SetDefault("SESSION_TTL", "30m")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("No .env file found, using environment variables and defaults")
	}

	origins := viper.GetString("ALLOWED_ORIGINS")
	allowedOrigins := strings.Split(origins, ",")
	for i := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
	}

	timeout, err := time.ParseDuration(viper.GetString("LLM_TIMEOUT"))
	if err != nil {
		timeout = 5 * time.Minute
	}

	sessionTTL, err := time.ParseDuration(viper.GetString("SESSION_TTL"))
	if err != nil {
		sessionTTL = 30 * time.Minute
	}

	// Parse allowed image types
	allowedTypesStr := viper.GetString("IMAGE_ALLOWED_TYPES")
	allowedTypes := strings.Split(allowedTypesStr, ",")
	for i := range allowedTypes {
		allowedTypes[i] = strings.TrimSpace(allowedTypes[i])
	}

	return &Config{
		ServerPort:     viper.GetString("SERVER_PORT"),
		ServerMode:     viper.GetString("SERVER_MODE"),
		AllowedOrigins: allowedOrigins,

		LLMProvider: viper.GetString("LLM_PROVIDER"),
		LLMTimeout:  timeout,

		EnableFewShot: viper.GetBool("ENABLE_FEW_SHOT"),

		OpenAIAPIKey:  viper.GetString("OPENAI_API_KEY"),
		OpenAIModel:   viper.GetString("OPENAI_MODEL"),
		OpenAIBaseURL: viper.GetString("OPENAI_BASE_URL"),

		DeepSeekAPIKey: viper.GetString("DEEPSEEK_API_KEY"),
		DeepSeekModel:  viper.GetString("DEEPSEEK_MODEL"),

		DoubaoAPIKey: viper.GetString("DOUBAO_API_KEY"),
		DoubaoModel:  viper.GetString("DOUBAO_MODEL"),

		GeminiAPIKey:    viper.GetString("GEMINI_API_KEY"),
		AnthropicAPIKey: viper.GetString("ANTHROPIC_API_KEY"),

		ImageMaxSize:      viper.GetInt64("IMAGE_MAX_SIZE"),
		ImageMaxTotal:     viper.GetInt64("IMAGE_MAX_TOTAL"),
		ImageQuality:      viper.GetInt("IMAGE_QUALITY"),
		ImageMaxDimension: viper.GetInt("IMAGE_MAX_DIMENSION"),
		ImageAllowedTypes: allowedTypes,

		SessionTTL: sessionTTL,
	}
}

// GetGatewayConfig returns the gateway configuration based on the current provider
func (c *Config) GetGatewayConfig() map[string]string {
	return map[string]string{
		"provider":          c.LLMProvider,
		"openai_api_key":    c.OpenAIAPIKey,
		"openai_model":      c.OpenAIModel,
		"openai_base_url":   c.OpenAIBaseURL,
		"deepseek_api_key":  c.DeepSeekAPIKey,
		"deepseek_model":    c.DeepSeekModel,
		"doubao_api_key":    c.DoubaoAPIKey,
		"doubao_model":      c.DoubaoModel,
		"gemini_api_key":    c.GeminiAPIKey,
		"anthropic_api_key": c.AnthropicAPIKey,
	}
}
