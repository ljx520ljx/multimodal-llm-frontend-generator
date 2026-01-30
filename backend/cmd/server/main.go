package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"multimodal-llm-frontend-generator/internal/app"
	"multimodal-llm-frontend-generator/internal/config"
)

func main() {
	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer application.Close()

	// Handle graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down server...")
		application.Close()
		os.Exit(0)
	}()

	if err := application.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
