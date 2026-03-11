package service

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CodeVersion represents a version of generated code for a session
type CodeVersion struct {
	ID              string    `json:"id"`
	SessionID       string    `json:"session_id"`
	VersionNumber   int       `json:"version_number"`
	HTMLCode        string    `json:"html_code"`
	DiffFromPrev    *string   `json:"diff_from_previous,omitempty"`
	Source          string    `json:"source"` // "generate" or "chat"
	TriggerMessage  *string   `json:"trigger_message,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// CodeVersionService manages code version history
type CodeVersionService struct {
	pool *pgxpool.Pool
}

// NewCodeVersionService creates a new code version service
func NewCodeVersionService(pool *pgxpool.Pool) *CodeVersionService {
	return &CodeVersionService{pool: pool}
}

// Create records a new code version
func (s *CodeVersionService) Create(ctx context.Context, sessionID, htmlCode, source string, triggerMessage *string) (*CodeVersion, error) {
	// Get the next version number
	var nextVersion int
	err := s.pool.QueryRow(ctx,
		`SELECT COALESCE(MAX(version_number), 0) + 1 FROM code_versions WHERE session_id = $1`,
		sessionID,
	).Scan(&nextVersion)
	if err != nil {
		return nil, err
	}

	var cv CodeVersion
	err = s.pool.QueryRow(ctx,
		`INSERT INTO code_versions (session_id, version_number, html_code, source, trigger_message)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, session_id, version_number, html_code, source, trigger_message, created_at`,
		sessionID, nextVersion, htmlCode, source, triggerMessage,
	).Scan(&cv.ID, &cv.SessionID, &cv.VersionNumber, &cv.HTMLCode, &cv.Source, &cv.TriggerMessage, &cv.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &cv, nil
}

// ListBySession returns all code versions for a session, ordered by version number
func (s *CodeVersionService) ListBySession(ctx context.Context, sessionID string) ([]CodeVersion, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, session_id, version_number, source, trigger_message, created_at
		 FROM code_versions WHERE session_id = $1
		 ORDER BY version_number ASC`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []CodeVersion
	for rows.Next() {
		var cv CodeVersion
		if err := rows.Scan(&cv.ID, &cv.SessionID, &cv.VersionNumber, &cv.Source, &cv.TriggerMessage, &cv.CreatedAt); err != nil {
			return nil, err
		}
		versions = append(versions, cv)
	}

	if versions == nil {
		versions = []CodeVersion{}
	}
	return versions, rows.Err()
}

// GetByID returns a specific code version with its full HTML code
func (s *CodeVersionService) GetByID(ctx context.Context, id string) (*CodeVersion, error) {
	var cv CodeVersion
	err := s.pool.QueryRow(ctx,
		`SELECT id, session_id, version_number, html_code, source, trigger_message, created_at
		 FROM code_versions WHERE id = $1`,
		id,
	).Scan(&cv.ID, &cv.SessionID, &cv.VersionNumber, &cv.HTMLCode, &cv.Source, &cv.TriggerMessage, &cv.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrCodeVersionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

// ErrCodeVersionNotFound is returned when a code version is not found
var ErrCodeVersionNotFound = &ErrSessionNotFound{ID: "code_version"}
