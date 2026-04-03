package csrf

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestRedisStore_Set_RedisDown verifies that Set returns an error when Redis is
// unreachable.
func TestRedisStore_Set_RedisDown(t *testing.T) {
	store, mr := setupRedisStore(t)
	ctx := context.Background()

	mr.Close()

	err := store.Set(ctx, "sess1", "hash123", 10*time.Minute)
	if err == nil {
		t.Fatal("Set() expected error when Redis is down, got nil")
	}
}

// TestRedisStore_Get_RedisDown verifies that Get returns a connection error
// (not ErrCSRFTokenNotFound) when Redis is unreachable.
func TestRedisStore_Get_RedisDown(t *testing.T) {
	store, mr := setupRedisStore(t)
	ctx := context.Background()

	mr.Close()

	_, _, err := store.Get(ctx, "sess1")
	if err == nil {
		t.Fatal("Get() expected error when Redis is down, got nil")
	}
	if errors.Is(err, ErrCSRFTokenNotFound) {
		t.Error("Get() should return a connection error, not ErrCSRFTokenNotFound")
	}
}

// TestRedisStore_Rotate_RedisDown verifies that Rotate returns an error when
// Redis is unreachable.
func TestRedisStore_Rotate_RedisDown(t *testing.T) {
	store, mr := setupRedisStore(t)
	ctx := context.Background()

	mr.Close()

	err := store.Rotate(ctx, "sess1", "newHash", 10*time.Minute)
	if err == nil {
		t.Fatal("Rotate() expected error when Redis is down, got nil")
	}
}

// TestRedisStore_RecoveryAfterRestart verifies that after Redis goes down and
// restarts on the same address, operations succeed again.
func TestRedisStore_RecoveryAfterRestart(t *testing.T) {
	store, mr := setupRedisStore(t)
	ctx := context.Background()

	// Phase 1: Redis is down — Set should fail.
	mr.Close()

	err := store.Set(ctx, "sess1", "hash1", 10*time.Minute)
	if err == nil {
		t.Fatal("Set() expected error during outage, got nil")
	}

	// Phase 2: Restart miniredis on the same address.
	if err := mr.Restart(); err != nil {
		t.Fatalf("failed to restart miniredis: %v", err)
	}

	// Set should now succeed.
	if err := store.Set(ctx, "sess1", "hash2", 10*time.Minute); err != nil {
		t.Fatalf("Set() after recovery returned error: %v", err)
	}

	val, err := mr.Get("csrf:sess1")
	if err != nil {
		t.Fatalf("miniredis Get failed: %v", err)
	}
	if val != "hash2" {
		t.Errorf("stored value = %q, want %q", val, "hash2")
	}
}
