package file

import (
	"context"
	"testing"

	"gct/internal/file"
	"gct/internal/file/application/command"
	"gct/internal/file/application/query"
	"gct/internal/file/domain"
	"gct/internal/shared/infrastructure/eventbus"
	"gct/internal/shared/infrastructure/logger"
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
		Name:         "test-file.png",
		OriginalName: "photo.png",
		MimeType:     "image/png",
		Size:         1024,
		Path:         "/uploads/test-file.png",
		URL:          "https://cdn.example.com/test-file.png",
		UploadedBy:   nil,
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
	if f.Name != "test-file.png" {
		t.Errorf("expected name test-file.png, got %s", f.Name)
	}
	if f.MimeType != "image/png" {
		t.Errorf("expected mime type image/png, got %s", f.MimeType)
	}

	view, err := bc.GetFile.Handle(ctx, query.GetFileQuery{ID: f.ID})
	if err != nil {
		t.Fatalf("GetFile: %v", err)
	}
	if view.ID != f.ID {
		t.Errorf("ID mismatch: %s vs %s", view.ID, f.ID)
	}
}

func TestIntegration_CreateMultipleAndListFiles(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	files := []command.CreateFileCommand{
		{
			Name:         "doc1.pdf",
			OriginalName: "document.pdf",
			MimeType:     "application/pdf",
			Size:         2048,
			Path:         "/uploads/doc1.pdf",
			URL:          "https://cdn.example.com/doc1.pdf",
		},
		{
			Name:         "img1.jpg",
			OriginalName: "image.jpg",
			MimeType:     "image/jpeg",
			Size:         4096,
			Path:         "/uploads/img1.jpg",
			URL:          "https://cdn.example.com/img1.jpg",
		},
	}

	for _, cmd := range files {
		err := bc.CreateFile.Handle(ctx, cmd)
		if err != nil {
			t.Fatalf("CreateFile %s: %v", cmd.Name, err)
		}
	}

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
