package query

import (
	"gct/internal/shared/infrastructure/logger"
	"context"
	"testing"
	"time"

	"gct/internal/file/domain"

	"github.com/google/uuid"
)

func TestListFilesHandler_Handle(t *testing.T) {
	now := time.Now()
	readRepo := &mockReadRepo{
		views: []*domain.FileView{
			{ID: uuid.New(), Name: "file1.png", OriginalName: "f1.png", MimeType: "image/png", Size: 100, CreatedAt: now},
			{ID: uuid.New(), Name: "file2.pdf", OriginalName: "f2.pdf", MimeType: "application/pdf", Size: 200, CreatedAt: now},
		},
		total: 2,
	}

	handler := NewListFilesHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListFilesQuery{
		Filter: domain.FileFilter{Limit: 10, Offset: 0},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
	if len(result.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(result.Files))
	}
	if result.Files[0].Name != "file1.png" {
		t.Errorf("expected 'file1.png', got %s", result.Files[0].Name)
	}
}

func TestListFilesHandler_Empty(t *testing.T) {
	readRepo := &mockReadRepo{views: []*domain.FileView{}, total: 0}

	handler := NewListFilesHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListFilesQuery{
		Filter: domain.FileFilter{},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
	if len(result.Files) != 0 {
		t.Errorf("expected 0 files, got %d", len(result.Files))
	}
}

func TestListFilesHandler_RepoError(t *testing.T) {
	readRepo := &errorReadRepo{err: errRepo}
	handler := NewListFilesHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), ListFilesQuery{Filter: domain.FileFilter{}})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
