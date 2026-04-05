package command

import (
	"context"
	"testing"
	"time"

	"gct/internal/context/content/announcement/domain"
	"gct/internal/platform/application"
	shared "gct/internal/platform/domain"

	"github.com/google/uuid"
)

// --- Mocks ---

type mockAnnouncementRepo struct {
	saved   *domain.Announcement
	updated *domain.Announcement
	deleted uuid.UUID
	findFn  func(ctx context.Context, id uuid.UUID) (*domain.Announcement, error)
}

func (m *mockAnnouncementRepo) Save(_ context.Context, e *domain.Announcement) error {
	m.saved = e
	return nil
}

func (m *mockAnnouncementRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Announcement, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, domain.ErrAnnouncementNotFound
}

func (m *mockAnnouncementRepo) Update(_ context.Context, e *domain.Announcement) error {
	m.updated = e
	return nil
}

func (m *mockAnnouncementRepo) Delete(_ context.Context, id uuid.UUID) error {
	m.deleted = id
	return nil
}

func (m *mockAnnouncementRepo) List(_ context.Context, _ domain.AnnouncementFilter) ([]*domain.Announcement, int64, error) {
	return nil, 0, nil
}

type mockEventBus struct {
	published []shared.DomainEvent
}

func (m *mockEventBus) Publish(_ context.Context, events ...shared.DomainEvent) error {
	m.published = append(m.published, events...)
	return nil
}

func (m *mockEventBus) Subscribe(_ string, _ application.EventHandler) error { return nil }

type mockLogger struct{}

func (m *mockLogger) Debug(_ ...any)                                {}
func (m *mockLogger) Debugf(_ string, _ ...any)                     {}
func (m *mockLogger) Debugw(_ string, _ ...any)                     {}
func (m *mockLogger) Info(_ ...any)                                 {}
func (m *mockLogger) Infof(_ string, _ ...any)                      {}
func (m *mockLogger) Infow(_ string, _ ...any)                      {}
func (m *mockLogger) Warn(_ ...any)                                 {}
func (m *mockLogger) Warnf(_ string, _ ...any)                      {}
func (m *mockLogger) Warnw(_ string, _ ...any)                      {}
func (m *mockLogger) Error(_ ...any)                                {}
func (m *mockLogger) Errorf(_ string, _ ...any)                     {}
func (m *mockLogger) Errorw(_ string, _ ...any)                     {}
func (m *mockLogger) Fatal(_ ...any)                                {}
func (m *mockLogger) Fatalf(_ string, _ ...any)                     {}
func (m *mockLogger) Fatalw(_ string, _ ...any)                     {}
func (m *mockLogger) Debugc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLogger) Infoc(_ context.Context, _ string, _ ...any)   {}
func (m *mockLogger) Warnc(_ context.Context, _ string, _ ...any)   {}
func (m *mockLogger) Errorc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLogger) Fatalc(_ context.Context, _ string, _ ...any)  {}

// --- Tests ---

func TestCreateAnnouncementHandler_Handle(t *testing.T) {
	repo := &mockAnnouncementRepo{}
	eb := &mockEventBus{}
	handler := NewCreateAnnouncementHandler(repo, eb, &mockLogger{})

	now := time.Now()
	cmd := CreateAnnouncementCommand{
		Title:     shared.Lang{Uz: "title_uz", Ru: "title_ru", En: "title_en"},
		Content:   shared.Lang{Uz: "content_uz", Ru: "content_ru", En: "content_en"},
		Priority:  1,
		StartDate: &now,
		EndDate:   nil,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo.saved == nil {
		t.Fatal("expected announcement to be saved")
	}
	if repo.saved.Title().Uz != "title_uz" {
		t.Errorf("expected title_uz, got %s", repo.saved.Title().Uz)
	}
	if repo.saved.Content().En != "content_en" {
		t.Errorf("expected content_en, got %s", repo.saved.Content().En)
	}
	if repo.saved.Priority() != 1 {
		t.Errorf("expected priority 1, got %d", repo.saved.Priority())
	}
	if repo.saved.Published() {
		t.Error("expected announcement to be unpublished")
	}
	if repo.saved.StartDate() == nil {
		t.Error("expected start_date to be set")
	}
}

func TestCreateAnnouncementHandler_MinimalFields(t *testing.T) {
	repo := &mockAnnouncementRepo{}
	handler := NewCreateAnnouncementHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), CreateAnnouncementCommand{
		Title:   shared.Lang{Uz: "t", Ru: "t", En: "t"},
		Content: shared.Lang{Uz: "c", Ru: "c", En: "c"},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if repo.saved == nil {
		t.Fatal("expected announcement to be saved")
	}
	if repo.saved.StartDate() != nil {
		t.Error("expected start_date to be nil")
	}
	if repo.saved.EndDate() != nil {
		t.Error("expected end_date to be nil")
	}
}
