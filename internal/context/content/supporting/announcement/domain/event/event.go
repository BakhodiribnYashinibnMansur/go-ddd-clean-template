package event

import (
	"time"

	"github.com/google/uuid"
)

// AnnouncementPublished is a domain event raised when an announcement transitions from draft to published.
// Downstream consumers may use this to trigger push notifications, cache invalidation, or read-model projections.
type AnnouncementPublished struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
}

func NewAnnouncementPublished(id uuid.UUID) AnnouncementPublished {
	return AnnouncementPublished{
		aggregateID: id,
		occurredAt:  time.Now(),
	}
}

func (e AnnouncementPublished) EventName() string      { return "announcement.published" }
func (e AnnouncementPublished) OccurredAt() time.Time  { return e.occurredAt }
func (e AnnouncementPublished) AggregateID() uuid.UUID { return e.aggregateID }
