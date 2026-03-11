package service

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// getTestPool creates a pgxpool.Pool for integration tests.
// Skips the test if DATABASE_URL is not set.
func getTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping PostgresStore integration test")
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool
}

func TestPostgresStore_CreateAndGet(t *testing.T) {
	pool := getTestPool(t)
	store := NewPostgresStore(pool, 30*time.Minute, 20)

	ctx := context.Background()

	session, err := store.Create(ctx)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if session.ID == "" {
		t.Error("expected session ID to be set")
	}

	got, err := store.Get(ctx, session.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != session.ID {
		t.Errorf("expected ID %s, got %s", session.ID, got.ID)
	}

	// Cleanup
	_ = store.Delete(ctx, session.ID)
}

func TestPostgresStore_Get_NotFound(t *testing.T) {
	pool := getTestPool(t)
	store := NewPostgresStore(pool, 30*time.Minute, 20)

	_, err := store.Get(context.Background(), "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected error for non-existent session")
	}
	if _, ok := err.(*ErrSessionNotFound); !ok {
		t.Errorf("expected ErrSessionNotFound, got %T: %v", err, err)
	}
}

func TestPostgresStore_AddImage(t *testing.T) {
	pool := getTestPool(t)
	store := NewPostgresStore(pool, 30*time.Minute, 20)
	ctx := context.Background()

	session, err := store.Create(ctx)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer store.Delete(ctx, session.ID)

	img := &ImageData{
		ID:       "img-pg-1",
		Filename: "test.png",
		MimeType: "image/png",
		Base64:   "iVBORw0KGgo=",
	}
	if err := store.AddImage(ctx, session.ID, img); err != nil {
		t.Fatalf("AddImage: %v", err)
	}

	got, err := store.Get(ctx, session.ID)
	if err != nil {
		t.Fatalf("Get after AddImage: %v", err)
	}
	if len(got.Images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(got.Images))
	}
	if got.Images[0].ID != "img-pg-1" {
		t.Errorf("expected image ID img-pg-1, got %s", got.Images[0].ID)
	}
	if got.Images[0].Order != 0 {
		t.Errorf("expected order 0, got %d", got.Images[0].Order)
	}
}

func TestPostgresStore_GetImages(t *testing.T) {
	pool := getTestPool(t)
	store := NewPostgresStore(pool, 30*time.Minute, 20)
	ctx := context.Background()

	session, _ := store.Create(ctx)
	defer store.Delete(ctx, session.ID)

	store.AddImage(ctx, session.ID, &ImageData{ID: "img-a", Filename: "a.png", MimeType: "image/png", Base64: "a"})
	store.AddImage(ctx, session.ID, &ImageData{ID: "img-b", Filename: "b.png", MimeType: "image/png", Base64: "b"})
	store.AddImage(ctx, session.ID, &ImageData{ID: "img-c", Filename: "c.png", MimeType: "image/png", Base64: "c"})

	images, err := store.GetImages(ctx, session.ID, []string{"img-a", "img-c"})
	if err != nil {
		t.Fatalf("GetImages: %v", err)
	}
	if len(images) != 2 {
		t.Fatalf("expected 2 images, got %d", len(images))
	}
	if images[0].ID != "img-a" {
		t.Errorf("expected first image img-a, got %s", images[0].ID)
	}
	if images[1].ID != "img-c" {
		t.Errorf("expected second image img-c, got %s", images[1].ID)
	}
}

func TestPostgresStore_UpdateCode(t *testing.T) {
	pool := getTestPool(t)
	store := NewPostgresStore(pool, 30*time.Minute, 20)
	ctx := context.Background()

	session, _ := store.Create(ctx)
	defer store.Delete(ctx, session.ID)

	code := "<div>Hello World</div>"
	if err := store.UpdateCode(ctx, session.ID, code); err != nil {
		t.Fatalf("UpdateCode: %v", err)
	}

	got, _ := store.Get(ctx, session.ID)
	if got.Code != code {
		t.Errorf("expected code %q, got %q", code, got.Code)
	}
}

func TestPostgresStore_AddAndGetHistory(t *testing.T) {
	pool := getTestPool(t)
	store := NewPostgresStore(pool, 30*time.Minute, 20)
	ctx := context.Background()

	session, _ := store.Create(ctx)
	defer store.Delete(ctx, session.ID)

	entries := []HistoryEntry{
		{Role: "user", Content: "Make button blue", Type: "text"},
		{Role: "assistant", Content: "Done", Type: "text"},
		{Role: "user", Content: "Add shadow", Type: "text"},
	}
	for _, e := range entries {
		if err := store.AddHistory(ctx, session.ID, e); err != nil {
			t.Fatalf("AddHistory: %v", err)
		}
	}

	history, err := store.GetHistory(ctx, session.ID, 2)
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(history))
	}
	// Should be chronological (oldest first among the last 2)
	if history[0].Content != "Done" {
		t.Errorf("expected first entry 'Done', got %q", history[0].Content)
	}
	if history[1].Content != "Add shadow" {
		t.Errorf("expected second entry 'Add shadow', got %q", history[1].Content)
	}
}

func TestPostgresStore_Update(t *testing.T) {
	pool := getTestPool(t)
	store := NewPostgresStore(pool, 30*time.Minute, 20)
	ctx := context.Background()

	session, _ := store.Create(ctx)
	defer store.Delete(ctx, session.ID)

	session.Code = "<div>Updated</div>"
	session.Framework = "vue"
	if err := store.Update(ctx, session); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, _ := store.Get(ctx, session.ID)
	if got.Code != "<div>Updated</div>" {
		t.Errorf("expected updated code, got %q", got.Code)
	}
	if got.Framework != "vue" {
		t.Errorf("expected framework vue, got %q", got.Framework)
	}
}

func TestPostgresStore_Delete(t *testing.T) {
	pool := getTestPool(t)
	store := NewPostgresStore(pool, 30*time.Minute, 20)
	ctx := context.Background()

	session, _ := store.Create(ctx)

	if err := store.Delete(ctx, session.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := store.Get(ctx, session.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestPostgresStore_SaveAndGetDesignState(t *testing.T) {
	pool := getTestPool(t)
	store := NewPostgresStore(pool, 30*time.Minute, 20)
	ctx := context.Background()

	session, _ := store.Create(ctx)
	defer store.Delete(ctx, session.ID)

	stateJSON := []byte(`{"session_id":"` + session.ID + `","completed_agents":["layout_analyzer"],"success":false}`)

	if err := store.SaveDesignState(ctx, session.ID, stateJSON); err != nil {
		t.Fatalf("SaveDesignState: %v", err)
	}

	got, err := store.GetDesignState(ctx, session.ID)
	if err != nil {
		t.Fatalf("GetDesignState: %v", err)
	}
	if string(got) == "" {
		t.Error("expected non-empty design state")
	}
}

func TestPostgresStore_HistoryLimit(t *testing.T) {
	pool := getTestPool(t)
	store := NewPostgresStore(pool, 30*time.Minute, 3) // limit = 3
	ctx := context.Background()

	session, _ := store.Create(ctx)
	defer store.Delete(ctx, session.ID)

	// Add 5 entries, expect only 3 to remain
	for i := range 5 {
		store.AddHistory(ctx, session.ID, HistoryEntry{
			Role:    "user",
			Content: string(rune('A' + i)),
			Type:    "text",
		})
	}

	history, _ := store.GetHistory(ctx, session.ID, 10)
	if len(history) != 3 {
		t.Errorf("expected 3 entries after trimming, got %d", len(history))
	}
}
