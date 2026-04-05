package csrf

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setupRedisStore(t *testing.T) (*RedisStore, *miniredis.Miniredis) {
	t.Helper()
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	return NewRedisStore(client), s
}

func TestNewRedisStore(t *testing.T) {
	store, _ := setupRedisStore(t)

	if store.prefix != "csrf:" {
		t.Errorf("expected prefix %q, got %q", "csrf:", store.prefix)
	}
	if store.client == nil {
		t.Fatal("expected client to be non-nil")
	}
}

func TestRedisStore_Key(t *testing.T) {
	store, _ := setupRedisStore(t)

	got := store.key("session123")
	want := "csrf:session123"
	if got != want {
		t.Errorf("key() = %q, want %q", got, want)
	}
}

func TestRedisStore_Set(t *testing.T) {
	store, mr := setupRedisStore(t)
	ctx := context.Background()

	err := store.Set(ctx, "sess1", "hash123", 10*time.Minute)
	if err != nil {
		t.Fatalf("Set() returned error: %v", err)
	}

	val, err := mr.Get("csrf:sess1")
	if err != nil {
		t.Fatalf("miniredis Get failed: %v", err)
	}
	if val != "hash123" {
		t.Errorf("stored value = %q, want %q", val, "hash123")
	}
}

func TestRedisStore_Get_Success(t *testing.T) {
	store, _ := setupRedisStore(t)
	ctx := context.Background()

	err := store.Set(ctx, "sess1", "hash456", 10*time.Minute)
	if err != nil {
		t.Fatalf("Set() returned error: %v", err)
	}

	tokenHash, expiresAt, err := store.Get(ctx, "sess1")
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}
	if tokenHash != "hash456" {
		t.Errorf("tokenHash = %q, want %q", tokenHash, "hash456")
	}
	if !expiresAt.After(time.Now()) {
		t.Errorf("expiresAt should be in the future, got %v", expiresAt)
	}
}

func TestRedisStore_Get_NotFound(t *testing.T) {
	store, _ := setupRedisStore(t)
	ctx := context.Background()

	_, _, err := store.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Get() expected error, got nil")
	}
	if !errors.Is(err, ErrCSRFTokenNotFound) {
		t.Errorf("expected ErrCSRFTokenNotFound, got: %v", err)
	}
}

func TestRedisStore_Get_RedisError(t *testing.T) {
	store, mr := setupRedisStore(t)
	ctx := context.Background()

	mr.Close()

	_, _, err := store.Get(ctx, "sess1")
	if err == nil {
		t.Fatal("Get() expected error after closing miniredis, got nil")
	}
	if errors.Is(err, ErrCSRFTokenNotFound) {
		t.Error("expected a connection error, not ErrCSRFTokenNotFound")
	}
}

func TestRedisStore_Delete(t *testing.T) {
	store, mr := setupRedisStore(t)
	ctx := context.Background()

	err := store.Set(ctx, "sess1", "hash789", 10*time.Minute)
	if err != nil {
		t.Fatalf("Set() returned error: %v", err)
	}

	err = store.Delete(ctx, "sess1")
	if err != nil {
		t.Fatalf("Delete() returned error: %v", err)
	}

	if mr.Exists("csrf:sess1") {
		t.Error("key csrf:sess1 should not exist after Delete")
	}
}

func TestRedisStore_Rotate(t *testing.T) {
	store, mr := setupRedisStore(t)
	ctx := context.Background()

	err := store.Set(ctx, "sess1", "oldHash", 10*time.Minute)
	if err != nil {
		t.Fatalf("Set() returned error: %v", err)
	}

	err = store.Rotate(ctx, "sess1", "newHash", 15*time.Minute)
	if err != nil {
		t.Fatalf("Rotate() returned error: %v", err)
	}

	val, err := mr.Get("csrf:sess1")
	if err != nil {
		t.Fatalf("miniredis Get failed: %v", err)
	}
	if val != "newHash" {
		t.Errorf("after Rotate, stored value = %q, want %q", val, "newHash")
	}
}

func TestRedisStore_Count(t *testing.T) {
	store, _ := setupRedisStore(t)
	ctx := context.Background()

	for i, id := range []string{"a", "b", "c"} {
		err := store.Set(ctx, id, "hash"+id, 10*time.Minute)
		if err != nil {
			t.Fatalf("Set() #%d returned error: %v", i, err)
		}
	}

	count, err := store.Count(ctx)
	if err != nil {
		t.Fatalf("Count() returned error: %v", err)
	}
	if count != 3 {
		t.Errorf("Count() = %d, want 3", count)
	}
}

func TestRedisStore_Count_Empty(t *testing.T) {
	store, _ := setupRedisStore(t)
	ctx := context.Background()

	count, err := store.Count(ctx)
	if err != nil {
		t.Fatalf("Count() returned error: %v", err)
	}
	if count != 0 {
		t.Errorf("Count() = %d, want 0", count)
	}
}

func TestRedisStore_Cleanup(t *testing.T) {
	store, _ := setupRedisStore(t)
	ctx := context.Background()

	err := store.Cleanup(ctx)
	if err != nil {
		t.Errorf("Cleanup() returned error: %v, want nil", err)
	}
}

func TestRedisStore_Set_WithExpiration(t *testing.T) {
	store, mr := setupRedisStore(t)
	ctx := context.Background()

	err := store.Set(ctx, "sess1", "hashTTL", 5*time.Minute)
	if err != nil {
		t.Fatalf("Set() returned error: %v", err)
	}

	ttl := mr.TTL("csrf:sess1")
	if ttl <= 0 {
		t.Fatalf("expected positive TTL, got %v", ttl)
	}
	if ttl > 5*time.Minute {
		t.Errorf("TTL = %v, expected <= 5m", ttl)
	}
}
