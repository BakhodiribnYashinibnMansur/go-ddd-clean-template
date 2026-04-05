package domain

import (
	"fmt"
	"strings"
	"time"

	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
)

// Announcement is the aggregate root for announcements.
// It enforces a one-way publish lifecycle: once published, an announcement cannot revert to draft.
// The priority field determines display ordering; startDate/endDate define the visibility window.
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
// Returns an error if title has no non-empty translation in any supported language.
func NewAnnouncement(
	title shared.Lang,
	content shared.Lang,
	priority int,
	startDate *time.Time,
	endDate *time.Time,
) (*Announcement, error) {
	if strings.TrimSpace(title.Uz) == "" && strings.TrimSpace(title.Ru) == "" && strings.TrimSpace(title.En) == "" {
		return nil, fmt.Errorf("new_announcement: %s", "title is required")
	}
	return &Announcement{
		AggregateRoot: shared.NewAggregateRoot(),
		title:         title,
		content:       content,
		published:     false,
		publishedAt:   nil,
		priority:      priority,
		startDate:     startDate,
		endDate:       endDate,
	}, nil
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

// Publish performs a one-way state transition from draft to published.
// Callers should guard against double-publish by checking Published() first — this method
// does not enforce idempotency internally, so calling it twice will overwrite publishedAt.
func (a *Announcement) Publish() {
	now := time.Now()
	a.published = true
	a.publishedAt = &now
	a.Touch()
	a.AddEvent(NewAnnouncementPublished(a.ID()))
}

// Update applies partial modifications to the announcement.
// Nil pointer arguments are treated as "no change" — only non-nil fields are overwritten.
// This does not check published state, so callers must decide whether editing a published announcement is allowed.
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

func (a *Announcement) TypedID() AnnouncementID { return AnnouncementID(a.ID()) }
func (a *Announcement) Title() shared.Lang      { return a.title }
func (a *Announcement) Content() shared.Lang    { return a.content }
func (a *Announcement) Published() bool         { return a.published }
func (a *Announcement) PublishedAt() *time.Time { return a.publishedAt }
func (a *Announcement) Priority() int           { return a.priority }
func (a *Announcement) StartDate() *time.Time   { return a.startDate }
func (a *Announcement) EndDate() *time.Time     { return a.endDate }
