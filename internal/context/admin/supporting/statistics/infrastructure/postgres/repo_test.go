package postgres

import (
	"testing"
)

func TestNewStatisticsReadRepo(t *testing.T) {
	repo := NewStatisticsReadRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

func TestNewStatisticsReadRepo_PoolIsNil(t *testing.T) {
	repo := NewStatisticsReadRepo(nil)
	if repo.pool != nil {
		t.Fatal("expected nil pool")
	}
}
