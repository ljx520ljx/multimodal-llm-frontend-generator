package app

import (
	"log"

	"multimodal-llm-frontend-generator/internal/config"
	"multimodal-llm-frontend-generator/internal/gateway"
	"multimodal-llm-frontend-generator/internal/handler"
	"multimodal-llm-frontend-generator/internal/middleware"
	"multimodal-llm-frontend-generator/internal/service"

	"github.com/gin-gonic/gin"
)

// App holds all application dependencies
type App struct {
	Config               *config.Config
	Router               *gin.Engine
	SessionStore         service.SessionStore
	ImageService         service.ImageService
	PromptService        service.PromptService
	GenerateService      service.GenerateService
	AgentGenerateService service.AgentGenerateService
	Gateway              gateway.LLMGateway
	AgentClient          service.AgentClient
}

// New creates and initializes the application
func New(cfg *config.Config) (*App, error) {
	app := &App{
		Config: cfg,
	}

	// Initialize services
	if err := app.initServices(); err != nil {
		return nil, err
	}

	// Initialize router
	app.initRouter()

	return app, nil
}

// initServices initializes all services
func (a *App) initServices() error {
	// Session store
	a.SessionStore = service.NewMemoryStore(a.Config.SessionTTL)

	// Image service
	a.ImageService = service.NewImageService(service.ImageServiceConfig{
		MaxSize:      a.Config.ImageMaxSize,
		Quality:      a.Config.ImageQuality,
		MaxDimension: a.Config.ImageMaxDimension,
		AllowedTypes: a.Config.ImageAllowedTypes,
	})

	// Prompt service
	a.PromptService = service.NewPromptServiceWithConfig(service.PromptServiceConfig{
		EnableFewShot: a.Config.EnableFewShot,
	})

	// LLM Gateway
	gwConfig := gateway.GatewayConfig{
		Provider:        a.Config.LLMProvider,
		OpenAIAPIKey:    a.Config.OpenAIAPIKey,
		OpenAIModel:     a.Config.OpenAIModel,
		OpenAIBaseURL:   a.Config.OpenAIBaseURL,
		DeepSeekAPIKey:  a.Config.DeepSeekAPIKey,
		DeepSeekModel:   a.Config.DeepSeekModel,
		DoubaoAPIKey:    a.Config.DoubaoAPIKey,
		DoubaoModel:     a.Config.DoubaoModel,
		GeminiAPIKey:    a.Config.GeminiAPIKey,
		AnthropicAPIKey: a.Config.AnthropicAPIKey,
	}

	gw, err := gateway.NewGateway(gwConfig)
	if err != nil {
		return err
	}
	a.Gateway = gw

	// Agent client (for Python Agent service)
	a.AgentClient = service.NewAgentClient(
		a.Config.AgentServiceURL,
		a.Config.AgentTimeout,
	)

	// Generate service (with AgentClient for chat via Python Agent)
	a.GenerateService = service.NewGenerateServiceWithAgent(
		a.SessionStore,
		a.PromptService,
		a.Gateway,
		a.AgentClient,
	)

	// Agent generate service (uses Python Agent for multi-agent pipeline)
	a.AgentGenerateService = service.NewAgentGenerateService(
		a.SessionStore,
		a.AgentClient,
	)

	return nil
}

// initRouter initializes the HTTP router
func (a *App) initRouter() {
	if a.Config.ServerMode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Middleware
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS(a.Config.AllowedOrigins))

	// Set max multipart memory (for file uploads)
	r.MaxMultipartMemory = a.Config.ImageMaxTotal

	// Health check
	r.GET("/health", handler.Health)

	// API routes
	api := r.Group("/api")
	{
		// Upload handler
		uploadHandler := handler.NewUploadHandler(a.SessionStore, a.ImageService)
		api.POST("/upload", uploadHandler.Handle)

		// Generate handler
		generateHandler := handler.NewGenerateHandler(a.GenerateService)
		api.POST("/generate", generateHandler.Handle)

		// Chat handler
		chatHandler := handler.NewChatHandler(a.GenerateService)
		api.POST("/chat", chatHandler.Handle)

		// Echo handler (for testing Go ↔ Python communication)
		echoHandler := handler.NewEchoHandler(a.AgentClient)
		api.POST("/echo", echoHandler.Handle)
		api.GET("/agent-health", echoHandler.HealthCheck)

		// Agent generate handler (multi-agent pipeline via Python)
		agentGenerateHandler := handler.NewAgentGenerateHandler(a.AgentGenerateService)
		api.POST("/agent/generate", agentGenerateHandler.Handle)
	}

	a.Router = r
}

// Run starts the HTTP server
func (a *App) Run() error {
	addr := ":" + a.Config.ServerPort
	log.Printf("Server starting on port %s in %s mode", a.Config.ServerPort, a.Config.ServerMode)
	log.Printf("LLM Provider: %s", a.Config.LLMProvider)
	log.Printf("Few-Shot Examples: %v", a.Config.EnableFewShot)
	return a.Router.Run(addr)
}

// Close cleans up resources
func (a *App) Close() {
	if store, ok := a.SessionStore.(*service.MemoryStore); ok {
		store.Close()
	}
}
