package csrf

import (
	"context"
	"errors"
	"testing"
	"time"
)

func newTestStore() *MemoryStore {
	// Create store without cleanup goroutine for testing
	return &MemoryStore{
		tokens: make(map[string]*storedToken),
	}
}

func TestMemoryStore_Set(t *testing.T) {
	store := newTestStore()
	ctx := context.Background()

	err := store.Set(ctx, "session-1", "hash-abc", 1*time.Hour)
	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	if store.Count() != 1 {
		t.Errorf("expected count 1, got %d", store.Count())
	}
}

func TestMemoryStore_Get(t *testing.T) {
	store := newTestStore()
	ctx := context.Background()

	err := store.Set(ctx, "session-1", "hash-abc", 1*time.Hour)
	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	hash, expiresAt, err := store.Get(ctx, "session-1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if hash != "hash-abc" {
		t.Errorf("expected hash 'hash-abc', got %q", hash)
	}
	if expiresAt.IsZero() {
		t.Error("expected non-zero expiresAt")
	}
	if !expiresAt.After(time.Now()) {
		t.Error("expected expiresAt to be in the future")
	}
}

func TestMemoryStore_Get_NotFound(t *testing.T) {
	store := newTestStore()
	ctx := context.Background()

	_, _, err := store.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent session, got nil")
	}
	if !errors.Is(err, ErrCSRFTokenNotFound) {
		t.Errorf("expected ErrCSRFTokenNotFound, got %v", err)
	}
}

func TestMemoryStore_Get_Expired(t *testing.T) {
	store := newTestStore()
	ctx := context.Background()

	// Set token with already-expired duration
	store.mu.Lock()
	store.tokens["session-1"] = &storedToken{
		hash:      "hash-abc",
		expiresAt: time.Now().Add(-1 * time.Hour),
	}
	store.mu.Unlock()

	_, _, err := store.Get(ctx, "session-1")
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
	if !errors.Is(err, ErrExpiredToken) {
		t.Errorf("expected ErrExpiredToken, got %v", err)
	}
}

func TestMemoryStore_Delete(t *testing.T) {
	store := newTestStore()
	ctx := context.Background()

	err := store.Set(ctx, "session-1", "hash-abc", 1*time.Hour)
	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	err = store.Delete(ctx, "session-1")
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	if store.Count() != 0 {
		t.Errorf("expected count 0 after delete, got %d", store.Count())
	}

	_, _, err = store.Get(ctx, "session-1")
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestMemoryStore_Delete_NonExistent(t *testing.T) {
	store := newTestStore()
	ctx := context.Background()

	// Deleting non-existent should not error
	err := store.Delete(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("Delete of nonexistent returned error: %v", err)
	}
}

func TestMemoryStore_Rotate(t *testing.T) {
	store := newTestStore()
	ctx := context.Background()

	err := store.Set(ctx, "session-1", "old-hash", 1*time.Hour)
	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	err = store.Rotate(ctx, "session-1", "new-hash", 2*time.Hour)
	if err != nil {
		t.Fatalf("Rotate returned error: %v", err)
	}

	hash, _, err := store.Get(ctx, "session-1")
	if err != nil {
		t.Fatalf("Get after rotate returned error: %v", err)
	}
	if hash != "new-hash" {
		t.Errorf("expected hash 'new-hash' after rotate, got %q", hash)
	}

	if store.Count() != 1 {
		t.Errorf("expected count 1 after rotate, got %d", store.Count())
	}
}

func TestMemoryStore_Set_Overwrite(t *testing.T) {
	store := newTestStore()
	ctx := context.Background()

	err := store.Set(ctx, "session-1", "hash-1", 1*time.Hour)
	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	err = store.Set(ctx, "session-1", "hash-2", 1*time.Hour)
	if err != nil {
		t.Fatalf("Set overwrite returned error: %v", err)
	}

	hash, _, err := store.Get(ctx, "session-1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if hash != "hash-2" {
		t.Errorf("expected hash 'hash-2' after overwrite, got %q", hash)
	}

	if store.Count() != 1 {
		t.Errorf("expected count 1 after overwrite, got %d", store.Count())
	}
}

func TestMemoryStore_MultipleSessions(t *testing.T) {
	store := newTestStore()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		err := store.Set(ctx, "session-"+string(rune('a'+i)), "hash-"+string(rune('a'+i)), 1*time.Hour)
		if err != nil {
			t.Fatalf("Set returned error: %v", err)
		}
	}

	if store.Count() != 5 {
		t.Errorf("expected count 5, got %d", store.Count())
	}
}

func TestMemoryStore_Count(t *testing.T) {
	store := newTestStore()
	ctx := context.Background()

	if store.Count() != 0 {
		t.Errorf("expected count 0 for empty store, got %d", store.Count())
	}

	_ = store.Set(ctx, "s1", "h1", 1*time.Hour)
	_ = store.Set(ctx, "s2", "h2", 1*time.Hour)

	if store.Count() != 2 {
		t.Errorf("expected count 2, got %d", store.Count())
	}

	_ = store.Delete(ctx, "s1")

	if store.Count() != 1 {
		t.Errorf("expected count 1 after delete, got %d", store.Count())
	}
}
