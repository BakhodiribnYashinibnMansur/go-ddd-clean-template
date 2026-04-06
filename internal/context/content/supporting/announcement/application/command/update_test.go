package command

import (
	"context"
	"testing"
	"time"

	shared "gct/internal/kernel/domain"

	announceentity "gct/internal/context/content/supporting/announcement/domain/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUpdateAnnouncementHandler_Handle(t *testing.T) {
	t.Parallel()

	a, _ := announceentity.NewAnnouncement(
		shared.Lang{Uz: "old", Ru: "old", En: "old"},
		shared.Lang{Uz: "old", Ru: "old", En: "old"},
		1, nil, nil,
	)

	repo := &mockAnnouncementRepo{
		findFn: func(_ context.Context, id announceentity.AnnouncementID) (*announceentity.Announcement, error) {
			if id == a.TypedID() {
				return a, nil
			}
			return nil, announceentity.ErrAnnouncementNotFound
		},
	}
	eb := &mockEventBus{}
	handler := NewUpdateAnnouncementHandler(repo, eb, &mockLogger{})

	newTitle := shared.Lang{Uz: "new_uz", Ru: "new_ru", En: "new_en"}
	newPriority := 5
	cmd := UpdateAnnouncementCommand{
		ID:       announceentity.AnnouncementID(a.ID()),
		Title:    &newTitle,
		Priority: &newPriority,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)
	if repo.updated == nil {
		t.Fatal("expected announcement to be updated")
	}
	if repo.updated.Title().Uz != "new_uz" {
		t.Errorf("expected title new_uz, got %s", repo.updated.Title().Uz)
	}
	if repo.updated.Priority() != 5 {
		t.Errorf("expected priority 5, got %d", repo.updated.Priority())
	}
}

func TestUpdateAnnouncementHandler_WithPublish(t *testing.T) {
	t.Parallel()

	a, _ := announceentity.NewAnnouncement(
		shared.Lang{Uz: "t", Ru: "t", En: "t"},
		shared.Lang{Uz: "c", Ru: "c", En: "c"},
		1, nil, nil,
	)

	repo := &mockAnnouncementRepo{
		findFn: func(_ context.Context, id announceentity.AnnouncementID) (*announceentity.Announcement, error) {
			return a, nil
		},
	}
	eb := &mockEventBus{}
	handler := NewUpdateAnnouncementHandler(repo, eb, &mockLogger{})

	cmd := UpdateAnnouncementCommand{
		ID:      announceentity.AnnouncementID(a.ID()),
		Publish: true,
	}
	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)
	if !repo.updated.Published() {
		t.Error("expected announcement to be published")
	}
	if repo.updated.PublishedAt() == nil {
		t.Error("expected publishedAt to be set")
	}

	// Should have published events
	found := false
	for _, e := range eb.published {
		if e.EventName() == "announcement.published" {
			found = true
		}
	}
	if !found {
		t.Error("expected announcement.published event")
	}
}

func TestUpdateAnnouncementHandler_AlreadyPublished(t *testing.T) {
	t.Parallel()

	now := time.Now()
	a := announceentity.ReconstructAnnouncement(
		uuid.New(), time.Now(), time.Now(),
		shared.Lang{Uz: "t", Ru: "t", En: "t"},
		shared.Lang{Uz: "c", Ru: "c", En: "c"},
		true, &now, 1, nil, nil,
	)

	repo := &mockAnnouncementRepo{
		findFn: func(_ context.Context, _ announceentity.AnnouncementID) (*announceentity.Announcement, error) {
			return a, nil
		},
	}
	eb := &mockEventBus{}
	handler := NewUpdateAnnouncementHandler(repo, eb, &mockLogger{})

	cmd := UpdateAnnouncementCommand{
		ID:      announceentity.AnnouncementID(a.ID()),
		Publish: true,
	}
	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)
	// Should not re-publish since already published
	for _, e := range eb.published {
		if e.EventName() == "announcement.published" {
			t.Error("should not publish event for already published announcement")
		}
	}
}

func TestUpdateAnnouncementHandler_NotFound(t *testing.T) {
	t.Parallel()

	repo := &mockAnnouncementRepo{}
	handler := NewUpdateAnnouncementHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), UpdateAnnouncementCommand{ID: announceentity.NewAnnouncementID()})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}
