package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// Announcement is the aggregate root for announcements.
type Announcement struct {
	shared.AggregateRoot
	title       shared.Lang
	content     shared.Lang
	published   bool
	publishedAt *time.Time
	priority    int
	startDate   *time.Time
	endDate     *time.Time
}

// NewAnnouncement creates a new Announcement aggregate.
func NewAnnouncement(
	title shared.Lang,
	content shared.Lang,
	priority int,
	startDate *time.Time,
	endDate *time.Time,
) *Announcement {
	return &Announcement{
		AggregateRoot: shared.NewAggregateRoot(),
		title:         title,
		content:       content,
		published:     false,
		publishedAt:   nil,
		priority:      priority,
		startDate:     startDate,
		endDate:       endDate,
	}
}

// ReconstructAnnouncement rebuilds an Announcement from persisted data. No events are raised.
func ReconstructAnnouncement(
	id uuid.UUID,
	createdAt, updatedAt time.Time,
	title shared.Lang,
	content shared.Lang,
	published bool,
	publishedAt *time.Time,
	priority int,
	startDate *time.Time,
	endDate *time.Time,
) *Announcement {
	return &Announcement{
		AggregateRoot: shared.NewAggregateRootWithID(id, createdAt, updatedAt, nil),
		title:         title,
		content:       content,
		published:     published,
		publishedAt:   publishedAt,
		priority:      priority,
		startDate:     startDate,
		endDate:       endDate,
	}
}

// Publish sets the announcement as published and raises an AnnouncementPublished event.
func (a *Announcement) Publish() {
	now := time.Now()
	a.published = true
	a.publishedAt = &now
	a.Touch()
	a.AddEvent(NewAnnouncementPublished(a.ID()))
}

// Update modifies the announcement fields and touches the updatedAt timestamp.
func (a *Announcement) Update(title *shared.Lang, content *shared.Lang, priority *int, startDate *time.Time, endDate *time.Time) {
	if title != nil {
		a.title = *title
	}
	if content != nil {
		a.content = *content
	}
	if priority != nil {
		a.priority = *priority
	}
	if startDate != nil {
		a.startDate = startDate
	}
	if endDate != nil {
		a.endDate = endDate
	}
	a.Touch()
}

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func (a *Announcement) Title() shared.Lang      { return a.title }
func (a *Announcement) Content() shared.Lang     { return a.content }
func (a *Announcement) Published() bool          { return a.published }
func (a *Announcement) PublishedAt() *time.Time  { return a.publishedAt }
func (a *Announcement) Priority() int            { return a.priority }
func (a *Announcement) StartDate() *time.Time    { return a.startDate }
func (a *Announcement) EndDate() *time.Time      { return a.endDate }
