package service

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// getTestPoolForProject creates a pgxpool.Pool for project integration tests.
func getTestPoolForProject(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping ProjectService integration test")
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool
}

// createTestUser inserts a user for testing and returns the user ID.
func createTestUser(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	var userID string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO users (email, password_hash, display_name)
		 VALUES ($1, $2, $3) RETURNING id`,
		"test-"+t.Name()+"@example.com", "$2a$10$dummy", "Test User",
	).Scan(&userID)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
	})
	return userID
}

func TestProjectService_Create(t *testing.T) {
	pool := getTestPoolForProject(t)
	svc := NewProjectService(pool)
	ctx := context.Background()
	userID := createTestUser(t, pool)

	project, err := svc.Create(ctx, userID, "Test Project", "A test project")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer svc.Delete(ctx, userID, project.ID)

	if project.ID == "" {
		t.Error("expected project ID to be set")
	}
	if project.Name != "Test Project" {
		t.Errorf("expected name 'Test Project', got %q", project.Name)
	}
	if project.UserID != userID {
		t.Errorf("expected user_id %s, got %s", userID, project.UserID)
	}
}

func TestProjectService_Get(t *testing.T) {
	pool := getTestPoolForProject(t)
	svc := NewProjectService(pool)
	ctx := context.Background()
	userID := createTestUser(t, pool)

	created, _ := svc.Create(ctx, userID, "Get Test", "")
	defer svc.Delete(ctx, userID, created.ID)

	got, err := svc.Get(ctx, userID, created.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("expected ID %s, got %s", created.ID, got.ID)
	}
}

func TestProjectService_Get_WrongUser(t *testing.T) {
	pool := getTestPoolForProject(t)
	svc := NewProjectService(pool)
	ctx := context.Background()
	userID := createTestUser(t, pool)

	created, _ := svc.Create(ctx, userID, "Ownership Test", "")
	defer svc.Delete(ctx, userID, created.ID)

	_, err := svc.Get(ctx, "00000000-0000-0000-0000-000000000000", created.ID)
	if err != ErrProjectNotFound {
		t.Errorf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestProjectService_List(t *testing.T) {
	pool := getTestPoolForProject(t)
	svc := NewProjectService(pool)
	ctx := context.Background()
	userID := createTestUser(t, pool)

	p1, _ := svc.Create(ctx, userID, "Project 1", "")
	p2, _ := svc.Create(ctx, userID, "Project 2", "")
	defer svc.Delete(ctx, userID, p1.ID)
	defer svc.Delete(ctx, userID, p2.ID)

	projects, err := svc.List(ctx, userID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(projects) < 2 {
		t.Errorf("expected at least 2 projects, got %d", len(projects))
	}
}

func TestProjectService_Update(t *testing.T) {
	pool := getTestPoolForProject(t)
	svc := NewProjectService(pool)
	ctx := context.Background()
	userID := createTestUser(t, pool)

	created, _ := svc.Create(ctx, userID, "Old Name", "Old Desc")
	defer svc.Delete(ctx, userID, created.ID)

	updated, err := svc.Update(ctx, userID, created.ID, "New Name", "New Desc")
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "New Name" {
		t.Errorf("expected name 'New Name', got %q", updated.Name)
	}
	if updated.Description != "New Desc" {
		t.Errorf("expected description 'New Desc', got %q", updated.Description)
	}
}

func TestProjectService_Delete(t *testing.T) {
	pool := getTestPoolForProject(t)
	svc := NewProjectService(pool)
	ctx := context.Background()
	userID := createTestUser(t, pool)

	created, _ := svc.Create(ctx, userID, "To Delete", "")

	if err := svc.Delete(ctx, userID, created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := svc.Get(ctx, userID, created.ID)
	if err != ErrProjectNotFound {
		t.Errorf("expected ErrProjectNotFound after delete, got %v", err)
	}
}

func TestProjectService_Delete_WrongUser(t *testing.T) {
	pool := getTestPoolForProject(t)
	svc := NewProjectService(pool)
	ctx := context.Background()
	userID := createTestUser(t, pool)

	created, _ := svc.Create(ctx, userID, "Cannot Delete", "")
	defer svc.Delete(ctx, userID, created.ID)

	err := svc.Delete(ctx, "00000000-0000-0000-0000-000000000000", created.ID)
	if err != ErrProjectNotFound {
		t.Errorf("expected ErrProjectNotFound for wrong user, got %v", err)
	}
}

func TestProjectService_List_Empty(t *testing.T) {
	pool := getTestPoolForProject(t)
	svc := NewProjectService(pool)
	ctx := context.Background()
	userID := createTestUser(t, pool)

	projects, err := svc.List(ctx, userID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if projects == nil {
		t.Error("expected non-nil empty slice, got nil")
	}
	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}
