package domain

import (
	"context"
	"time"

	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
)

// ActivityLogFilter carries optional criteria for querying activity logs.
// All pointer fields are treated as "no filter" when nil.
type ActivityLogFilter struct {
	ActorID    *uuid.UUID
	EntityType *string
	EntityID   *uuid.UUID
	FieldName  *string
	Action     *string
	FromDate   *time.Time
	ToDate     *time.Time
	Pagination *shared.Pagination
}

// ActivityLogView is a flat read-model DTO for activity log queries.
type ActivityLogView struct {
	ID         int64      `json:"id"`
	ActorID    uuid.UUID  `json:"actor_id"`
	Action     string     `json:"action"`
	EntityType string     `json:"entity_type"`
	EntityID   uuid.UUID  `json:"entity_id"`
	FieldName  *string    `json:"field_name,omitempty"`
	OldValue   *string    `json:"old_value,omitempty"`
	NewValue   *string    `json:"new_value,omitempty"`
	Metadata   *string    `json:"metadata,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// ActivityLogWriteRepository is the write-side repository for activity logs.
// It is append-only — entries are never modified or deleted.
type ActivityLogWriteRepository interface {
	SaveBatch(ctx context.Context, entries []*ActivityLogEntry) error
}

// ActivityLogReadRepository is the read-side (CQRS query) repository.
type ActivityLogReadRepository interface {
	List(ctx context.Context, filter ActivityLogFilter) ([]*ActivityLogView, int64, error)
}
