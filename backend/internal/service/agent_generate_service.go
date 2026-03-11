package service

import (
	"context"
	"log"
)

// AgentGenerateService handles code generation using Python Agent
type AgentGenerateService interface {
	// Generate generates code from images or description via Python Agent (streaming)
	Generate(ctx context.Context, sessionID string, imageIDs []string, description string) (<-chan SSEEvent, error)
}

// agentGenerateService implements AgentGenerateService
type agentGenerateService struct {
	sessionStore SessionStore
	agentClient  AgentClient
}

// NewAgentGenerateService creates a new AgentGenerateService
func NewAgentGenerateService(
	sessionStore SessionStore,
	agentClient AgentClient,
) AgentGenerateService {
	return &agentGenerateService{
		sessionStore: sessionStore,
		agentClient:  agentClient,
	}
}

// Generate generates code from images or description via Python Agent
func (s *agentGenerateService) Generate(ctx context.Context, sessionID string, imageIDs []string, description string) (<-chan SSEEvent, error) {
	// Get session
	session, err := s.sessionStore.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Get images (may be empty for text-to-UI)
	var agentImages []map[string]any
	if len(imageIDs) > 0 {
		images, err := s.sessionStore.GetImages(ctx, sessionID, imageIDs)
		if err != nil {
			return nil, err
		}

		log.Printf("[AgentGenerate] Session %s: sending %d images to Python Agent", sessionID, len(images))

		agentImages = make([]map[string]any, len(images))
		for i, img := range images {
			agentImages[i] = map[string]any{
				"id":     img.ID,
				"base64": img.Base64,
				"order":  img.Order,
			}
		}
	} else {
		log.Printf("[AgentGenerate] Session %s: text-to-UI generation with description", sessionID)
		agentImages = []map[string]any{}
	}

	// Create agent request
	req := &AgentGenerateRequest{
		SessionID:   sessionID,
		Images:      agentImages,
		Description: description,
		Options: map[string]any{
			"max_retries": 3,
		},
	}

	// Call Python Agent
	eventChan, err := s.agentClient.Generate(ctx, req)
	if err != nil {
		return nil, err
	}

	// Create output channel
	outChan := make(chan SSEEvent, 16)

	// Process agent output
	go func() {
		defer close(outChan)
		s.processAgentOutput(ctx, sessionID, session, eventChan, outChan)
	}()

	return outChan, nil
}

// processAgentOutput handles SSE events from Python Agent
func (s *agentGenerateService) processAgentOutput(
	ctx context.Context,
	sessionID string,
	session *Session,
	eventChan <-chan SSEEvent,
	outChan chan<- SSEEvent,
) {
	for event := range eventChan {
		select {
		case <-ctx.Done():
			outChan <- SSEEvent{Type: SSETypeError, Content: "Request cancelled"}
			return
		default:
		}

		// Forward event to client
		outChan <- event

		// Handle code event - save to session
		if event.Type == SSETypeCode {
			if err := s.sessionStore.UpdateCode(ctx, sessionID, event.Content); err != nil {
				log.Printf("[AgentGenerate] Failed to save code: %v", err)
			}
		}

		// Handle done event
		if event.Type == SSETypeDone {
			return
		}

		// Handle error event
		if event.Type == SSETypeError {
			log.Printf("[AgentGenerate] Error from Python Agent: %s", event.Content)
			return
		}
	}
}
