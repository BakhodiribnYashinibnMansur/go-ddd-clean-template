package outbox

import (
	"context"
	"fmt"

	"gct/internal/kernel/application"
	shareddomain "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/jackc/pgx/v5"
)

// EventCommitter wraps the transactional outbox pattern into a single call.
//
// Production (writer != nil): begins a transaction, runs fn, serializes domain
// events into outbox entries, appends them within the same transaction, and
// commits. The relay goroutine picks up the rows and publishes them later.
//
// Dev mode (writer == nil): runs fn, then publishes events directly via the
// event bus — preserving the pre-outbox behavior for in-memory setups.
type EventCommitter struct {
	pool     pgxutil.TxBeginner
	writer   Writer // nil disables outbox (dev mode)
	eventBus application.EventBus
	logger   logger.Log
}

// NewEventCommitter creates an EventCommitter. Pass a nil writer to disable
// the outbox and fall back to direct event bus publishing.
func NewEventCommitter(
	pool pgxutil.TxBeginner,
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
// fn receives a context that may carry a transaction (via pgxutil.InjectTx).
// Repositories that call pgxutil.WithTx or pgxutil.QuerierFromContext will
// automatically participate in that transaction.
//
// events is a function (typically aggregate.Events) that returns the domain
// events collected during fn. It is called after fn succeeds.
func (c *EventCommitter) Commit(
	ctx context.Context,
	fn func(ctx context.Context) error,
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
	fn func(ctx context.Context) error,
	events func() []shareddomain.DomainEvent,
) error {
	return pgxutil.WithTx(ctx, c.pool, func(tx pgx.Tx) error {
		txCtx := pgxutil.InjectTx(ctx, tx)

		if err := fn(txCtx); err != nil {
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

		return c.writer.Append(txCtx, tx, entries...)
	})
}

// commitDirect preserves the pre-outbox behavior: run fn, then publish.
func (c *EventCommitter) commitDirect(
	ctx context.Context,
	fn func(ctx context.Context) error,
	events func() []shareddomain.DomainEvent,
) error {
	if err := fn(ctx); err != nil {
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
