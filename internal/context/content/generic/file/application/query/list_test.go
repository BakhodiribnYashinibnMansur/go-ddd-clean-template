package query

import (
	"gct/internal/kernel/infrastructure/logger"
	"context"
	"testing"
	"time"

	"gct/internal/context/content/generic/file/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestListFilesHandler_Handle(t *testing.T) {
	t.Parallel()

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
	require.NoError(t, err)
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
	t.Parallel()

	readRepo := &mockReadRepo{views: []*domain.FileView{}, total: 0}

	handler := NewListFilesHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListFilesQuery{
		Filter: domain.FileFilter{},
	})
	require.NoError(t, err)
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
	if len(result.Files) != 0 {
		t.Errorf("expected 0 files, got %d", len(result.Files))
	}
}

func TestListFilesHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &errorReadRepo{err: errRepo}
	handler := NewListFilesHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), ListFilesQuery{Filter: domain.FileFilter{}})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
