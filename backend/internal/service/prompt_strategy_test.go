package service

import (
	"strings"
	"testing"
)

func TestNewPromptBuilder(t *testing.T) {
	builder := NewPromptBuilder(false)
	if builder == nil {
		t.Error("NewPromptBuilder should not return nil")
	}

	builderWithFewShot := NewPromptBuilder(true)
	if builderWithFewShot == nil {
		t.Error("NewPromptBuilder with few-shot should not return nil")
	}
}

func TestPromptBuilder_BuildSystemPrompt(t *testing.T) {
	builder := NewPromptBuilder(false)

	// Current implementation uses HTML + Alpine.js mode for all frameworks
	tests := []struct {
		framework string
		contains  []string
	}{
		{
			framework: "react",
			contains:  []string{"HTML", "Alpine.js", "核心能力", "技术栈"},
		},
		{
			framework: "vue",
			contains:  []string{"HTML", "Alpine.js", "核心能力", "技术栈"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.framework, func(t *testing.T) {
			prompt := builder.BuildSystemPrompt(tt.framework)
			for _, substr := range tt.contains {
				if !strings.Contains(prompt, substr) {
					t.Errorf("System prompt for %s should contain %q", tt.framework, substr)
				}
			}
		})
	}
}

func TestPromptBuilder_BuildGeneratePrompt_SingleImage(t *testing.T) {
	builder := NewPromptBuilder(false)
	images := []ImageData{
		{ID: "img1", Base64: "data:image/png;base64,abc123"},
	}

	messages := builder.BuildGeneratePrompt(images, "react")

	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
}

func TestPromptBuilder_BuildGeneratePrompt_MultiImage(t *testing.T) {
	builder := NewPromptBuilder(false)
	images := []ImageData{
		{ID: "img1", Base64: "data:image/png;base64,abc123"},
		{ID: "img2", Base64: "data:image/png;base64,def456"},
		{ID: "img3", Base64: "data:image/png;base64,ghi789"},
	}

	messages := builder.BuildGeneratePrompt(images, "react")

	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
}

func TestPromptBuilder_BuildGeneratePrompt_WithFewShot(t *testing.T) {
	builderWithout := NewPromptBuilder(false)
	builderWith := NewPromptBuilder(true)

	images := []ImageData{
		{ID: "img1", Base64: "data:image/png;base64,abc123"},
		{ID: "img2", Base64: "data:image/png;base64,def456"},
	}

	messagesWithout := builderWithout.BuildGeneratePrompt(images, "react")
	messagesWith := builderWith.BuildGeneratePrompt(images, "react")

	// Both should return the same number of messages
	if len(messagesWithout) != len(messagesWith) {
		t.Error("Both configurations should return same number of messages")
	}
}

func TestPromptBuilder_BuildDiffPrompt(t *testing.T) {
	builder := NewPromptBuilder(false)
	images := []ImageData{
		{ID: "img1", Base64: "data:image/png;base64,abc123"},
		{ID: "img2", Base64: "data:image/png;base64,def456"},
	}

	messages := builder.BuildDiffPrompt(images, "react")

	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
}

func TestPromptBuilder_BuildDiffPrompt_InsufficientImages(t *testing.T) {
	builder := NewPromptBuilder(false)
	images := []ImageData{
		{ID: "img1", Base64: "data:image/png;base64,abc123"},
	}

	messages := builder.BuildDiffPrompt(images, "react")

	// Should return empty when less than 2 images
	if len(messages) != 0 {
		t.Errorf("Expected 0 messages for insufficient images, got %d", len(messages))
	}
}

func TestPromptBuilder_BuildChatPrompt(t *testing.T) {
	builder := NewPromptBuilder(false)
	code := "export default function App() { return <div>Hello</div> }"
	message := "把文字改成红色"
	history := []HistoryEntry{}

	messages := builder.BuildChatPrompt(code, message, history)

	if len(messages) != 1 {
		t.Errorf("Expected 1 message (no history), got %d", len(messages))
	}
}

func TestPromptBuilder_BuildChatPrompt_WithHistory(t *testing.T) {
	builder := NewPromptBuilder(false)
	code := "export default function App() { return <div>Hello</div> }"
	message := "把文字改成红色"
	history := []HistoryEntry{
		{Role: "user", Content: "生成代码", Type: "text"},
		{Role: "assistant", Content: code, Type: "code"},
	}

	messages := builder.BuildChatPrompt(code, message, history)

	// Should include history + current message
	if len(messages) != 3 {
		t.Errorf("Expected 3 messages (2 history + 1 current), got %d", len(messages))
	}
}

func TestBuildImageMessage(t *testing.T) {
	images := []ImageData{
		{ID: "img1", Base64: "data:image/png;base64,abc123"},
		{ID: "img2", Base64: "data:image/png;base64,def456"},
	}
	text := "Test prompt"

	message := buildImageMessage(text, images)

	// Should have 1 text part + 2 image parts
	if len(message.Content) != 3 {
		t.Errorf("Expected 3 content parts, got %d", len(message.Content))
	}
}
