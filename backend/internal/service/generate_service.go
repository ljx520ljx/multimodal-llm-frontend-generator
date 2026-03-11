package service

import (
	"context"
	"log"
	"regexp"
	"strings"

	"multimodal-llm-frontend-generator/internal/gateway"
	"multimodal-llm-frontend-generator/internal/gateway/types"
)

// GenerateService handles code generation using LLM
type GenerateService interface {
	// Generate generates code from images (streaming)
	Generate(ctx context.Context, sessionID string, imageIDs []string, framework string) (<-chan SSEEvent, error)

	// Chat modifies code based on natural language instruction (streaming)
	Chat(ctx context.Context, sessionID string, message string) (<-chan SSEEvent, error)
}

// generateService implements GenerateService
type generateService struct {
	sessionStore  SessionStore
	promptService PromptService
	gateway       gateway.LLMGateway
	agentClient   AgentClient
}

// NewGenerateService creates a new GenerateService
func NewGenerateService(
	sessionStore SessionStore,
	promptService PromptService,
	gw gateway.LLMGateway,
) GenerateService {
	return &generateService{
		sessionStore:  sessionStore,
		promptService: promptService,
		gateway:       gw,
	}
}

// NewGenerateServiceWithAgent creates a new GenerateService with AgentClient support
func NewGenerateServiceWithAgent(
	sessionStore SessionStore,
	promptService PromptService,
	gw gateway.LLMGateway,
	agentClient AgentClient,
) GenerateService {
	return &generateService{
		sessionStore:  sessionStore,
		promptService: promptService,
		gateway:       gw,
		agentClient:   agentClient,
	}
}

// ErrNoCodeGenerated is returned when trying to chat without generated code
type ErrNoCodeGenerated struct{}

func (e *ErrNoCodeGenerated) Error() string {
	return "no code generated yet, please generate code first"
}

// Generate generates code from images
func (s *generateService) Generate(ctx context.Context, sessionID string, imageIDs []string, framework string) (<-chan SSEEvent, error) {
	// Get session
	session, err := s.sessionStore.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Get images
	images, err := s.sessionStore.GetImages(ctx, sessionID, imageIDs)
	if err != nil {
		return nil, err
	}

	// Update session framework
	session.Framework = framework
	if err := s.sessionStore.Update(ctx, session); err != nil {
		return nil, err
	}

	// Build prompt
	systemPrompt := s.promptService.BuildSystemPrompt(framework)
	var messages []types.Message

	if len(images) == 2 {
		// Use diff analysis for exactly 2 images
		messages = s.promptService.BuildDiffPrompt(images, framework)
	} else {
		messages = s.promptService.BuildGeneratePrompt(images, framework)
	}

	// Create chat request with larger max_tokens for complex pages
	req := &types.ChatRequest{
		Messages: append(
			[]types.Message{types.NewTextMessage(types.RoleSystem, systemPrompt)},
			messages...,
		),
		Options: &types.ChatOptions{
			MaxTokens: 8192, // Claude 3.5 Sonnet 的最大限制
		},
	}

	// Call LLM
	llmChan, err := s.gateway.ChatStream(ctx, req)
	if err != nil {
		return nil, err
	}

	// Create output channel
	outChan := make(chan SSEEvent, 16)

	// Process LLM output
	go func() {
		defer close(outChan)
		s.processLLMOutput(ctx, sessionID, llmChan, outChan)
	}()

	return outChan, nil
}

