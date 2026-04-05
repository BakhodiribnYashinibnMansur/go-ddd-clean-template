package domain_test

import (
	"testing"
	"time"

	domain "gct/internal/context/content/file/domain"

	"github.com/google/uuid"
)

func TestNewFile(t *testing.T) {
	uploaderID := uuid.New()
	f := domain.NewFile("stored_abc.png", "photo.png", "image/png", 1024, "/uploads/stored_abc.png", "https://cdn.example.com/stored_abc.png", &uploaderID)

	if f.Name() != "stored_abc.png" {
		t.Fatalf("expected name stored_abc.png, got %s", f.Name())
	}
	if f.OriginalName() != "photo.png" {
		t.Fatalf("expected original name photo.png, got %s", f.OriginalName())
	}
	if f.MimeType() != "image/png" {
		t.Fatalf("expected mime type image/png, got %s", f.MimeType())
	}
	if f.Size() != 1024 {
		t.Fatalf("expected size 1024, got %d", f.Size())
	}
	if f.Path() != "/uploads/stored_abc.png" {
		t.Fatalf("expected path /uploads/stored_abc.png, got %s", f.Path())
	}
	if f.URL() != "https://cdn.example.com/stored_abc.png" {
		t.Fatalf("expected URL https://cdn.example.com/stored_abc.png, got %s", f.URL())
	}
	if f.UploadedBy() == nil || *f.UploadedBy() != uploaderID {
		t.Fatal("uploaded_by mismatch")
	}

	events := f.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventName() != "file.uploaded" {
		t.Fatalf("expected file.uploaded, got %s", events[0].EventName())
	}
}

func TestNewFile_NoUploader(t *testing.T) {
	f := domain.NewFile("doc.pdf", "document.pdf", "application/pdf", 2048, "/uploads/doc.pdf", "https://cdn.example.com/doc.pdf", nil)

	if f.UploadedBy() != nil {
		t.Fatal("uploaded_by should be nil")
	}
}

func TestReconstructFile(t *testing.T) {
	id := uuid.New()
	createdAt := time.Now()
	uploaderID := uuid.New()

	f := domain.ReconstructFile(id, createdAt, "stored.jpg", "original.jpg", "image/jpeg", 512, "/uploads/stored.jpg", "https://cdn.example.com/stored.jpg", &uploaderID)

	if f.ID() != id {
		t.Fatal("ID mismatch")
	}
	if f.Name() != "stored.jpg" {
		t.Fatal("name mismatch")
	}
	if len(f.Events()) != 0 {
		t.Fatalf("expected 0 events, got %d", len(f.Events()))
	}
}
