package query

import (
	"gct/internal/platform/infrastructure/logger"
	"context"
	"testing"
	"time"

	"gct/internal/context/content/announcement/domain"

	"github.com/google/uuid"
)

func TestListAnnouncementsHandler_Handle(t *testing.T) {
	readRepo := &mockReadRepo{
		views: []*domain.AnnouncementView{
			{ID: uuid.New(), TitleEn: "A1", Priority: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: uuid.New(), TitleEn: "A2", Priority: 2, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
		total: 2,
	}

	handler := NewListAnnouncementsHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListAnnouncementsQuery{
		Filter: domain.AnnouncementFilter{Limit: 10, Offset: 0},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
	if len(result.Announcements) != 2 {
		t.Fatalf("expected 2 announcements, got %d", len(result.Announcements))
	}
	if result.Announcements[0].Title.En != "A1" {
		t.Errorf("expected A1, got %s", result.Announcements[0].Title.En)
	}
}

func TestListAnnouncementsHandler_Empty(t *testing.T) {
	readRepo := &mockReadRepo{views: []*domain.AnnouncementView{}, total: 0}

	handler := NewListAnnouncementsHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListAnnouncementsQuery{
		Filter: domain.AnnouncementFilter{},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
	if len(result.Announcements) != 0 {
		t.Errorf("expected 0 announcements, got %d", len(result.Announcements))
	}
}

func TestListAnnouncementsHandler_WithFilters(t *testing.T) {
	published := true
	readRepo := &mockReadRepo{
		views: []*domain.AnnouncementView{
			{ID: uuid.New(), TitleEn: "Pub", Published: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
		total: 1,
	}

	handler := NewListAnnouncementsHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListAnnouncementsQuery{
		Filter: domain.AnnouncementFilter{Published: &published, Limit: 10},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
}

func TestListAnnouncementsHandler_RepoError(t *testing.T) {
	readRepo := &errorReadRepo{err: errRepo}
	handler := NewListAnnouncementsHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), ListAnnouncementsQuery{Filter: domain.AnnouncementFilter{}})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
