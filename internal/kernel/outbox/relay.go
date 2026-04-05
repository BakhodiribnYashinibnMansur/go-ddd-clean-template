package outbox

import (
	"context"
	"time"

	"gct/internal/kernel/infrastructure/logger"
)

// RawPublisher publishes an outbox entry to the event bus. Implementations
// re-hydrate the payload into a typed event (or publish the raw bytes) and
// forward it to subscribers. Kept as an interface so the relay is
// decoupled from whichever bus implementation the app uses.
type RawPublisher interface {
	PublishRaw(ctx context.Context, eventName string, payload []byte) error
}

// Relay polls the outbox for pending entries and forwards them to a
// RawPublisher. It is intentionally simple: single poller, no leader
// election — suitable for single-instance deployments. Scale-out would
// add SELECT ... FOR UPDATE SKIP LOCKED.
type Relay struct {
	store     Store
	publisher RawPublisher
	log       logger.Log
	interval  time.Duration
	batch     int
}

// NewRelay builds a relay. Reasonable defaults: interval 2s, batch 100.
func NewRelay(store Store, pub RawPublisher, log logger.Log, interval time.Duration, batch int) *Relay {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	if batch <= 0 {
		batch = 100
	}
	return &Relay{store: store, publisher: pub, log: log, interval: interval, batch: batch}
}

// Run blocks, polling at the configured interval until ctx is cancelled.
func (r *Relay) Run(ctx context.Context) {
	t := time.NewTicker(r.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			r.tick(ctx)
		}
	}
}

func (r *Relay) tick(ctx context.Context) {
	entries, err := r.store.Pending(ctx, r.batch)
	if err != nil {
		r.log.Errorc(ctx, "outbox relay: fetch pending failed", "error", err)
		return
	}
	if len(entries) == 0 {
		return
	}
	var dispatched []int64
	for _, e := range entries {
		if err := r.publisher.PublishRaw(ctx, e.EventName, e.Payload); err != nil {
			r.log.Warnc(ctx, "outbox relay: publish failed", "event_id", e.EventID, "event_name", e.EventName, "error", err)
			if mfErr := r.store.MarkFailed(ctx, e.ID, err.Error()); mfErr != nil {
				r.log.Errorc(ctx, "outbox relay: mark failed", "error", mfErr)
			}
			continue
		}
		dispatched = append(dispatched, e.ID)
	}
	if len(dispatched) > 0 {
		if err := r.store.MarkDispatched(ctx, dispatched); err != nil {
			r.log.Errorc(ctx, "outbox relay: mark dispatched failed", "count", len(dispatched), "error", err)
		}
	}
}
