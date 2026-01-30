package service

import (
	"context"
	"testing"
	"time"

	"multimodal-llm-frontend-generator/internal/gateway/types"
)

// mockGateway implements gateway.LLMGateway for testing
type mockGateway struct {
	chunks   []types.StreamChunk
	provider string
}

func (m *mockGateway) ChatStream(ctx context.Context, req *types.ChatRequest) (<-chan types.StreamChunk, error) {
	ch := make(chan types.StreamChunk)
	go func() {
		defer close(ch)
		for _, chunk := range m.chunks {
			select {
			case <-ctx.Done():
				return
			case ch <- chunk:
			}
		}
	}()
	return ch, nil
}

func (m *mockGateway) Provider() string {
	return m.provider
}

func TestGenerateService_Generate(t *testing.T) {
	store := NewMemoryStore(30 * time.Minute)
	defer store.Close()

	promptService := NewPromptService()

	mockGw := &mockGateway{
		provider: "mock",
		chunks: []types.StreamChunk{
			{Type: types.ChunkTypeContent, Content: "<thinking>Analyzing..."},
			{Type: types.ChunkTypeContent, Content: "</thinking>\n\n```jsx\n"},
			{Type: types.ChunkTypeContent, Content: "export default function App() {\n"},
			{Type: types.ChunkTypeContent, Content: "  return <div>Hello</div>\n"},
			{Type: types.ChunkTypeContent, Content: "}\n```"},
			{Type: types.ChunkTypeDone},
		},
	}

	service := NewGenerateService(store, promptService, mockGw)

	ctx := context.Background()

	// Create session and add image
	session, _ := store.Create(ctx)
	store.AddImage(ctx, session.ID, &ImageData{
		ID:       "img-1",
		Filename: "test.png",
		Base64:   "data:image/png;base64,test",
	})

	// Generate code
	eventChan, err := service.Generate(ctx, session.ID, []string{"img-1"}, "react")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Collect events
	var events []SSEEvent
	for event := range eventChan {
		events = append(events, event)
	}

	// Should have multiple events
	if len(events) < 2 {
		t.Errorf("expected at least 2 events, got %d", len(events))
	}

	// Last event should be done
	lastEvent := events[len(events)-1]
	if lastEvent.Type != SSETypeDone {
		t.Errorf("expected last event to be done, got %s", lastEvent.Type)
	}

	// Code should be saved to session
	updatedSession, _ := store.Get(ctx, session.ID)
	if updatedSession.Code == "" {
		t.Error("expected code to be saved to session")
	}
}

func TestGenerateService_Generate_SessionNotFound(t *testing.T) {
	store := NewMemoryStore(30 * time.Minute)
	defer store.Close()

	promptService := NewPromptService()
	mockGw := &mockGateway{provider: "mock"}

	service := NewGenerateService(store, promptService, mockGw)

	ctx := context.Background()
	_, err := service.Generate(ctx, "non-existent", []string{"img-1"}, "react")
	if err == nil {
		t.Error("expected error for non-existent session")
	}
}

func TestGenerateService_Chat(t *testing.T) {
	store := NewMemoryStore(30 * time.Minute)
	defer store.Close()

	promptService := NewPromptService()

	newCode := `export default function App() {
  return <div className="text-blue-500">Hello</div>
}`

	mockGw := &mockGateway{
		provider: "mock",
		chunks: []types.StreamChunk{
			{Type: types.ChunkTypeContent, Content: "```jsx\n"},
			{Type: types.ChunkTypeContent, Content: newCode},
			{Type: types.ChunkTypeContent, Content: "\n```"},
			{Type: types.ChunkTypeDone},
		},
	}

	service := NewGenerateService(store, promptService, mockGw)

	ctx := context.Background()

	// Create session with existing code
	session, _ := store.Create(ctx)
	session.Framework = "react"
	originalCode := "export default function App() { return <div>Hello</div> }"
	session.Code = originalCode
	store.Update(ctx, session)

	// Chat to modify code
	eventChan, err := service.Chat(ctx, session.ID, "把文字改成蓝色")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Collect events
	var events []SSEEvent
	for event := range eventChan {
		events = append(events, event)
	}

	if len(events) == 0 {
		t.Error("expected at least 1 event")
	}

	// Code should be updated
	updatedSession, _ := store.Get(ctx, session.ID)
	if updatedSession.Code == originalCode {
		t.Error("expected code to be updated")
	}

	// History should be updated (1 user message + 1 assistant response)
	history, _ := store.GetHistory(ctx, session.ID, 10)
	if len(history) < 2 {
		t.Errorf("expected at least 2 history entries, got %d", len(history))
	}
}

func TestGenerateService_Chat_NoCode(t *testing.T) {
	store := NewMemoryStore(30 * time.Minute)
	defer store.Close()

	promptService := NewPromptService()
	mockGw := &mockGateway{provider: "mock"}

	service := NewGenerateService(store, promptService, mockGw)

	ctx := context.Background()

	// Create session without code
	session, _ := store.Create(ctx)

	_, err := service.Chat(ctx, session.ID, "修改代码")
	if err == nil {
		t.Error("expected error when no code exists")
	}

	if _, ok := err.(*ErrNoCodeGenerated); !ok {
		t.Errorf("expected ErrNoCodeGenerated, got %T", err)
	}
}

func TestExtractCode(t *testing.T) {
	service := &generateService{}

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "HTML code block",
			content: `Here's the code:

` + "```html" + `
<!DOCTYPE html>
<html><body>Hello</body></html>
` + "```",
			expected: "<!DOCTYPE html>\n<html><body>Hello</body></html>",
		},
		{
			name:     "JSX code block",
			content:  "```jsx\nconst App = () => <div />\n```",
			expected: "const App = () => <div />",
		},
		{
			name:     "Multiple code blocks",
			content:  "```html\nfirst\n```\n\n```html\nsecond\n```",
			expected: "second",
		},
		{
			name:     "No code block",
			content:  "Just some text",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.extractCode(tt.content)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestDetectContentType(t *testing.T) {
	service := &generateService{}

	tests := []struct {
		name        string
		fullContent string
		newContent  string
		expected    string
	}{
		{
			name:        "Plain text",
			fullContent: "Analyzing the image",
			newContent:  "...",
			expected:    SSETypeThinking,
		},
		{
			name:        "Inside thinking tag",
			fullContent: "<thinking>Analyzing",
			newContent:  " more",
			expected:    SSETypeThinking,
		},
		{
			name:        "Inside code block",
			fullContent: "```jsx\nconst x = 1",
			newContent:  "\nconst y = 2",
			expected:    SSETypeCode,
		},
		{
			name:        "After code block",
			fullContent: "```jsx\ncode\n```\nExplanation",
			newContent:  " here",
			expected:    SSETypeThinking,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.detectContentType(tt.fullContent, tt.newContent)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
