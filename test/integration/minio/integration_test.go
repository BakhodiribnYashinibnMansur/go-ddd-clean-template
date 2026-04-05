package minio

import (
	"context"
	"testing"

	"gct/internal/context/content/file"
	"gct/internal/context/content/file/application/command"
	"gct/internal/context/content/file/application/query"
	"gct/internal/context/content/file/domain"
	"gct/internal/kernel/infrastructure/eventbus"
	"gct/internal/kernel/infrastructure/logger"
	"gct/test/integration/common/setup"
)

func newTestBC(t *testing.T) *file.BoundedContext {
	t.Helper()
	eb := eventbus.NewInMemoryEventBus()
	l := logger.New("error")
	return file.NewBoundedContext(setup.TestPG.Pool, eb, l)
}

func TestIntegration_CreateAndGetFile(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateFile.Handle(ctx, command.CreateFileCommand{
		Name:         "test-image.jpg",
		OriginalName: "photo.jpg",
		MimeType:     "image/jpeg",
		Size:         1024,
		Path:         "/uploads/test-image.jpg",
		URL:          "https://cdn.example.com/test-image.jpg",
	})
	if err != nil {
		t.Fatalf("CreateFile: %v", err)
	}

	result, err := bc.ListFiles.Handle(ctx, query.ListFilesQuery{
		Filter: domain.FileFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListFiles: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 file, got %d", result.Total)
	}

	f := result.Files[0]
	if f.Name != "test-image.jpg" {
		t.Errorf("expected name test-image.jpg, got %s", f.Name)
	}
	if f.MimeType != "image/jpeg" {
		t.Errorf("expected mime image/jpeg, got %s", f.MimeType)
	}

	view, err := bc.GetFile.Handle(ctx, query.GetFileQuery{ID: f.ID})
	if err != nil {
		t.Fatalf("GetFile: %v", err)
	}
	if view.ID != f.ID {
		t.Errorf("ID mismatch: %s vs %s", view.ID, f.ID)
	}
}

func TestIntegration_ListFilesWithFilter(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	// Create two files with different mime types
	_ = bc.CreateFile.Handle(ctx, command.CreateFileCommand{
		Name:         "doc.pdf",
		OriginalName: "document.pdf",
		MimeType:     "application/pdf",
		Size:         2048,
		Path:         "/uploads/doc.pdf",
		URL:          "https://cdn.example.com/doc.pdf",
	})
	_ = bc.CreateFile.Handle(ctx, command.CreateFileCommand{
		Name:         "image.png",
		OriginalName: "screenshot.png",
		MimeType:     "image/png",
		Size:         4096,
		Path:         "/uploads/image.png",
		URL:          "https://cdn.example.com/image.png",
	})

	result, err := bc.ListFiles.Handle(ctx, query.ListFilesQuery{
		Filter: domain.FileFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListFiles: %v", err)
	}
	if result.Total != 2 {
		t.Fatalf("expected 2 files, got %d", result.Total)
	}
}
