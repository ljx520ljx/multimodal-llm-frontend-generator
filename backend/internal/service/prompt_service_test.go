package service

import (
	"strings"
	"testing"

	"multimodal-llm-frontend-generator/internal/gateway/types"
)

func TestPromptService_BuildSystemPrompt(t *testing.T) {
	service := NewPromptService()

	// Current implementation uses HTML + Alpine.js mode for all frameworks
	tests := []struct {
		framework string
		expected  []string
	}{
		{"react", []string{"HTML", "Alpine.js", "核心能力"}},
		{"vue", []string{"HTML", "Alpine.js", "核心能力"}},
		{"unknown", []string{"HTML", "Alpine.js", "核心能力"}},
	}

	for _, tt := range tests {
		result := service.BuildSystemPrompt(tt.framework)
		for _, exp := range tt.expected {
			if !strings.Contains(result, exp) {
				t.Errorf("expected system prompt to contain %s for framework %s", exp, tt.framework)
			}
		}
	}
}

func TestPromptService_BuildGeneratePrompt_SingleImage(t *testing.T) {
	service := NewPromptService()

	images := []ImageData{
		{
			ID:       "img-1",
			Filename: "test.png",
			Base64:   "data:image/png;base64,iVBORw0KGgo=",
		},
	}

	messages := service.BuildGeneratePrompt(images, "react")

	if len(messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(messages))
	}

	// Check message structure
	msg := messages[0]
	if msg.Role != types.RoleUser {
		t.Errorf("expected role user, got %s", msg.Role)
	}

	// Should have text + 1 image
	if len(msg.Content) != 2 {
		t.Errorf("expected 2 content parts, got %d", len(msg.Content))
	}

	// First part should be text
	if msg.Content[0].Type != types.ContentTypeText {
		t.Errorf("expected first part to be text, got %s", msg.Content[0].Type)
	}

	// Second part should be image
	if msg.Content[1].Type != types.ContentTypeImageURL {
		t.Errorf("expected second part to be image_url, got %s", msg.Content[1].Type)
	}

	if msg.Content[1].ImageURL.URL != "data:image/png;base64,iVBORw0KGgo=" {
		t.Errorf("unexpected image URL: %s", msg.Content[1].ImageURL.URL)
	}
}

func TestPromptService_BuildGeneratePrompt_MultipleImages(t *testing.T) {
	service := NewPromptService()

	images := []ImageData{
		{ID: "img-1", Base64: "data:image/png;base64,img1"},
		{ID: "img-2", Base64: "data:image/png;base64,img2"},
		{ID: "img-3", Base64: "data:image/png;base64,img3"},
	}

	messages := service.BuildGeneratePrompt(images, "react")

	if len(messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	// Should have text + 3 images
	if len(msg.Content) != 4 {
		t.Errorf("expected 4 content parts (1 text + 3 images), got %d", len(msg.Content))
	}

	// Check that prompt mentions the number of images
	if !strings.Contains(msg.Content[0].Text, "3") {
		t.Error("expected prompt to mention number of images")
	}
}

func TestPromptService_BuildChatPrompt(t *testing.T) {
	service := NewPromptService()

	code := "export default function App() { return <div>Hello</div> }"
	message := "把文字改成蓝色"

	history := []HistoryEntry{
		{Role: "user", Content: "之前的请求", Type: "text"},
		{Role: "assistant", Content: "之前的回复", Type: "text"},
	}

	messages := service.BuildChatPrompt(code, message, history)

	// Should have 2 history entries + 1 current request
	if len(messages) != 3 {
		t.Errorf("expected 3 messages, got %d", len(messages))
	}

	// Check history is preserved
	if messages[0].Content[0].Text != "之前的请求" {
		t.Errorf("expected first message to be history, got %s", messages[0].Content[0].Text)
	}

	// Check current request contains code and message
	lastMsg := messages[len(messages)-1]
	if !strings.Contains(lastMsg.Content[0].Text, code) {
		t.Error("expected last message to contain code")
	}
	if !strings.Contains(lastMsg.Content[0].Text, message) {
		t.Error("expected last message to contain user message")
	}
}

func TestPromptService_BuildChatPrompt_NoHistory(t *testing.T) {
	service := NewPromptService()

	code := "export default function App() { return <div>Hello</div> }"
	message := "添加一个按钮"

	messages := service.BuildChatPrompt(code, message, nil)

	if len(messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(messages))
	}
}

func TestPromptService_BuildDiffPrompt(t *testing.T) {
	service := NewPromptService()

	images := []ImageData{
		{ID: "img-1", Base64: "data:image/png;base64,before"},
		{ID: "img-2", Base64: "data:image/png;base64,after"},
	}

	messages := service.BuildDiffPrompt(images, "react")

	if len(messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	// Should have text + 2 images
	if len(msg.Content) != 3 {
		t.Errorf("expected 3 content parts, got %d", len(msg.Content))
	}

	// Check diff analysis prompt
	if !strings.Contains(msg.Content[0].Text, "对比") {
		t.Error("expected diff analysis prompt")
	}
}
