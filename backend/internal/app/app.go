package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"multimodal-llm-frontend-generator/internal/config"
	"multimodal-llm-frontend-generator/internal/gateway"
	"multimodal-llm-frontend-generator/internal/handler"
	"multimodal-llm-frontend-generator/internal/migrate"
	"multimodal-llm-frontend-generator/internal/middleware"
	"multimodal-llm-frontend-generator/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
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
	PreviewService       *service.PreviewService
	AuthService          *service.AuthService
	ProjectService       *service.ProjectService
	CodeVersionService   *service.CodeVersionService
	FeedbackService      *service.FeedbackService
	Gateway              gateway.LLMGateway
	AgentClient          service.AgentClient
	RateLimiter          *middleware.RateLimiter
	dbPool               *pgxpool.Pool
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
	if a.Config.DatabaseURL != "" {
		pool, err := pgxpool.New(context.Background(), a.Config.DatabaseURL)
		if err != nil {
			return err
		}
		if err := pool.Ping(context.Background()); err != nil {
			pool.Close()
			return err
		}
		a.dbPool = pool

		// Auto-run database migrations on startup
		if err := migrate.Run(context.Background(), pool); err != nil {
			pool.Close()
			return fmt.Errorf("auto-migrate: %w", err)
		}

		a.SessionStore = service.NewPostgresStore(pool, a.Config.SessionTTL, a.Config.SessionHistoryLimit)
		a.PreviewService = service.NewPreviewService(pool, a.SessionStore)
		a.CodeVersionService = service.NewCodeVersionService(pool)
		a.FeedbackService = service.NewFeedbackService(pool)

		// Auth service (only with PostgreSQL)
		if a.Config.JWTSecret != "" {
			a.AuthService = service.NewAuthService(
				pool,
				a.Config.JWTSecret,
				a.Config.JWTExpiry,
				a.Config.GitHubClientID,
				a.Config.GitHubClientSecret,
				a.Config.BaseURL,
			)
			a.ProjectService = service.NewProjectService(pool)
			log.Println("Auth and project services enabled")
		}

		log.Println("Using PostgreSQL session store")
	} else {
		a.SessionStore = service.NewMemoryStore(a.Config.SessionTTL, a.Config.SessionHistoryLimit)
		log.Println("Using in-memory session store (no DATABASE_URL configured)")
	}

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

	// Rate limiter
	a.RateLimiter = middleware.NewRateLimiter(middleware.RateLimitConfig{
		IPRate:         a.Config.RateLimitIPRate,
		IPBurst:        a.Config.RateLimitIPBurst,
		IPCleanupTTL:   a.Config.RateLimitIPCleanupTTL,
		MaxConcurrent:  a.Config.RateLimitMaxConcurrent,
	})

	// Health check (no rate limiting)
	r.GET("/health", handler.Health)

	// API routes (with rate limiting)
	api := r.Group("/api")
	api.Use(a.RateLimiter.Middleware())

	// Apply optional auth middleware BEFORE routes so it takes effect
	if a.AuthService != nil {
		api.Use(middleware.OptionalJWTAuth(a.AuthService))
	}

	{
		// Upload handler
		uploadHandler := handler.NewUploadHandler(a.SessionStore, a.ImageService)
		api.POST("/upload", uploadHandler.Handle)

		// Generate handler
		generateHandler := handler.NewGenerateHandler(a.GenerateService, a.Config.HandlerTimeout)
		api.POST("/generate", generateHandler.Handle)

		// Chat handler
		chatHandler := handler.NewChatHandler(a.GenerateService, a.Config.HandlerTimeout)
		api.POST("/chat", chatHandler.Handle)

		// Echo handler (for testing Go ↔ Python communication)
		echoHandler := handler.NewEchoHandler(a.AgentClient)
		api.POST("/echo", echoHandler.Handle)
		api.GET("/agent-health", echoHandler.HealthCheck)

		// Agent generate handler (multi-agent pipeline via Python)
		agentGenerateHandler := handler.NewAgentGenerateHandler(a.AgentGenerateService, a.Config.HandlerTimeout)
		api.POST("/agent/generate", agentGenerateHandler.Handle)
	}

	// Auth routes (with rate limiting, under /api group)
	if a.AuthService != nil {
		authHandler := handler.NewAuthHandler(a.AuthService)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.GET("/github", authHandler.GitHubAuth)
			auth.GET("/github/callback", authHandler.GitHubCallback)
			auth.GET("/me", middleware.JWTAuth(a.AuthService), authHandler.GetCurrentUser)
		}

		// Project routes (require authentication)
		if a.ProjectService != nil {
			projectHandler := handler.NewProjectHandler(a.ProjectService)
			projects := api.Group("/projects")
			projects.Use(middleware.JWTAuth(a.AuthService))
			{
				projects.GET("", projectHandler.List)
				projects.POST("", projectHandler.Create)
				projects.GET("/:id", projectHandler.Get)
				projects.PUT("/:id", projectHandler.Update)
				projects.DELETE("/:id", projectHandler.Delete)
			}
		}
	}

	// Code version and feedback routes (only available with PostgreSQL)
	if a.CodeVersionService != nil {
		cvHandler := handler.NewCodeVersionHandler(a.CodeVersionService)
		fbHandler := handler.NewFeedbackHandler(a.FeedbackService)

		api.GET("/sessions/:session_id/versions", cvHandler.List)
		api.GET("/versions/:id", cvHandler.Get)
		api.POST("/feedback", fbHandler.Submit)
		api.GET("/sessions/:session_id/feedback", fbHandler.ListBySession)
		api.GET("/sessions/:session_id/feedback/stats", fbHandler.Stats)
	}

	// Preview sharing routes (only available with PostgreSQL)
	if a.PreviewService != nil {
		previewHandler := handler.NewPreviewHandler(a.PreviewService)
		// Public preview access (no rate limiting)
		r.GET("/p/:code", previewHandler.ServePreview)
		// Share management API (with rate limiting)
		api.POST("/share", previewHandler.CreateShare)
		api.PUT("/share/:code", previewHandler.UpdateShare)
		api.DELETE("/share/:code", previewHandler.DeleteShare)
	}

	// Internal API routes (no rate limiting - for inter-service communication only)
	internal := r.Group("/api/internal")
	if a.Config.InternalAPIToken != "" {
		internal.Use(internalAPIAuth(a.Config.InternalAPIToken))
	}
	{
		checkpointHandler := handler.NewCheckpointHandler(a.SessionStore)
		internal.PUT("/checkpoint/:session_id", checkpointHandler.SaveCheckpoint)
		internal.GET("/checkpoint/:session_id", checkpointHandler.GetCheckpoint)
	}

	a.Router = r
}

