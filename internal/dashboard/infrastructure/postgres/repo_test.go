package postgres

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Constructor tests
// ---------------------------------------------------------------------------

func TestNewDashboardReadRepo(t *testing.T) {
	repo := NewDashboardReadRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

func TestNewDashboardReadRepo_PoolIsNil(t *testing.T) {
	repo := NewDashboardReadRepo(nil)
	if repo.pool != nil {
		t.Fatal("expected nil pool")
	}
}
