package outbox

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgStore is a PostgreSQL-backed outbox repository. It implements both Writer
// (in-transaction appends) and Store (relay reads/updates).
type PgStore struct {
	pool *pgxpool.Pool
}

// NewPgStore wires the outbox repository to a pgx pool.
func NewPgStore(pool *pgxpool.Pool) *PgStore {
	return &PgStore{pool: pool}
}

// Append inserts one or more outbox rows within the caller's transaction.
// The producer MUST pass the same pgx.Tx that mutated the aggregate, so the
// outbox write commits or rolls back atomically with the business change.
func (s *PgStore) Append(ctx context.Context, tx pgx.Tx, entries ...Entry) error {
	if tx == nil {
		return errors.New("outbox: nil transaction")
	}
	if len(entries) == 0 {
		return nil
	}
	const stmt = `INSERT INTO outbox_events
		(event_id, event_name, aggregate_id, payload, occurred_at)
		VALUES ($1, $2, $3, $4, $5)`
	for _, e := range entries {
		if _, err := tx.Exec(ctx, stmt,
			e.EventID, e.EventName, e.AggregateID, e.Payload, e.OccurredAt); err != nil {
			return fmt.Errorf("outbox append: %w", err)
		}
	}
	return nil
}

// Pending returns up to limit undispatched rows ordered by id.
func (s *PgStore) Pending(ctx context.Context, limit int) ([]Entry, error) {
	const q = `SELECT id, event_id, event_name, aggregate_id, payload, occurred_at,
		    created_at, dispatched_at, attempts, last_error
		FROM outbox_events
		WHERE dispatched_at IS NULL
		ORDER BY id
		LIMIT $1`
	rows, err := s.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("outbox pending: %w", err)
	}
	defer rows.Close()

	var out []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.EventID, &e.EventName, &e.AggregateID, &e.Payload,
			&e.OccurredAt, &e.CreatedAt, &e.DispatchedAt, &e.Attempts, &e.LastError); err != nil {
			return nil, fmt.Errorf("outbox scan: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// MarkDispatched flags rows as successfully published.
func (s *PgStore) MarkDispatched(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	const q = `UPDATE outbox_events SET dispatched_at = now() WHERE id = ANY($1)`
	_, err := s.pool.Exec(ctx, q, ids)
	if err != nil {
		return fmt.Errorf("outbox mark dispatched: %w", err)
	}
	return nil
}

// MarkFailed increments attempts and records the error message.
func (s *PgStore) MarkFailed(ctx context.Context, id int64, errMsg string) error {
	const q = `UPDATE outbox_events
		SET attempts = attempts + 1, last_error = $2
		WHERE id = $1`
	_, err := s.pool.Exec(ctx, q, id, errMsg)
	if err != nil {
		return fmt.Errorf("outbox mark failed: %w", err)
	}
	return nil
}
