package service

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryStore implements SessionStore using in-memory storage
type MemoryStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	ttl      time.Duration
	done     chan struct{}
}

// NewMemoryStore creates a new in-memory session store
func NewMemoryStore(ttl time.Duration) *MemoryStore {
	store := &MemoryStore{
		sessions: make(map[string]*Session),
		ttl:      ttl,
		done:     make(chan struct{}),
	}

	// Start background cleanup goroutine
	go store.cleanupLoop()

	return store
}

// cleanupLoop periodically removes expired sessions
func (s *MemoryStore) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanup()
		case <-s.done:
			return
		}
	}
}

// cleanup removes expired sessions
func (s *MemoryStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, session := range s.sessions {
		if now.Sub(session.UpdatedAt) > s.ttl {
			delete(s.sessions, id)
		}
	}
}

// Create creates a new session
func (s *MemoryStore) Create(ctx context.Context) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	session := &Session{
		ID:        uuid.New().String(),
		Images:    make([]ImageData, 0),
		History:   make([]HistoryEntry, 0),
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.sessions[session.ID] = session
	return session, nil
}

// Get retrieves a session by ID
func (s *MemoryStore) Get(ctx context.Context, id string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[id]
	if !ok {
		return nil, &ErrSessionNotFound{ID: id}
	}

	return session, nil
}

// Update updates an existing session
func (s *MemoryStore) Update(ctx context.Context, session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.sessions[session.ID]; !ok {
		return &ErrSessionNotFound{ID: session.ID}
	}

	session.UpdatedAt = time.Now()
	s.sessions[session.ID] = session
	return nil
}

// Delete removes a session
func (s *MemoryStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.sessions[id]; !ok {
		return &ErrSessionNotFound{ID: id}
	}

	delete(s.sessions, id)
	return nil
}

// AddImage adds an image to a session
func (s *MemoryStore) AddImage(ctx context.Context, sessionID string, image *ImageData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return &ErrSessionNotFound{ID: sessionID}
	}

	// Assign order based on current image count
	image.Order = len(session.Images)
	session.Images = append(session.Images, *image)
	session.UpdatedAt = time.Now()

	return nil
}

// GetImages retrieves specific images from a session
func (s *MemoryStore) GetImages(ctx context.Context, sessionID string, imageIDs []string) ([]ImageData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, &ErrSessionNotFound{ID: sessionID}
	}

	// Create a map for quick lookup
	imageMap := make(map[string]ImageData)
	for _, img := range session.Images {
		imageMap[img.ID] = img
	}

	// Retrieve requested images in order
	result := make([]ImageData, 0, len(imageIDs))
	for _, id := range imageIDs {
		img, ok := imageMap[id]
		if !ok {
			return nil, &ErrImageNotFound{ID: id}
		}
		result = append(result, img)
	}

	return result, nil
}

// UpdateCode updates the generated code in a session
func (s *MemoryStore) UpdateCode(ctx context.Context, sessionID string, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return &ErrSessionNotFound{ID: sessionID}
	}

	session.Code = code
	session.UpdatedAt = time.Now()

	return nil
}

// AddHistory adds a conversation entry to the session history
func (s *MemoryStore) AddHistory(ctx context.Context, sessionID string, entry HistoryEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return &ErrSessionNotFound{ID: sessionID}
	}

	session.History = append(session.History, entry)
	session.UpdatedAt = time.Now()

	return nil
}

// GetHistory retrieves the conversation history with optional limit
func (s *MemoryStore) GetHistory(ctx context.Context, sessionID string, limit int) ([]HistoryEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, &ErrSessionNotFound{ID: sessionID}
	}

	history := session.History
	if limit > 0 && len(history) > limit {
		// Return the most recent entries
		history = history[len(history)-limit:]
	}

	return history, nil
}

// Close stops the background cleanup goroutine
func (s *MemoryStore) Close() {
	close(s.done)
}
