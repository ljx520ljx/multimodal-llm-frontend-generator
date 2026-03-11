package main

import (
	"log"

	"multimodal-llm-frontend-generator/internal/app"
	"multimodal-llm-frontend-generator/internal/config"
)

func main() {
	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
