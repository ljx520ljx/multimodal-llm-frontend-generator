package service

import "context"

// SessionStore defines the interface for session storage
type SessionStore interface {
	// Create creates a new session and returns it
	Create(ctx context.Context) (*Session, error)

	// Get retrieves a session by ID
	Get(ctx context.Context, id string) (*Session, error)

	// Update updates an existing session
	Update(ctx context.Context, session *Session) error

	// Delete removes a session
	Delete(ctx context.Context, id string) error

	// AddImage adds an image to a session
	AddImage(ctx context.Context, sessionID string, image *ImageData) error

	// GetImages retrieves specific images from a session
	GetImages(ctx context.Context, sessionID string, imageIDs []string) ([]ImageData, error)

	// UpdateCode updates the generated code in a session
	UpdateCode(ctx context.Context, sessionID string, code string) error

	// AddHistory adds a conversation entry to the session history
	AddHistory(ctx context.Context, sessionID string, entry HistoryEntry) error

	// GetHistory retrieves the conversation history (with optional limit)
	GetHistory(ctx context.Context, sessionID string, limit int) ([]HistoryEntry, error)

	// Close stops any background goroutines
	Close()
}

// ErrSessionNotFound is returned when a session is not found
type ErrSessionNotFound struct {
	ID string
}

func (e *ErrSessionNotFound) Error() string {
	return "session not found: " + e.ID
}

// ErrImageNotFound is returned when an image is not found
type ErrImageNotFound struct {
	ID string
}

func (e *ErrImageNotFound) Error() string {
	return "image not found: " + e.ID
}
