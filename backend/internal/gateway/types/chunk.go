package types

// StreamChunk represents a chunk of streaming response
type StreamChunk struct {
	Type    ChunkType `json:"type"`
	Content string    `json:"content,omitempty"`
	Error   error     `json:"-"`
}

// ChunkType defines the type of stream chunk
type ChunkType string

const (
	ChunkTypeContent ChunkType = "content" // Incremental content
	ChunkTypeError   ChunkType = "error"   // Error occurred
	ChunkTypeDone    ChunkType = "done"    // Stream completed
)

// NewContentChunk creates a content chunk
func NewContentChunk(content string) StreamChunk {
	return StreamChunk{
		Type:    ChunkTypeContent,
		Content: content,
	}
}

// NewErrorChunk creates an error chunk
func NewErrorChunk(err error) StreamChunk {
	return StreamChunk{
		Type:  ChunkTypeError,
		Error: err,
	}
}

// NewDoneChunk creates a done chunk
func NewDoneChunk() StreamChunk {
	return StreamChunk{
		Type: ChunkTypeDone,
	}
}
