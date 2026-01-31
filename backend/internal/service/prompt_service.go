package service

import (
	"log"

	"multimodal-llm-frontend-generator/internal/gateway/types"
)

// PromptService handles prompt construction for LLM calls
type PromptService interface {
	// BuildSystemPrompt builds the system prompt for the given framework
	BuildSystemPrompt(framework string) string

	// BuildGeneratePrompt builds the prompt for code generation
	BuildGeneratePrompt(images []ImageData, framework string) []types.Message

	// BuildChatPrompt builds the prompt for chat-based code modification
	// images: 原始设计稿图片，用于 AI 参考进行精修
	BuildChatPrompt(code string, message string, history []HistoryEntry, images []ImageData) []types.Message

	// BuildDiffPrompt builds the prompt for diff analysis (multi-image)
	BuildDiffPrompt(images []ImageData, framework string) []types.Message
}

// PromptServiceConfig holds configuration for PromptService
type PromptServiceConfig struct {
	EnableFewShot bool // Whether to include few-shot examples
}

// promptService implements PromptService
type promptService struct {
	builder *PromptBuilder
}

// NewPromptService creates a new PromptService with default configuration
func NewPromptService() PromptService {
	return NewPromptServiceWithConfig(PromptServiceConfig{
		EnableFewShot: false,
	})
}

// NewPromptServiceWithConfig creates a new PromptService with the specified configuration
func NewPromptServiceWithConfig(cfg PromptServiceConfig) PromptService {
	builder := NewPromptBuilder(cfg.EnableFewShot)
	log.Printf("PromptService initialized (few-shot: %v)", cfg.EnableFewShot)
	return &promptService{
		builder: builder,
	}
}

// BuildSystemPrompt builds the system prompt for the given framework
func (s *promptService) BuildSystemPrompt(framework string) string {
	return s.builder.BuildSystemPrompt(framework)
}

// BuildGeneratePrompt builds the prompt for code generation
func (s *promptService) BuildGeneratePrompt(images []ImageData, framework string) []types.Message {
	return s.builder.BuildGeneratePrompt(images, framework)
}

// BuildChatPrompt builds the prompt for chat-based code modification
// images: 原始设计稿图片，用于 AI 参考进行精修
func (s *promptService) BuildChatPrompt(code string, message string, history []HistoryEntry, images []ImageData) []types.Message {
	return s.builder.BuildChatPrompt(code, message, history, images)
}

// BuildDiffPrompt builds the prompt for diff analysis (multi-image)
func (s *promptService) BuildDiffPrompt(images []ImageData, framework string) []types.Message {
	return s.builder.BuildDiffPrompt(images, framework)
}
