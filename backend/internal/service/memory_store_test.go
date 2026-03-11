package service

import (
	"context"
	"testing"
	"time"
)

func TestMemoryStore_Create(t *testing.T) {
	store := NewMemoryStore(30*time.Minute, 20)
	defer store.Close()

	ctx := context.Background()
	session, err := store.Create(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if session.ID == "" {
		t.Error("expected session ID to be set")
	}
	if session.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if session.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestMemoryStore_Get(t *testing.T) {
	store := NewMemoryStore(30*time.Minute, 20)
	defer store.Close()

	ctx := context.Background()

	// Create a session
	session, _ := store.Create(ctx)

	// Get the session
	retrieved, err := store.Get(ctx, session.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if retrieved.ID != session.ID {
		t.Errorf("expected session ID %s, got %s", session.ID, retrieved.ID)
	}
}

func TestMemoryStore_Get_NotFound(t *testing.T) {
	store := NewMemoryStore(30*time.Minute, 20)
	defer store.Close()

	ctx := context.Background()

	_, err := store.Get(ctx, "non-existent-id")
	if err == nil {
		t.Error("expected error for non-existent session")
	}

	if _, ok := err.(*ErrSessionNotFound); !ok {
		t.Errorf("expected ErrSessionNotFound, got %T", err)
	}
}

func TestMemoryStore_AddImage(t *testing.T) {
	store := NewMemoryStore(30*time.Minute, 20)
	defer store.Close()

	ctx := context.Background()
	session, _ := store.Create(ctx)

	image := &ImageData{
		ID:       "img-1",
		Filename: "test.png",
		MimeType: "image/png",
		Base64:   "data:image/png;base64,iVBORw0KGgo=",
	}

	err := store.AddImage(ctx, session.ID, image)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify image was added
	retrieved, _ := store.Get(ctx, session.ID)
	if len(retrieved.Images) != 1 {
		t.Errorf("expected 1 image, got %d", len(retrieved.Images))
	}
	if retrieved.Images[0].ID != "img-1" {
		t.Errorf("expected image ID img-1, got %s", retrieved.Images[0].ID)
	}
	if retrieved.Images[0].Order != 0 {
		t.Errorf("expected order 0, got %d", retrieved.Images[0].Order)
	}
}

func TestMemoryStore_AddImage_MultipleImages(t *testing.T) {
	store := NewMemoryStore(30*time.Minute, 20)
	defer store.Close()

	ctx := context.Background()
	session, _ := store.Create(ctx)

	for i := 0; i < 3; i++ {
		image := &ImageData{
			ID:       "img-" + string(rune('1'+i)),
			Filename: "test.png",
			MimeType: "image/png",
			Base64:   "data:image/png;base64,test",
		}
		store.AddImage(ctx, session.ID, image)
	}

	retrieved, _ := store.Get(ctx, session.ID)
	if len(retrieved.Images) != 3 {
		t.Errorf("expected 3 images, got %d", len(retrieved.Images))
	}

	// Check order
	for i, img := range retrieved.Images {
		if img.Order != i {
			t.Errorf("expected order %d, got %d", i, img.Order)
		}
	}
}

func TestMemoryStore_GetImages(t *testing.T) {
	store := NewMemoryStore(30*time.Minute, 20)
	defer store.Close()

	ctx := context.Background()
	session, _ := store.Create(ctx)

	// Add images
	store.AddImage(ctx, session.ID, &ImageData{ID: "img-1", Filename: "1.png"})
	store.AddImage(ctx, session.ID, &ImageData{ID: "img-2", Filename: "2.png"})
	store.AddImage(ctx, session.ID, &ImageData{ID: "img-3", Filename: "3.png"})

	// Get specific images
	images, err := store.GetImages(ctx, session.ID, []string{"img-1", "img-3"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(images) != 2 {
		t.Errorf("expected 2 images, got %d", len(images))
	}
	if images[0].ID != "img-1" {
		t.Errorf("expected first image ID img-1, got %s", images[0].ID)
	}
	if images[1].ID != "img-3" {
		t.Errorf("expected second image ID img-3, got %s", images[1].ID)
	}
}

func TestMemoryStore_GetImages_NotFound(t *testing.T) {
	store := NewMemoryStore(30*time.Minute, 20)
	defer store.Close()

	ctx := context.Background()
	session, _ := store.Create(ctx)

	store.AddImage(ctx, session.ID, &ImageData{ID: "img-1", Filename: "1.png"})

	_, err := store.GetImages(ctx, session.ID, []string{"img-1", "non-existent"})
	if err == nil {
		t.Error("expected error for non-existent image")
	}

	if _, ok := err.(*ErrImageNotFound); !ok {
		t.Errorf("expected ErrImageNotFound, got %T", err)
	}
}

func TestMemoryStore_UpdateCode(t *testing.T) {
	store := NewMemoryStore(30*time.Minute, 20)
	defer store.Close()

	ctx := context.Background()
	session, _ := store.Create(ctx)

	code := "export default function App() { return <div>Hello</div> }"
	err := store.UpdateCode(ctx, session.ID, code)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	retrieved, _ := store.Get(ctx, session.ID)
	if retrieved.Code != code {
		t.Errorf("expected code %s, got %s", code, retrieved.Code)
	}
}

func TestMemoryStore_AddHistory(t *testing.T) {
	store := NewMemoryStore(30*time.Minute, 20)
	defer store.Close()

	ctx := context.Background()
	session, _ := store.Create(ctx)

	entry := HistoryEntry{
		Role:    "user",
		Content: "Make the button blue",
		Type:    "text",
	}

	err := store.AddHistory(ctx, session.ID, entry)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	retrieved, _ := store.Get(ctx, session.ID)
	if len(retrieved.History) != 1 {
		t.Errorf("expected 1 history entry, got %d", len(retrieved.History))
	}
	if retrieved.History[0].Content != "Make the button blue" {
		t.Errorf("unexpected history content: %s", retrieved.History[0].Content)
	}
}

func TestMemoryStore_GetHistory_WithLimit(t *testing.T) {
	store := NewMemoryStore(30*time.Minute, 20)
	defer store.Close()

	ctx := context.Background()
	session, _ := store.Create(ctx)

	// Add 5 history entries
	for i := 0; i < 5; i++ {
		store.AddHistory(ctx, session.ID, HistoryEntry{
			Role:    "user",
			Content: "Message " + string(rune('1'+i)),
			Type:    "text",
		})
	}

	// Get last 3 entries
	history, err := store.GetHistory(ctx, session.ID, 3)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(history) != 3 {
		t.Errorf("expected 3 history entries, got %d", len(history))
	}

	// Should be the most recent entries
	if history[0].Content != "Message 3" {
		t.Errorf("expected Message 3, got %s", history[0].Content)
	}
}

func TestMemoryStore_Delete(t *testing.T) {
	store := NewMemoryStore(30*time.Minute, 20)
	defer store.Close()

	ctx := context.Background()
	session, _ := store.Create(ctx)

	err := store.Delete(ctx, session.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = store.Get(ctx, session.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestMemoryStore_Cleanup(t *testing.T) {
	// Use very short TTL for testing
	store := NewMemoryStore(10*time.Millisecond, 20)
	defer store.Close()

	ctx := context.Background()
	session, _ := store.Create(ctx)

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Trigger cleanup manually
	store.cleanup()

	_, err := store.Get(ctx, session.ID)
	if err == nil {
		t.Error("expected session to be cleaned up")
	}
}
