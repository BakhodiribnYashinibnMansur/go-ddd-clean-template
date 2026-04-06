package query

import (
	"gct/internal/kernel/infrastructure/logger"
	"context"
	"testing"
	"time"

	fileentity "gct/internal/context/content/generic/file/domain/entity"
	filerepo "gct/internal/context/content/generic/file/domain/repository"

	"github.com/stretchr/testify/require"
)

func TestListFilesHandler_Handle(t *testing.T) {
	t.Parallel()

	now := time.Now()
	readRepo := &mockReadRepo{
		views: []*filerepo.FileView{
			{ID: fileentity.NewFileID(), Name: "file1.png", OriginalName: "f1.png", MimeType: "image/png", Size: 100, CreatedAt: now},
			{ID: fileentity.NewFileID(), Name: "file2.pdf", OriginalName: "f2.pdf", MimeType: "application/pdf", Size: 200, CreatedAt: now},
		},
		total: 2,
	}

	handler := NewListFilesHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListFilesQuery{
		Filter: filerepo.FileFilter{Limit: 10, Offset: 0},
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

	readRepo := &mockReadRepo{views: []*filerepo.FileView{}, total: 0}

	handler := NewListFilesHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListFilesQuery{
		Filter: filerepo.FileFilter{},
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
	_, err := handler.Handle(context.Background(), ListFilesQuery{Filter: filerepo.FileFilter{}})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
