package service

import (
	"multimodal-llm-frontend-generator/internal/gateway/types"
	"multimodal-llm-frontend-generator/pkg/prompt"
)

// PromptBuilder handles prompt construction using V2 optimized templates
type PromptBuilder struct {
	enableFewShot bool
}

// NewPromptBuilder creates a new PromptBuilder
func NewPromptBuilder(enableFewShot bool) *PromptBuilder {
	return &PromptBuilder{
		enableFewShot: enableFewShot,
	}
}

// BuildSystemPrompt builds the system prompt for the given framework
func (b *PromptBuilder) BuildSystemPrompt(framework string) string {
	// Use HTML mode for better stability
	return prompt.BuildSystemPromptHTML()
}

// BuildGeneratePrompt builds the prompt for code generation
func (b *PromptBuilder) BuildGeneratePrompt(images []ImageData, framework string) []types.Message {
	var messages []types.Message

	if len(images) == 1 {
		// Single image: use HTML single image prompt
		userPrompt := prompt.BuildSingleImagePromptHTML()
		messages = append(messages, buildImageMessage(userPrompt, images))
	} else {
		// Multiple images: use HTML multi-image prompt
		userPrompt := prompt.BuildMultiImagePromptHTML(len(images))
		messages = append(messages, buildImageMessage(userPrompt, images))
	}

	return messages
}

// BuildChatPrompt builds the prompt for chat-based code modification
// images: 原始设计稿图片，用于 AI 参考进行精修
func (b *PromptBuilder) BuildChatPrompt(code string, message string, history []HistoryEntry, images []ImageData) []types.Message {
	var messages []types.Message

	// Add history entries
	for _, entry := range history {
		role := entry.Role
		if role == "user" {
			role = types.RoleUser
		} else {
			role = types.RoleAssistant
		}
		messages = append(messages, types.NewTextMessage(role, entry.Content))
	}

	// Use HTML chat prompt
	userPrompt := prompt.BuildChatModifyPromptHTML(code, message)

	// If we have original design images, include them for reference
	if len(images) > 0 {
		// Build message with images for better visual reference
		messages = append(messages, buildChatMessageWithImages(userPrompt, images))
	} else {
		messages = append(messages, types.NewTextMessage(types.RoleUser, userPrompt))
	}

	return messages
}

// buildChatMessageWithImages creates a chat message with original design images for reference
func buildChatMessageWithImages(text string, images []ImageData) types.Message {
	parts := make([]types.ContentPart, 0, len(images)+1)

	// Add text part first
	parts = append(parts, types.ContentPart{
		Type: types.ContentTypeText,
		Text: text + "\n\n[以下是原始设计稿图片，请参考进行精修]",
	})

	// Add image parts
	for _, img := range images {
		parts = append(parts, types.ContentPart{
			Type: types.ContentTypeImageURL,
			ImageURL: &types.ImageURL{
				URL:    img.Base64, // Already in data:image/...;base64,... format
				Detail: "high",
			},
		})
	}

	return types.Message{
		Role:    types.RoleUser,
		Content: parts,
	}
}

// BuildDiffPrompt builds the prompt for diff analysis (multi-image)
func (b *PromptBuilder) BuildDiffPrompt(images []ImageData, framework string) []types.Message {
	var messages []types.Message

	if len(images) >= 2 {
		// Use HTML diff analysis prompt with 5-step framework
		userPrompt := prompt.BuildDiffAnalysisPromptHTML()
		messages = append(messages, buildImageMessage(userPrompt, images[:2]))
	}

	return messages
}

// buildImageMessage creates a message with images and text
func buildImageMessage(text string, images []ImageData) types.Message {
	parts := make([]types.ContentPart, 0, len(images)+1)

	// Add text part
	parts = append(parts, types.ContentPart{
		Type: types.ContentTypeText,
		Text: text,
	})

	// Add image parts
	for _, img := range images {
		parts = append(parts, types.ContentPart{
			Type: types.ContentTypeImageURL,
			ImageURL: &types.ImageURL{
				URL:    img.Base64, // Already in data:image/...;base64,... format
				Detail: "high",
			},
		})
	}

	return types.Message{
		Role:    types.RoleUser,
		Content: parts,
	}
}
