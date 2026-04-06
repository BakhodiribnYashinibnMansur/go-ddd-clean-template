package query

import (
	"context"
	"errors"
	"gct/internal/kernel/infrastructure/logger"
	"testing"
	"time"

	announceentity "gct/internal/context/content/supporting/announcement/domain/entity"
	announcerepo "gct/internal/context/content/supporting/announcement/domain/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockReadRepo struct {
	view  *announcerepo.AnnouncementView
	views []*announcerepo.AnnouncementView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id announceentity.AnnouncementID) (*announcerepo.AnnouncementView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, announceentity.ErrAnnouncementNotFound
}

func (m *mockReadRepo) List(_ context.Context, _ announcerepo.AnnouncementFilter) ([]*announcerepo.AnnouncementView, int64, error) {
	return m.views, m.total, nil
}

type errorReadRepo struct{ err error }

func (m *errorReadRepo) FindByID(_ context.Context, _ announceentity.AnnouncementID) (*announcerepo.AnnouncementView, error) {
	return nil, m.err
}

func (m *errorReadRepo) List(_ context.Context, _ announcerepo.AnnouncementFilter) ([]*announcerepo.AnnouncementView, int64, error) {
	return nil, 0, m.err
}

var errRepo = errors.New("repo failure")

// --- Tests ---

func TestGetAnnouncementHandler_Handle(t *testing.T) {
	t.Parallel()

	id := announceentity.NewAnnouncementID()
	now := time.Now()
	readRepo := &mockReadRepo{
		view: &announcerepo.AnnouncementView{
			ID:        id,
			TitleUz:   "title_uz",
			TitleRu:   "title_ru",
			TitleEn:   "title_en",
			ContentUz: "content_uz",
			ContentRu: "content_ru",
			ContentEn: "content_en",
			Published: false,
			Priority:  3,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	handler := NewGetAnnouncementHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetAnnouncementQuery{ID: announceentity.AnnouncementID(id)})
	require.NoError(t, err)
	if result == nil {
		t.Fatal("expected result")
	}
	if result.Title.Uz != "title_uz" {
		t.Errorf("expected title_uz, got %s", result.Title.Uz)
	}
	if result.Content.En != "content_en" {
		t.Errorf("expected content_en, got %s", result.Content.En)
	}
	if result.Priority != 3 {
		t.Errorf("expected priority 3, got %d", result.Priority)
	}
	if result.Published {
		t.Error("expected unpublished")
	}
}

func TestGetAnnouncementHandler_NotFound(t *testing.T) {
	t.Parallel()

	readRepo := &mockReadRepo{}
	handler := NewGetAnnouncementHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetAnnouncementQuery{ID: announceentity.AnnouncementID(uuid.New())})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetAnnouncementHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &errorReadRepo{err: errRepo}
	handler := NewGetAnnouncementHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetAnnouncementQuery{ID: announceentity.AnnouncementID(uuid.New())})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestGetAnnouncementHandler_AllFieldsMapped(t *testing.T) {
	t.Parallel()

	id := announceentity.NewAnnouncementID()
	now := time.Now()
	readRepo := &mockReadRepo{
		view: &announcerepo.AnnouncementView{
			ID:          id,
			TitleUz:     "uz",
			TitleRu:     "ru",
			TitleEn:     "en",
			ContentUz:   "cuz",
			ContentRu:   "cru",
			ContentEn:   "cen",
			Published:   true,
			PublishedAt: &now,
			Priority:    10,
			StartDate:   &now,
			EndDate:     &now,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	handler := NewGetAnnouncementHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetAnnouncementQuery{ID: announceentity.AnnouncementID(id)})
	require.NoError(t, err)
	if !result.Published {
		t.Error("expected published")
	}
	if result.PublishedAt == nil {
		t.Error("expected publishedAt to be set")
	}
	if result.StartDate == nil {
		t.Error("expected startDate to be set")
	}
	if result.EndDate == nil {
		t.Error("expected endDate to be set")
	}
	if result.Title.Ru != "ru" {
		t.Errorf("expected ru, got %s", result.Title.Ru)
	}
}