// Chat modifies code based on natural language instruction
func (s *generateService) Chat(ctx context.Context, sessionID string, message string) (<-chan SSEEvent, error) {
	// Get session
	session, err := s.sessionStore.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Check if code exists
	if session.Code == "" {
		return nil, &ErrNoCodeGenerated{}
	}

	// Get history (limit to last 10 entries)
	history, err := s.sessionStore.GetHistory(ctx, sessionID, 10)
	if err != nil {
		return nil, err
	}

	log.Printf("[Chat] Session %s has %d images for reference", sessionID, len(session.Images))

	// Add user message to history
	if err := s.sessionStore.AddHistory(ctx, sessionID, HistoryEntry{
		Role:    "user",
		Content: message,
		Type:    "text",
	}); err != nil {
		log.Printf("[Chat] Failed to add user message to history: %v", err)
	}

	// If agentClient is available, use Python Agent for chat (with tool calling support)
	if s.agentClient != nil {
		return s.chatViaAgent(ctx, sessionID, session, message, history)
	}

	// Fallback: direct LLM call without tool calling
	return s.chatViaLLM(ctx, sessionID, session, message, history)
}

// chatViaAgent uses Python Agent for chat with tool calling support
func (s *generateService) chatViaAgent(ctx context.Context, sessionID string, session *Session, message string, history []HistoryEntry) (<-chan SSEEvent, error) {
	// Convert images to agent format
	agentImages := make([]map[string]interface{}, len(session.Images))
	for i, img := range session.Images {
		agentImages[i] = map[string]interface{}{
			"id":     img.ID,
			"base64": img.Base64,
			"order":  img.Order,
		}
	}

	// Convert history to agent format
	agentHistory := make([]ChatHistoryEntry, len(history))
	for i, h := range history {
		agentHistory[i] = ChatHistoryEntry{
			Role:    h.Role,
			Content: h.Content,
		}
	}

	// Create agent chat request
	req := &AgentChatRequest{
		SessionID:   sessionID,
		Message:     message,
		CurrentCode: session.Code,
		Images:      agentImages,
		History:     agentHistory,
	}

	// Call Python Agent
	agentChan, err := s.agentClient.Chat(ctx, req)
	if err != nil {
		return nil, err
	}

	// Create output channel
	outChan := make(chan SSEEvent, 16)

	// Forward agent events and save code on completion
	go func() {
		defer close(outChan)
		s.processAgentChatOutput(ctx, sessionID, agentChan, outChan)
	}()

	return outChan, nil
}

// processAgentChatOutput forwards agent events and saves code on completion
func (s *generateService) processAgentChatOutput(ctx context.Context, sessionID string, agentChan <-chan SSEEvent, outChan chan<- SSEEvent) {
	var codeContent string

	for event := range agentChan {
		select {
		case <-ctx.Done():
			outChan <- SSEEvent{Type: SSETypeError, Content: "Request cancelled"}
			return
		default:
		}

		// Forward the event
		outChan <- event

		// Capture code content for saving
		// Note: event.Content is already extracted HTML (parseSSEData handles JSON parsing)
		if event.Type == SSETypeCode && event.Content != "" {
			codeContent = event.Content
		}

		// On done, save the code
		if event.Type == SSETypeDone && codeContent != "" {
			if err := s.sessionStore.UpdateCode(ctx, sessionID, codeContent); err != nil {
				log.Printf("[Chat] Failed to save code: %v", err)
			}
			if err := s.sessionStore.AddHistory(ctx, sessionID, HistoryEntry{
				Role:    "assistant",
				Content: codeContent,
				Type:    "code",
			}); err != nil {
				log.Printf("[Chat] Failed to add assistant history: %v", err)
			}
		}
	}
}

// chatViaLLM uses direct LLM call without tool calling (fallback)
func (s *generateService) chatViaLLM(ctx context.Context, sessionID string, session *Session, message string, history []HistoryEntry) (<-chan SSEEvent, error) {
	// Build prompt with original images for reference
	systemPrompt := s.promptService.BuildSystemPrompt(session.Framework)
	messages := s.promptService.BuildChatPrompt(session.Code, message, history, session.Images)

	// Create chat request with larger max_tokens
	req := &types.ChatRequest{
		Messages: append(
			[]types.Message{types.NewTextMessage(types.RoleSystem, systemPrompt)},
			messages...,
		),
		Options: &types.ChatOptions{
			MaxTokens: 8192,
		},
	}

	// Call LLM
	llmChan, err := s.gateway.ChatStream(ctx, req)
	if err != nil {
		return nil, err
	}

	// Create output channel
	outChan := make(chan SSEEvent, 16)

	// Process LLM output
	go func() {
		defer close(outChan)
		s.processLLMOutput(ctx, sessionID, llmChan, outChan)
	}()

	return outChan, nil
}

