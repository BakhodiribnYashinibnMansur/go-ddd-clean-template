// Package outbox implements the transactional outbox pattern. A producer
// appends a serialized event record to the outbox inside the same database
// transaction as the aggregate mutation; a relay goroutine reads the table
// and forwards rows to the event bus, guaranteeing at-least-once delivery
// even if the process crashes between commit and publish.
//
// Consumers must dedupe by event_id (stored in contracts/events.Envelope).
package outbox

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Entry is one outbox row.
type Entry struct {
	ID            int64
	EventID       uuid.UUID
	EventName     string
	AggregateID   uuid.UUID
	Payload       []byte
	OccurredAt    time.Time
	CreatedAt     time.Time
	DispatchedAt  *time.Time
	Attempts      int
	LastError     *string
}

// Writer inserts rows into the outbox within an existing transaction. Call
// this from a command handler after mutating the aggregate but BEFORE the
// transaction commits, using the same pgx.Tx.
type Writer interface {
	Append(ctx context.Context, tx pgx.Tx, entries ...Entry) error
}

// Store is the read/update side used by the relay.
type Store interface {
	// Pending returns up to limit undispatched rows ordered by id.
	Pending(ctx context.Context, limit int) ([]Entry, error)
	// MarkDispatched flags rows as published.
	MarkDispatched(ctx context.Context, ids []int64) error
	// MarkFailed records a publish attempt error.
	MarkFailed(ctx context.Context, id int64, errMsg string) error
}
