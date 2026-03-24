package postgres

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestNewUserWriteRepo_NotNil(t *testing.T) {
	// We pass a nil pool because we are only testing construction, not DB access.
	repo := NewUserWriteRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil UserWriteRepo")
	}
}

func TestNewUserWriteRepo_PoolReference(t *testing.T) {
	// Use an unsafe cast to verify the pool field is stored correctly.
	// Since we cannot create a real pool without a database, we just verify nil is stored.
	var pool *pgxpool.Pool
	repo := NewUserWriteRepo(pool)
	if repo.pool != pool {
		t.Fatal("expected repo.pool to match the provided pool")
	}
}

func TestNewUserReadRepo_NotNil(t *testing.T) {
	repo := NewUserReadRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil UserReadRepo")
	}
}

func TestNewUserReadRepo_PoolReference(t *testing.T) {
	var pool *pgxpool.Pool
	repo := NewUserReadRepo(pool)
	if repo.pool != pool {
		t.Fatal("expected repo.pool to match the provided pool")
	}
}
