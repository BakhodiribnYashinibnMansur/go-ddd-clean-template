package domain

import (
	"time"

	"github.com/google/uuid"
)

// BaseEntity provides common fields for all domain entities.
type BaseEntity struct {
	id        uuid.UUID
	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
}

// NewBaseEntity creates a new BaseEntity with a generated UUID and current timestamps.
func NewBaseEntity() BaseEntity {
	now := time.Now()
	return BaseEntity{
		id:        uuid.New(),
		createdAt: now,
		updatedAt: now,
		deletedAt: nil,
	}
}

// NewBaseEntityWithID reconstructs a BaseEntity from persisted data.
func NewBaseEntityWithID(id uuid.UUID, createdAt, updatedAt time.Time, deletedAt *time.Time) BaseEntity {
	return BaseEntity{
		id:        id,
		createdAt: createdAt,
		updatedAt: updatedAt,
		deletedAt: deletedAt,
	}
}

// ID returns the entity's unique identifier.
func (e *BaseEntity) ID() uuid.UUID { return e.id }

// CreatedAt returns the entity's creation timestamp.
func (e *BaseEntity) CreatedAt() time.Time { return e.createdAt }

// UpdatedAt returns the entity's last update timestamp.
func (e *BaseEntity) UpdatedAt() time.Time { return e.updatedAt }

// DeletedAt returns the entity's soft-delete timestamp, or nil if not deleted.
func (e *BaseEntity) DeletedAt() *time.Time { return e.deletedAt }

// IsDeleted returns true if the entity has been soft-deleted.
func (e *BaseEntity) IsDeleted() bool { return e.deletedAt != nil }

// Touch updates the updatedAt timestamp to the current time.
func (e *BaseEntity) Touch() { e.updatedAt = time.Now() }

// SoftDelete marks the entity as deleted by setting deletedAt to the current time.
func (e *BaseEntity) SoftDelete() {
	now := time.Now()
	e.deletedAt = &now
}

// Restore clears the soft-delete timestamp.
func (e *BaseEntity) Restore() { e.deletedAt = nil }
