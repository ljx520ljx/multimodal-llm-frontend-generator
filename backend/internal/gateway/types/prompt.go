package types

// ChatRequest represents a request to the LLM API
type ChatRequest struct {
	Model    string       `json:"model"`
	Messages []Message    `json:"messages"`
	Options  *ChatOptions `json:"options,omitempty"`
}

// ChatOptions contains optional parameters for the chat request
type ChatOptions struct {
	Temperature      float32 `json:"temperature,omitempty"`
	MaxTokens        int     `json:"max_tokens,omitempty"`
	TopP             float32 `json:"top_p,omitempty"`
	FrequencyPenalty float32 `json:"frequency_penalty,omitempty"`
	PresencePenalty  float32 `json:"presence_penalty,omitempty"`
}

// Message represents a single message in the conversation
type Message struct {
	Role    string        `json:"role"` // system, user, assistant
	Content []ContentPart `json:"content"`
}

// NewTextMessage creates a message with text content only
func NewTextMessage(role, text string) Message {
	return Message{
		Role: role,
		Content: []ContentPart{
			{Type: ContentTypeText, Text: text},
		},
	}
}

// NewImageMessage creates a message with image and text content
func NewImageMessage(role, text, imageURL string) Message {
	return Message{
		Role: role,
		Content: []ContentPart{
			{Type: ContentTypeText, Text: text},
			{Type: ContentTypeImageURL, ImageURL: &ImageURL{URL: imageURL, Detail: "auto"}},
		},
	}
}

// ContentPart represents a part of message content (text or image)
type ContentPart struct {
	Type     ContentType `json:"type"`
	Text     string      `json:"text,omitempty"`
	ImageURL *ImageURL   `json:"image_url,omitempty"`
}

// ContentType defines the type of content
type ContentType string

const (
	ContentTypeText     ContentType = "text"
	ContentTypeImageURL ContentType = "image_url"
)

// ImageURL contains image URL information
type ImageURL struct {
	URL    string `json:"url"`    // data:image/jpeg;base64,... or https://...
	Detail string `json:"detail"` // low, high, auto
}

// Role constants
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)
