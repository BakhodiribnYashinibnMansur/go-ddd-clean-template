package revocation_test

import (
	"context"
	"testing"
	"time"

	"gct/internal/kernel/infrastructure/security/revocation"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setup(t *testing.T) (*revocation.Store, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { client.Close() })
	return revocation.New(client), mr
}

func TestRevokeAndIsRevoked(t *testing.T) {
	store, _ := setup(t)
	ctx := context.Background()

	err := store.Revoke(ctx, "sess-1", 5*time.Minute)
	if err != nil {
		t.Fatalf("Revoke returned error: %v", err)
	}

	if !store.IsRevoked(ctx, "sess-1") {
		t.Fatal("expected session to be revoked")
	}
}

func TestNotRevokedReturnsFalse(t *testing.T) {
	store, _ := setup(t)
	ctx := context.Background()

	if store.IsRevoked(ctx, "unknown-session") {
		t.Fatal("expected non-revoked session to return false")
	}
}

func TestRevokedEntryExpiresAfterTTL(t *testing.T) {
	store, mr := setup(t)
	ctx := context.Background()

	err := store.Revoke(ctx, "sess-ttl", 10*time.Second)
	if err != nil {
		t.Fatalf("Revoke returned error: %v", err)
	}

	if !store.IsRevoked(ctx, "sess-ttl") {
		t.Fatal("expected session to be revoked before TTL expires")
	}

	mr.FastForward(11 * time.Second)

	if store.IsRevoked(ctx, "sess-ttl") {
		t.Fatal("expected session to no longer be revoked after TTL")
	}
}

func TestRevokeMany(t *testing.T) {
	store, _ := setup(t)
	ctx := context.Background()

	ids := []string{"s1", "s2", "s3", "s4", "s5"}
	err := store.RevokeMany(ctx, ids, 5*time.Minute)
	if err != nil {
		t.Fatalf("RevokeMany returned error: %v", err)
	}

	for _, id := range ids {
		if !store.IsRevoked(ctx, id) {
			t.Fatalf("expected session %q to be revoked", id)
		}
	}
}

func TestRevokeManyEmpty(t *testing.T) {
	store, _ := setup(t)
	ctx := context.Background()

	err := store.RevokeMany(ctx, nil, 5*time.Minute)
	if err != nil {
		t.Fatalf("RevokeMany with empty slice returned error: %v", err)
	}
}

func TestNoopStoreAlwaysReturnsFalse(t *testing.T) {
	var noop revocation.NoopStore
	ctx := context.Background()

	if err := noop.Revoke(ctx, "x", time.Minute); err != nil {
		t.Fatalf("NoopStore.Revoke returned error: %v", err)
	}

	if noop.IsRevoked(ctx, "x") {
		t.Fatal("NoopStore.IsRevoked should always return false")
	}

	if err := noop.RevokeMany(ctx, []string{"a", "b"}, time.Minute); err != nil {
		t.Fatalf("NoopStore.RevokeMany returned error: %v", err)
	}
}
