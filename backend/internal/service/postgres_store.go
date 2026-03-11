package service

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStore implements SessionStore using PostgreSQL
type PostgresStore struct {
	pool         *pgxpool.Pool
	historyLimit int
	ttl          time.Duration
}

// NewPostgresStore creates a new PostgreSQL-backed session store
func NewPostgresStore(pool *pgxpool.Pool, ttl time.Duration, historyLimit int) *PostgresStore {
	if historyLimit <= 0 {
		historyLimit = 20
	}
	return &PostgresStore{
		pool:         pool,
		historyLimit: historyLimit,
		ttl:          ttl,
	}
}

func (s *PostgresStore) Create(ctx context.Context) (*Session, error) {
	expiresAt := time.Now().Add(s.ttl)
	var session Session
	err := s.pool.QueryRow(ctx,
		`INSERT INTO sessions (expires_at) VALUES ($1)
		 RETURNING id, code, framework, created_at, updated_at`,
		expiresAt,
	).Scan(&session.ID, &session.Code, &session.Framework, &session.CreatedAt, &session.UpdatedAt)
	if err != nil {
		return nil, err
	}
	session.Images = make([]ImageData, 0)
	session.History = make([]HistoryEntry, 0)
	return &session, nil
}

func (s *PostgresStore) Get(ctx context.Context, id string) (*Session, error) {
	// Refresh expiration on access
	var session Session
	err := s.pool.QueryRow(ctx,
		`UPDATE sessions SET expires_at = $1
		 WHERE id = $2 AND expires_at > NOW()
		 RETURNING id, code, framework, created_at, updated_at`,
		time.Now().Add(s.ttl), id,
	).Scan(&session.ID, &session.Code, &session.Framework, &session.CreatedAt, &session.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, &ErrSessionNotFound{ID: id}
		}
		return nil, err
	}

	// Load images
	images, err := s.loadImages(ctx, id)
	if err != nil {
		return nil, err
	}
	session.Images = images

	// Load history (most recent, ordered chronologically)
	history, err := s.loadHistory(ctx, id, s.historyLimit)
	if err != nil {
		return nil, err
	}
	session.History = history

	return &session, nil
}

func (s *PostgresStore) Update(ctx context.Context, session *Session) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE sessions SET code = $1, framework = $2, updated_at = NOW(), expires_at = $3
		 WHERE id = $4 AND expires_at > NOW()`,
		session.Code, session.Framework, time.Now().Add(s.ttl), session.ID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return &ErrSessionNotFound{ID: session.ID}
	}
	return nil
}

func (s *PostgresStore) Delete(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM sessions WHERE id = $1`, id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return &ErrSessionNotFound{ID: id}
	}
	return nil
}

func (s *PostgresStore) AddImage(ctx context.Context, sessionID string, image *ImageData) error {
	// Direct INSERT - let FK constraint catch missing sessions (avoids TOCTOU race)
	_, err := s.pool.Exec(ctx,
		`INSERT INTO session_images (id, session_id, filename, mime_type, base64_data, sort_order)
		 VALUES ($1, $2, $3, $4, $5,
		   (SELECT COALESCE(MAX(sort_order), -1) + 1 FROM session_images WHERE session_id = $2))`,
		image.ID, sessionID, image.Filename, image.MimeType, image.Base64,
	)
	if err != nil {
		if isForeignKeyViolation(err) {
			return &ErrSessionNotFound{ID: sessionID}
		}
		return err
	}

	// Update session timestamp
	_, err = s.pool.Exec(ctx,
		`UPDATE sessions SET updated_at = NOW(), expires_at = $1 WHERE id = $2`,
		time.Now().Add(s.ttl), sessionID,
	)
	return err
}

func (s *PostgresStore) GetImages(ctx context.Context, sessionID string, imageIDs []string) ([]ImageData, error) {
	// Verify session exists
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM sessions WHERE id = $1 AND expires_at > NOW())`,
		sessionID,
	).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ErrSessionNotFound{ID: sessionID}
	}

	rows, err := s.pool.Query(ctx,
		`SELECT id, filename, mime_type, base64_data, sort_order
		 FROM session_images
		 WHERE session_id = $1 AND id = ANY($2)
		 ORDER BY sort_order`,
		sessionID, imageIDs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	imageMap := make(map[string]ImageData)
	for rows.Next() {
		var img ImageData
		if err := rows.Scan(&img.ID, &img.Filename, &img.MimeType, &img.Base64, &img.Order); err != nil {
			return nil, err
		}
		imageMap[img.ID] = img
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Return in requested order, verify all found
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

func (s *PostgresStore) UpdateCode(ctx context.Context, sessionID string, code string) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE sessions SET code = $1, updated_at = NOW(), expires_at = $2
		 WHERE id = $3 AND expires_at > NOW()`,
		code, time.Now().Add(s.ttl), sessionID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return &ErrSessionNotFound{ID: sessionID}
	}
	return nil
}