// Run starts the HTTP server with graceful shutdown support
func (a *App) Run() error {
	addr := ":" + a.Config.ServerPort
	log.Printf("Server starting on port %s in %s mode", a.Config.ServerPort, a.Config.ServerMode)
	log.Printf("LLM Provider: %s", a.Config.LLMProvider)
	log.Printf("Few-Shot Examples: %v", a.Config.EnableFewShot)

	srv := &http.Server{
		Addr:    addr,
		Handler: a.Router,
	}

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for interrupt signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		log.Println("Shutting down server...")
	case err := <-errChan:
		return err
	}

	// Graceful shutdown with 30s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	a.Close()
	log.Println("Server exited gracefully")
	return nil
}

// internalAPIAuth returns a middleware that validates the X-Internal-Token header
func internalAPIAuth(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		provided := c.GetHeader("X-Internal-Token")
		if provided == "" || provided != token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Invalid or missing internal API token",
			})
			return
		}
		c.Next()
	}
}

// Close cleans up resources
func (a *App) Close() {
	if a.SessionStore != nil {
		a.SessionStore.Close()
	}
	if a.RateLimiter != nil {
		a.RateLimiter.Close()
	}
	// Close the database pool last (shared by AuthService, ProjectService, PreviewService)
	if a.dbPool != nil {
		a.dbPool.Close()
	}
}
