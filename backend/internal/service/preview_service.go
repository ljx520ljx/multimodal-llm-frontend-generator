package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

// SharedPreview represents a shared preview record
type SharedPreview struct {
	ID          string
	SessionID   string
	ShortCode   string
	HTMLSnapshot string
	ViewCount   int
	IsActive    bool
	ExpiresAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PreviewService manages shared preview links
type PreviewService struct {
	pool         *pgxpool.Pool
	sessionStore SessionStore
}

// NewPreviewService creates a new preview service
func NewPreviewService(pool *pgxpool.Pool, sessionStore SessionStore) *PreviewService {
	return &PreviewService{pool: pool, sessionStore: sessionStore}
}

// CreateShare creates a shareable preview link for a session
func (s *PreviewService) CreateShare(ctx context.Context, sessionID string) (*SharedPreview, error) {
	// Get session code
	session, err := s.sessionStore.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session.Code == "" {
		return nil, fmt.Errorf("session has no generated code")
	}

	// Wrap HTML if needed
	html := wrapHTML(session.Code)

	// Generate short code with collision retry
	var shortCode string
	for attempt := 0; attempt < 3; attempt++ {
		shortCode, err = gonanoid.New(8)
		if err != nil {
			return nil, fmt.Errorf("failed to generate short code: %w", err)
		}

		// Check for collision
		var exists bool
		err = s.pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM shared_previews WHERE short_code = $1)`,
			shortCode,
		).Scan(&exists)
		if err != nil {
			return nil, err
		}
		if !exists {
			break
		}
		if attempt == 2 {
			return nil, fmt.Errorf("failed to generate unique short code after 3 attempts")
		}
	}

	var preview SharedPreview
	err = s.pool.QueryRow(ctx,
		`INSERT INTO shared_previews (session_id, short_code, html_snapshot)
		 VALUES ($1, $2, $3)
		 RETURNING id, session_id, short_code, html_snapshot, view_count, is_active, expires_at, created_at, updated_at`,
		sessionID, shortCode, html,
	).Scan(
		&preview.ID, &preview.SessionID, &preview.ShortCode,
		&preview.HTMLSnapshot, &preview.ViewCount, &preview.IsActive,
		&preview.ExpiresAt, &preview.CreatedAt, &preview.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &preview, nil
}

// GetByShortCode retrieves a shared preview by its short code
func (s *PreviewService) GetByShortCode(ctx context.Context, code string) (*SharedPreview, error) {
	var preview SharedPreview
	err := s.pool.QueryRow(ctx,
		`SELECT id, session_id, short_code, html_snapshot, view_count, is_active, expires_at, created_at, updated_at
		 FROM shared_previews
		 WHERE short_code = $1 AND is_active = TRUE
		   AND (expires_at IS NULL OR expires_at > NOW())`,
		code,
	).Scan(
		&preview.ID, &preview.SessionID, &preview.ShortCode,
		&preview.HTMLSnapshot, &preview.ViewCount, &preview.IsActive,
		&preview.ExpiresAt, &preview.CreatedAt, &preview.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("preview not found or expired")
		}
		return nil, err
	}

	// Async increment view count
	go func() {
		_, _ = s.pool.Exec(context.Background(),
			`UPDATE shared_previews SET view_count = view_count + 1 WHERE id = $1`,
			preview.ID,
		)
	}()

	return &preview, nil
}

// UpdateShare updates the HTML snapshot for an existing share
func (s *PreviewService) UpdateShare(ctx context.Context, sessionID string, shortCode string) error {
	// Get fresh code from session
	session, err := s.sessionStore.Get(ctx, sessionID)
	if err != nil {
		return err
	}
	if session.Code == "" {
		return fmt.Errorf("session has no generated code")
	}

	html := wrapHTML(session.Code)

	tag, err := s.pool.Exec(ctx,
		`UPDATE shared_previews SET html_snapshot = $1, updated_at = NOW()
		 WHERE short_code = $2 AND session_id = $3 AND is_active = TRUE`,
		html, shortCode, sessionID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("share link not found")
	}
	return nil
}

// DeleteShare deactivates a share link
func (s *PreviewService) DeleteShare(ctx context.Context, shortCode string) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE shared_previews SET is_active = FALSE, updated_at = NOW()
		 WHERE short_code = $1 AND is_active = TRUE`,
		shortCode,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("share link not found")
	}
	return nil
}

// wrapHTML ensures the code is a complete HTML document with required CDN scripts
func wrapHTML(code string) string {
	trimmed := strings.TrimSpace(code)
	if strings.HasPrefix(strings.ToLower(trimmed), "<!doctype") ||
		strings.HasPrefix(strings.ToLower(trimmed), "<html") {
		return trimmed
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Interactive Prototype</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body>
%s
</body>
</html>`, trimmed)
}
