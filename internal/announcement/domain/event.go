package domain

import (
	"time"

	"github.com/google/uuid"
)

// AnnouncementPublished is raised when an announcement is published.
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
func (e AnnouncementPublished) OccurredAt() time.Time   { return e.occurredAt }
func (e AnnouncementPublished) AggregateID() uuid.UUID  { return e.aggregateID }
