package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"gct/internal/platform/domain"
)

func TestNewBaseEntity_FieldsNotZero(t *testing.T) {
	e := domain.NewBaseEntity()

	if e.ID() == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
	if e.CreatedAt().IsZero() {
		t.Error("expected non-zero createdAt")
	}
	if e.UpdatedAt().IsZero() {
		t.Error("expected non-zero updatedAt")
	}
	if e.DeletedAt() != nil {
		t.Error("expected nil deletedAt")
	}
	if e.IsDeleted() {
		t.Error("expected IsDeleted to be false")
	}
}

func TestNewBaseEntityWithID(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	deleted := now.Add(-time.Hour)

	e := domain.NewBaseEntityWithID(id, now, now, &deleted)

	if e.ID() != id {
		t.Errorf("expected ID %s, got %s", id, e.ID())
	}
	if !e.CreatedAt().Equal(now) {
		t.Error("createdAt mismatch")
	}
	if e.DeletedAt() == nil || !e.DeletedAt().Equal(deleted) {
		t.Error("deletedAt mismatch")
	}
	if !e.IsDeleted() {
		t.Error("expected IsDeleted to be true")
	}
}

func TestBaseEntity_SoftDeleteAndRestore(t *testing.T) {
	e := domain.NewBaseEntity()

	if e.IsDeleted() {
		t.Fatal("should not be deleted initially")
	}

	e.SoftDelete()
	if !e.IsDeleted() {
		t.Error("expected IsDeleted after SoftDelete")
	}
	if e.DeletedAt() == nil {
		t.Error("expected non-nil deletedAt after SoftDelete")
	}

	e.Restore()
	if e.IsDeleted() {
		t.Error("expected not deleted after Restore")
	}
	if e.DeletedAt() != nil {
		t.Error("expected nil deletedAt after Restore")
	}
}

func TestBaseEntity_Touch(t *testing.T) {
	e := domain.NewBaseEntity()
	original := e.UpdatedAt()

	// Small sleep to ensure time advances
	time.Sleep(time.Millisecond)
	e.Touch()

	if !e.UpdatedAt().After(original) {
		t.Error("expected Touch to advance updatedAt")
	}
}
