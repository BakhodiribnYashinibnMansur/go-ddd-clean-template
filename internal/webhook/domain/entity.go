package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// Webhook is the aggregate root for outbound webhook subscriptions.
// The events slice holds domain event names (e.g., "user.created") this webhook listens to.
// The secret field is used to sign payloads (HMAC-SHA256) so receivers can verify authenticity.
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

// Trigger raises a WebhookTriggered event that the application layer uses to enqueue HTTP delivery.
// Callers should verify the webhook is enabled before calling this.
func (w *Webhook) Trigger() {
	w.AddEvent(NewWebhookTriggered(w.ID(), w.url))
}

// UpdateDetails applies partial modifications using pointer semantics — nil fields are left unchanged.
// Note: events uses slice semantics (nil = no change, empty slice = clear all subscriptions).
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