// processLLMOutput converts LLM chunks to SSE events
func (s *generateService) processLLMOutput(ctx context.Context, sessionID string, llmChan <-chan types.StreamChunk, outChan chan<- SSEEvent) {
	var fullContent strings.Builder

	for chunk := range llmChan {
		select {
		case <-ctx.Done():
			outChan <- SSEEvent{Type: SSETypeError, Content: "Request cancelled"}
			return
		default:
		}

		switch chunk.Type {
		case types.ChunkTypeContent:
			fullContent.WriteString(chunk.Content)
			// Determine event type based on content pattern
			eventType := s.detectContentType(fullContent.String(), chunk.Content)
			outChan <- SSEEvent{Type: eventType, Content: chunk.Content}

		case types.ChunkTypeError:
			outChan <- SSEEvent{Type: SSETypeError, Content: chunk.Error.Error()}
			return

		case types.ChunkTypeDone:
			// Extract and save code
			code := s.extractCode(fullContent.String())
			if code != "" {
				if err := s.sessionStore.UpdateCode(ctx, sessionID, code); err != nil {
					log.Printf("[Generate] Failed to save code: %v", err)
				}
				// Add assistant response to history
				if err := s.sessionStore.AddHistory(ctx, sessionID, HistoryEntry{
					Role:    "assistant",
					Content: code,
					Type:    "code",
				}); err != nil {
					log.Printf("[Generate] Failed to add assistant history: %v", err)
				}
			}
			outChan <- SSEEvent{Type: SSETypeDone, Content: ""}
			return
		}
	}
}

// detectContentType determines if content is thinking or code
func (s *generateService) detectContentType(fullContent, newContent string) string {
	// Count code block markers (```) to detect if we're in a code block
	codeBlockCount := strings.Count(fullContent, "```")

	// If odd number of ```, we're inside a code block (code)
	if codeBlockCount%2 == 1 {
		return SSETypeCode
	}

	// If we have closed code blocks, the rest is thinking
	if codeBlockCount >= 2 {
		return SSETypeThinking
	}

	// Before any code block, it's thinking (analysis text)
	return SSETypeThinking
}

// extractCode extracts code from the LLM response
func (s *generateService) extractCode(content string) string {
	// Pattern for code blocks: ```html, ```jsx, ```tsx, or just ```
	codeBlockPattern := regexp.MustCompile("```(?:html|jsx|tsx|javascript|typescript)?\\s*\\n([\\s\\S]*?)\\n```")
	matches := codeBlockPattern.FindAllStringSubmatch(content, -1)

	var code string

	if len(matches) > 0 {
		// Return the last code block (most likely the final version)
		lastMatch := matches[len(matches)-1]
		if len(lastMatch) > 1 {
			code = strings.TrimSpace(lastMatch[1])
		}
	} else {
		// No code block found, try to find HTML by looking for DOCTYPE or html tag
		doctypeIdx := strings.Index(content, "<!DOCTYPE")
		htmlIdx := strings.Index(content, "<html")

		startIdx := -1
		if doctypeIdx >= 0 {
			startIdx = doctypeIdx
		} else if htmlIdx >= 0 {
			startIdx = htmlIdx
		}

		if startIdx >= 0 {
			code = strings.TrimSpace(content[startIdx:])
		}
	}

	// Final cleanup: remove any remaining code block markers
	code = regexp.MustCompile("^```[\\w]*\\s*\\n?").ReplaceAllString(code, "")
	code = regexp.MustCompile("\\n?```\\s*$").ReplaceAllString(code, "")

	return strings.TrimSpace(code)
}
