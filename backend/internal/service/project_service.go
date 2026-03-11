package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Project represents a user's project
type Project struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProjectService handles project CRUD operations
type ProjectService struct {
	pool *pgxpool.Pool
}

// NewProjectService creates a new ProjectService
func NewProjectService(pool *pgxpool.Pool) *ProjectService {
	return &ProjectService{pool: pool}
}

// Create creates a new project for a user
func (s *ProjectService) Create(ctx context.Context, userID, name, description string) (*Project, error) {
	var p Project
	err := s.pool.QueryRow(ctx,
		`INSERT INTO projects (user_id, name, description)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, name, description, created_at, updated_at`,
		userID, name, description,
	).Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	return &p, nil
}

// Get retrieves a project by ID, checking ownership
func (s *ProjectService) Get(ctx context.Context, userID, projectID string) (*Project, error) {
	var p Project
	err := s.pool.QueryRow(ctx,
		`SELECT id, user_id, name, description, created_at, updated_at
		 FROM projects WHERE id = $1 AND user_id = $2`,
		projectID, userID,
	).Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("get project: %w", err)
	}
	return &p, nil
}

// List returns all projects for a user
func (s *ProjectService) List(ctx context.Context, userID string) ([]Project, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, name, description, created_at, updated_at
		 FROM projects WHERE user_id = $1 ORDER BY updated_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, p)
	}
	if projects == nil {
		projects = []Project{}
	}
	return projects, nil
}

// Update updates a project's name and description
func (s *ProjectService) Update(ctx context.Context, userID, projectID, name, description string) (*Project, error) {
	var p Project
	err := s.pool.QueryRow(ctx,
		`UPDATE projects SET name = $1, description = $2, updated_at = NOW()
		 WHERE id = $3 AND user_id = $4
		 RETURNING id, user_id, name, description, created_at, updated_at`,
		name, description, projectID, userID,
	).Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("update project: %w", err)
	}
	return &p, nil
}

// Delete deletes a project (CASCADE deletes associated sessions)
func (s *ProjectService) Delete(ctx context.Context, userID, projectID string) error {
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM projects WHERE id = $1 AND user_id = $2`,
		projectID, userID,
	)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrProjectNotFound
	}
	return nil
}

// LinkSession associates a session with a user and optionally a project
func (s *ProjectService) LinkSession(ctx context.Context, sessionID, userID string, projectID *string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE sessions SET user_id = $1, project_id = $2 WHERE id = $3`,
		userID, projectID, sessionID,
	)
	if err != nil {
		return fmt.Errorf("link session: %w", err)
	}
	return nil
}

var ErrProjectNotFound = errors.New("project not found")
