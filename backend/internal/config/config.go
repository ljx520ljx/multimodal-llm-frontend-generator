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

	// Database
	DatabaseURL string // PostgreSQL connection URL (empty = use MemoryStore)

	// Auth
	JWTSecret          string        // JWT signing secret
	JWTExpiry          time.Duration // JWT token expiry (default 24h)
	GitHubClientID     string        // GitHub OAuth2 client ID
	GitHubClientSecret string        // GitHub OAuth2 client secret
	BaseURL            string        // Public base URL for OAuth callbacks

	// Session
	SessionTTL          time.Duration // Session expiration time (default 30 minutes)
	SessionHistoryLimit int           // Max number of history entries per session (default 20)

	// Agent Service (Python)
	AgentServiceURL string        // Agent service URL (default http://localhost:8081)
	AgentTimeout    time.Duration // Agent service timeout (default 180s)
	HandlerTimeout  time.Duration // Handler-level context timeout for SSE endpoints (default 240s)

	// Rate Limiting
	RateLimitIPRate       float64       // Requests per second per IP (default 10)
	RateLimitIPBurst      int           // Max burst size per IP (default 20)
	RateLimitIPCleanupTTL time.Duration // TTL for idle IP entries (default 10m)
	RateLimitMaxConcurrent int          // Max concurrent requests for heavy endpoints (default 20)

	// Internal API
	InternalAPIToken string // Token for inter-service communication
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
	viper.SetDefault("LLM_TIMEOUT", "120s")
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

	// Database defaults
	viper.SetDefault("DATABASE_URL", "")

	// Auth defaults
	viper.SetDefault("JWT_SECRET", "")
	viper.SetDefault("JWT_EXPIRY", "24h")
	viper.SetDefault("GITHUB_CLIENT_ID", "")
	viper.SetDefault("GITHUB_CLIENT_SECRET", "")
	viper.SetDefault("BASE_URL", "http://localhost:8080")

	// Session defaults
	viper.SetDefault("SESSION_TTL", "30m")
	viper.SetDefault("SESSION_HISTORY_LIMIT", 20)

	// Agent service defaults
	// Timeout chain (inner → outer): Python LLM 120s < AGENT_TIMEOUT 360s < HANDLER_TIMEOUT 480s < Frontend SSE 600s
	// Quality 模式需要 4+ 个 Agent 串行调用（各 40-90s），总时间可达 200-400s
	viper.SetDefault("AGENT_SERVICE_URL", "http://localhost:8081")
	viper.SetDefault("AGENT_TIMEOUT", "360s")
	viper.SetDefault("HANDLER_TIMEOUT", "480s")

	// Rate limiting defaults
	viper.SetDefault("RATE_LIMIT_IP_RATE", 10.0)
	viper.SetDefault("RATE_LIMIT_IP_BURST", 20)
	viper.SetDefault("RATE_LIMIT_IP_CLEANUP_TTL", "10m")
	viper.SetDefault("RATE_LIMIT_MAX_CONCURRENT", 20)

	// Internal API defaults
	viper.SetDefault("INTERNAL_API_TOKEN", "")

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
		timeout = 120 * time.Second
	}

	sessionTTL, err := time.ParseDuration(viper.GetString("SESSION_TTL"))
	if err != nil {
		sessionTTL = 30 * time.Minute
	}

	agentTimeout, err := time.ParseDuration(viper.GetString("AGENT_TIMEOUT"))
	if err != nil {
		agentTimeout = 180 * time.Second
	}

	handlerTimeout, err := time.ParseDuration(viper.GetString("HANDLER_TIMEOUT"))
	if err != nil {
		handlerTimeout = 240 * time.Second
	}

	jwtExpiry, err := time.ParseDuration(viper.GetString("JWT_EXPIRY"))
	if err != nil {
		jwtExpiry = 24 * time.Hour
	}

	rateLimitCleanupTTL, err := time.ParseDuration(viper.GetString("RATE_LIMIT_IP_CLEANUP_TTL"))
	if err != nil {
		rateLimitCleanupTTL = 10 * time.Minute
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

		DatabaseURL: viper.GetString("DATABASE_URL"),

		JWTSecret:          viper.GetString("JWT_SECRET"),
		JWTExpiry:          jwtExpiry,
		GitHubClientID:     viper.GetString("GITHUB_CLIENT_ID"),
		GitHubClientSecret: viper.GetString("GITHUB_CLIENT_SECRET"),
		BaseURL:            viper.GetString("BASE_URL"),

		SessionTTL:          sessionTTL,
		SessionHistoryLimit: viper.GetInt("SESSION_HISTORY_LIMIT"),

		AgentServiceURL: viper.GetString("AGENT_SERVICE_URL"),
		AgentTimeout:    agentTimeout,
		HandlerTimeout:  handlerTimeout,

		RateLimitIPRate:        viper.GetFloat64("RATE_LIMIT_IP_RATE"),
		RateLimitIPBurst:       viper.GetInt("RATE_LIMIT_IP_BURST"),
		RateLimitIPCleanupTTL:  rateLimitCleanupTTL,
		RateLimitMaxConcurrent: viper.GetInt("RATE_LIMIT_MAX_CONCURRENT"),

		InternalAPIToken: viper.GetString("INTERNAL_API_TOKEN"),
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
