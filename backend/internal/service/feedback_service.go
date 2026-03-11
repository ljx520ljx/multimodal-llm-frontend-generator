package service

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// UserFeedback represents feedback on a generated code version
type UserFeedback struct {
	ID            string    `json:"id"`
	SessionID     string    `json:"session_id"`
	UserID        *string   `json:"user_id,omitempty"`
	CodeVersionID *string   `json:"code_version_id,omitempty"`
	Rating        int       `json:"rating"` // 1-5
	FeedbackText  *string   `json:"feedback_text,omitempty"`
	FeedbackType  string    `json:"feedback_type"` // general, visual, interaction, code_quality
	CreatedAt     time.Time `json:"created_at"`
}

// FeedbackStats holds aggregate feedback statistics
type FeedbackStats struct {
	TotalCount   int     `json:"total_count"`
	AverageRating float64 `json:"average_rating"`
}

// ErrInvalidRating is returned when the rating is out of range
var ErrInvalidRating = errors.New("rating must be between 1 and 5")

// FeedbackService manages user feedback
type FeedbackService struct {
	pool *pgxpool.Pool
}

// NewFeedbackService creates a new feedback service
func NewFeedbackService(pool *pgxpool.Pool) *FeedbackService {
	return &FeedbackService{pool: pool}
}

// Submit records a new feedback entry
func (s *FeedbackService) Submit(ctx context.Context, fb *UserFeedback) (*UserFeedback, error) {
	if fb.Rating < 1 || fb.Rating > 5 {
		return nil, ErrInvalidRating
	}
	if fb.FeedbackType == "" {
		fb.FeedbackType = "general"
	}

	err := s.pool.QueryRow(ctx,
		`INSERT INTO user_feedback (session_id, user_id, code_version_id, rating, feedback_text, feedback_type)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at`,
		fb.SessionID, fb.UserID, fb.CodeVersionID, fb.Rating, fb.FeedbackText, fb.FeedbackType,
	).Scan(&fb.ID, &fb.CreatedAt)
	if err != nil {
		return nil, err
	}

	return fb, nil
}

// ListBySession returns all feedback for a session
func (s *FeedbackService) ListBySession(ctx context.Context, sessionID string) ([]UserFeedback, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, session_id, user_id, code_version_id, rating, feedback_text, feedback_type, created_at
		 FROM user_feedback WHERE session_id = $1
		 ORDER BY created_at DESC`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feedbacks []UserFeedback
	for rows.Next() {
		var fb UserFeedback
		if err := rows.Scan(&fb.ID, &fb.SessionID, &fb.UserID, &fb.CodeVersionID, &fb.Rating, &fb.FeedbackText, &fb.FeedbackType, &fb.CreatedAt); err != nil {
			return nil, err
		}
		feedbacks = append(feedbacks, fb)
	}

	if feedbacks == nil {
		feedbacks = []UserFeedback{}
	}
	return feedbacks, rows.Err()
}

// GetStats returns aggregate feedback statistics for a session
func (s *FeedbackService) GetStats(ctx context.Context, sessionID string) (*FeedbackStats, error) {
	var stats FeedbackStats
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*), COALESCE(AVG(rating)::numeric(3,2), 0)
		 FROM user_feedback WHERE session_id = $1`,
		sessionID,
	).Scan(&stats.TotalCount, &stats.AverageRating)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}