func (s *PostgresStore) AddHistory(ctx context.Context, sessionID string, entry HistoryEntry) error {
	// Direct INSERT - let FK constraint catch missing sessions (avoids TOCTOU race)
	_, err := s.pool.Exec(ctx,
		`INSERT INTO session_history (session_id, role, content, entry_type)
		 VALUES ($1, $2, $3, $4)`,
		sessionID, entry.Role, entry.Content, entry.Type,
	)
	if err != nil {
		if isForeignKeyViolation(err) {
			return &ErrSessionNotFound{ID: sessionID}
		}
		return err
	}

	// Trim old entries beyond historyLimit
	_, err = s.pool.Exec(ctx,
		`DELETE FROM session_history
		 WHERE session_id = $1 AND id NOT IN (
			SELECT id FROM session_history
			WHERE session_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		 )`,
		sessionID, s.historyLimit,
	)
	if err != nil {
		return err
	}

	// Update session timestamp
	_, err = s.pool.Exec(ctx,
		`UPDATE sessions SET updated_at = NOW(), expires_at = $1 WHERE id = $2`,
		time.Now().Add(s.ttl), sessionID,
	)
	return err
}

func (s *PostgresStore) GetHistory(ctx context.Context, sessionID string, limit int) ([]HistoryEntry, error) {
	// Verify session exists
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM sessions WHERE id = $1 AND expires_at > NOW())`,
		sessionID,
	).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ErrSessionNotFound{ID: sessionID}
	}

	return s.loadHistory(ctx, sessionID, limit)
}

// SaveDesignState saves the pipeline design state as JSONB
func (s *PostgresStore) SaveDesignState(ctx context.Context, sessionID string, stateJSON []byte) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE sessions SET design_state = $1, updated_at = NOW(), expires_at = $2
		 WHERE id = $3 AND expires_at > NOW()`,
		stateJSON, time.Now().Add(s.ttl), sessionID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return &ErrSessionNotFound{ID: sessionID}
	}
	return nil
}

// GetDesignState retrieves the pipeline design state
func (s *PostgresStore) GetDesignState(ctx context.Context, sessionID string) ([]byte, error) {
	var stateJSON []byte
	err := s.pool.QueryRow(ctx,
		`SELECT design_state FROM sessions WHERE id = $1 AND expires_at > NOW()`,
		sessionID,
	).Scan(&stateJSON)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, &ErrSessionNotFound{ID: sessionID}
		}
		return nil, err
	}
	return stateJSON, nil
}

// isForeignKeyViolation checks if a PostgreSQL error is a FK constraint violation (23503)
func isForeignKeyViolation(err error) bool {
	if err == nil {
		return false
	}
	// pgx wraps pgconn.PgError; check the SQLSTATE code
	msg := err.Error()
	return strings.Contains(msg, "23503") || strings.Contains(msg, "foreign key constraint")
}

func (s *PostgresStore) Close() {
	// Pool lifecycle is managed by the App, not by PostgresStore.
	// This is a no-op to satisfy the SessionStore interface.
}

// loadImages loads all images for a session
func (s *PostgresStore) loadImages(ctx context.Context, sessionID string) ([]ImageData, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, filename, mime_type, base64_data, sort_order
		 FROM session_images
		 WHERE session_id = $1
		 ORDER BY sort_order`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	images := make([]ImageData, 0)
	for rows.Next() {
		var img ImageData
		if err := rows.Scan(&img.ID, &img.Filename, &img.MimeType, &img.Base64, &img.Order); err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, rows.Err()
}

// loadHistory loads history entries ordered chronologically
func (s *PostgresStore) loadHistory(ctx context.Context, sessionID string, limit int) ([]HistoryEntry, error) {
	if limit <= 0 {
		limit = s.historyLimit
	}

	// Get most recent entries, then reverse to chronological order
	rows, err := s.pool.Query(ctx,
		`SELECT role, content, entry_type FROM (
			SELECT role, content, entry_type, created_at
			FROM session_history
			WHERE session_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		) sub ORDER BY created_at ASC`,
		sessionID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := make([]HistoryEntry, 0)
	for rows.Next() {
		var entry HistoryEntry
		if err := rows.Scan(&entry.Role, &entry.Content, &entry.Type); err != nil {
			return nil, err
		}
		history = append(history, entry)
	}
	return history, rows.Err()
}
