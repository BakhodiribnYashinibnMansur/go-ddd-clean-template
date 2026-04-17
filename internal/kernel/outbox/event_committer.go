package outbox

import (
	"context"
	"fmt"

	"gct/internal/kernel/application"
	shareddomain "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// EventCommitter wraps the transactional outbox pattern into a single call.
//
// Production (writer != nil): begins a transaction, runs fn with the tx as
// Querier, serializes domain events into outbox entries, appends them within
// the same transaction, and commits.
//
// Dev mode (writer == nil): runs fn with the pool as Querier, then publishes
// events directly via the event bus.
type EventCommitter struct {
	pool     shareddomain.DB
	writer   Writer
	eventBus application.EventBus
	logger   logger.Log
}

// NewEventCommitter creates an EventCommitter. Pass a nil writer to disable
// the outbox and fall back to direct event bus publishing.
func NewEventCommitter(
	pool shareddomain.DB,
	writer Writer,
	eventBus application.EventBus,
	l logger.Log,
) *EventCommitter {
	return &EventCommitter{
		pool:     pool,
		writer:   writer,
		eventBus: eventBus,
		logger:   l,
	}
}

// Commit executes fn and ensures domain events are delivered reliably.
//
// fn receives a Querier that is either a pgx.Tx (outbox mode) or the pool
// (direct mode). Repositories should use this Querier for all writes so they
// participate in the same transaction when one exists.
func (c *EventCommitter) Commit(
	ctx context.Context,
	fn func(ctx context.Context, q shareddomain.Querier) error,
	events func() []shareddomain.DomainEvent,
) error {
	if c.writer == nil {
		return c.commitDirect(ctx, fn, events)
	}
	return c.commitOutbox(ctx, fn, events)
}

// commitOutbox wraps fn + outbox append in a single database transaction.
func (c *EventCommitter) commitOutbox(
	ctx context.Context,
	fn func(ctx context.Context, q shareddomain.Querier) error,
	events func() []shareddomain.DomainEvent,
) error {
	return pgxutil.WithTx(ctx, c.pool, func(tx shareddomain.Querier) error {
		if err := fn(ctx, tx); err != nil {
			return err
		}

		domainEvents := events()
		if len(domainEvents) == 0 {
			return nil
		}

		entries, err := ToEntries(domainEvents)
		if err != nil {
			return fmt.Errorf("outbox serialize: %w", err)
		}

		return c.writer.Append(ctx, tx, entries...)
	})
}

// commitDirect preserves the pre-outbox behavior: run fn, then publish.
func (c *EventCommitter) commitDirect(
	ctx context.Context,
	fn func(ctx context.Context, q shareddomain.Querier) error,
	events func() []shareddomain.DomainEvent,
) error {
	if err := fn(ctx, c.pool); err != nil {
		return err
	}

	domainEvents := events()
	if len(domainEvents) == 0 {
		return nil
	}

	if err := c.eventBus.Publish(ctx, domainEvents...); err != nil {
		c.logger.Warnc(ctx, "event publish failed (direct mode)", "error", err)
	}

	return nil
}
