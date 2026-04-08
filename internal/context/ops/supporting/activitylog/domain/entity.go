package domain

import (
	"time"

	"github.com/google/uuid"
)

// ActivityLogEntry represents a single field-level change record.
// It is intentionally immutable after creation — there are no Update or Delete methods.
// Each changed field in an update produces a separate ActivityLogEntry.
type ActivityLogEntry struct {
	id         int64
	actorID    uuid.UUID
	action     string
	entityType string
	entityID   uuid.UUID
	fieldName  *string
	oldValue   *string
	newValue   *string
	metadata   *string
	createdAt  time.Time
}

// NewActivityLogEntry creates a new ActivityLogEntry.
func NewActivityLogEntry(
	actorID uuid.UUID,
	action string,
	entityType string,
	entityID uuid.UUID,
	fieldName *string,
	oldValue *string,
	newValue *string,
	metadata *string,
) *ActivityLogEntry {
	return &ActivityLogEntry{
		actorID:    actorID,
		action:     action,
		entityType: entityType,
		entityID:   entityID,
		fieldName:  fieldName,
		oldValue:   oldValue,
		newValue:   newValue,
		metadata:   metadata,
		createdAt:  time.Now().UTC(),
	}
}

// ReconstructActivityLogEntry rebuilds an entry from persisted data.
func ReconstructActivityLogEntry(
	id int64,
	actorID uuid.UUID,
	action string,
	entityType string,
	entityID uuid.UUID,
	fieldName *string,
	oldValue *string,
	newValue *string,
	metadata *string,
	createdAt time.Time,
) *ActivityLogEntry {
	return &ActivityLogEntry{
		id:         id,
		actorID:    actorID,
		action:     action,
		entityType: entityType,
		entityID:   entityID,
		fieldName:  fieldName,
		oldValue:   oldValue,
		newValue:   newValue,
		metadata:   metadata,
		createdAt:  createdAt,
	}
}

// Getters

func (e *ActivityLogEntry) ID() int64            { return e.id }
func (e *ActivityLogEntry) ActorID() uuid.UUID   { return e.actorID }
func (e *ActivityLogEntry) Action() string        { return e.action }
func (e *ActivityLogEntry) EntityType() string    { return e.entityType }
func (e *ActivityLogEntry) EntityID() uuid.UUID   { return e.entityID }
func (e *ActivityLogEntry) FieldName() *string    { return e.fieldName }
func (e *ActivityLogEntry) OldValue() *string     { return e.oldValue }
func (e *ActivityLogEntry) NewValue() *string     { return e.newValue }
func (e *ActivityLogEntry) Metadata() *string     { return e.metadata }
func (e *ActivityLogEntry) CreatedAt() time.Time  { return e.createdAt }
