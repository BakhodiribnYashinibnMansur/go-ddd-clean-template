package file

import (
	"context"
	"testing"

	"gct/internal/context/content/generic/file/application/command"
	"gct/internal/context/content/generic/file/application/query"
	fileentity "gct/internal/context/content/generic/file/domain/entity"
	filerepo "gct/internal/context/content/generic/file/domain/repository"
)

func TestIntegration_FileCreateAndGet(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	// Create file metadata
	err := bc.CreateFile.Handle(ctx, command.CreateFileCommand{
		Name:         "report.pdf",
		OriginalName: "quarterly-report.pdf",
		MimeType:     "application/pdf",
		Size:         8192,
		Path:         "/uploads/report.pdf",
		URL:          "https://cdn.example.com/report.pdf",
		UploadedBy:   nil,
	})
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	// List to find the created file
	result, err := bc.ListFiles.Handle(ctx, query.ListFilesQuery{
		Filter: filerepo.FileFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 file, got %d", result.Total)
	}

	created := result.Files[0]

	// Get by ID
	got, err := bc.GetFile.Handle(ctx, query.GetFileQuery{ID: fileentity.FileID(created.ID)})
	if err != nil {
		t.Fatalf("GetFile failed: %v", err)
	}

	// Verify all fields match
	if got.ID != created.ID {
		t.Errorf("ID mismatch: got %s, want %s", got.ID, created.ID)
	}
	if got.Name != "report.pdf" {
		t.Errorf("Name = %q, want %q", got.Name, "report.pdf")
	}
	if got.OriginalName != "quarterly-report.pdf" {
		t.Errorf("OriginalName = %q, want %q", got.OriginalName, "quarterly-report.pdf")
	}
	if got.MimeType != "application/pdf" {
		t.Errorf("MimeType = %q, want %q", got.MimeType, "application/pdf")
	}
	if got.Size != 8192 {
		t.Errorf("Size = %d, want %d", got.Size, 8192)
	}
	if got.Path != "/uploads/report.pdf" {
		t.Errorf("Path = %q, want %q", got.Path, "/uploads/report.pdf")
	}
	if got.URL != "https://cdn.example.com/report.pdf" {
		t.Errorf("URL = %q, want %q", got.URL, "https://cdn.example.com/report.pdf")
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestIntegration_FileListPagination(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	// Create 3 files
	files := []command.CreateFileCommand{
		{
			Name:         "file-a.txt",
			OriginalName: "a.txt",
			MimeType:     "text/plain",
			Size:         100,
			Path:         "/uploads/file-a.txt",
			URL:          "https://cdn.example.com/file-a.txt",
		},
		{
			Name:         "file-b.jpg",
			OriginalName: "b.jpg",
			MimeType:     "image/jpeg",
			Size:         2048,
			Path:         "/uploads/file-b.jpg",
			URL:          "https://cdn.example.com/file-b.jpg",
		},
		{
			Name:         "file-c.pdf",
			OriginalName: "c.pdf",
			MimeType:     "application/pdf",
			Size:         4096,
			Path:         "/uploads/file-c.pdf",
			URL:          "https://cdn.example.com/file-c.pdf",
		},
	}

	for _, cmd := range files {
		err := bc.CreateFile.Handle(ctx, cmd)
		if err != nil {
			t.Fatalf("CreateFile %s: %v", cmd.Name, err)
		}
	}

	// List with limit=2 (first page)
	page1, err := bc.ListFiles.Handle(ctx, query.ListFilesQuery{
		Filter: filerepo.FileFilter{Limit: 2, Offset: 0},
	})
	if err != nil {
		t.Fatalf("ListFiles page 1: %v", err)
	}
	if page1.Total != 3 {
		t.Errorf("Total = %d, want 3", page1.Total)
	}
	if len(page1.Files) != 2 {
		t.Errorf("page 1 files count = %d, want 2", len(page1.Files))
	}

	// List with limit=2, offset=2 (second page)
	page2, err := bc.ListFiles.Handle(ctx, query.ListFilesQuery{
		Filter: filerepo.FileFilter{Limit: 2, Offset: 2},
	})
	if err != nil {
		t.Fatalf("ListFiles page 2: %v", err)
	}
	if page2.Total != 3 {
		t.Errorf("Total = %d, want 3", page2.Total)
	}
	if len(page2.Files) != 1 {
		t.Errorf("page 2 files count = %d, want 1", len(page2.Files))
	}

	// Verify no overlap between pages
	page1IDs := make(map[string]bool)
	for _, f := range page1.Files {
		page1IDs[f.ID.String()] = true
	}
	for _, f := range page2.Files {
		if page1IDs[f.ID.String()] {
			t.Errorf("file %s appears on both pages", f.ID)
		}
	}
}
