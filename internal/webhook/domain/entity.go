package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// Webhook is the aggregate root for webhook management.
type Webhook struct {
	shared.AggregateRoot
	name    string
	url     string
	secret  string
	events  []string
	enabled bool
}

// NewWebhook creates a new Webhook aggregate.
func NewWebhook(name, url, secret string, events []string, enabled bool) *Webhook {
	if events == nil {
		events = make([]string, 0)
	}
	w := &Webhook{
		AggregateRoot: shared.NewAggregateRoot(),
		name:          name,
		url:           url,
		secret:        secret,
		events:        events,
		enabled:       enabled,
	}
	return w
}

// ReconstructWebhook rebuilds a Webhook aggregate from persisted data. No events are raised.
func ReconstructWebhook(
	id uuid.UUID,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
	name, url, secret string,
	events []string,
	enabled bool,
) *Webhook {
	if events == nil {
		events = make([]string, 0)
	}
	return &Webhook{
		AggregateRoot: shared.NewAggregateRootWithID(id, createdAt, updatedAt, deletedAt),
		name:          name,
		url:           url,
		secret:        secret,
		events:        events,
		enabled:       enabled,
	}
}

// Trigger adds a WebhookTriggered event.
func (w *Webhook) Trigger() {
	w.AddEvent(NewWebhookTriggered(w.ID(), w.url))
}

// UpdateDetails updates mutable fields.
func (w *Webhook) UpdateDetails(name, url, secret *string, events []string, enabled *bool) {
	if name != nil {
		w.name = *name
	}
	if url != nil {
		w.url = *url
	}
	if secret != nil {
		w.secret = *secret
	}
	if events != nil {
		w.events = events
	}
	if enabled != nil {
		w.enabled = *enabled
	}
	w.Touch()
}

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func (w *Webhook) Name() string      { return w.name }
func (w *Webhook) URL() string       { return w.url }
func (w *Webhook) Secret() string    { return w.secret }
func (w *Webhook) Events_() []string { return w.events }
func (w *Webhook) Enabled() bool     { return w.enabled }
