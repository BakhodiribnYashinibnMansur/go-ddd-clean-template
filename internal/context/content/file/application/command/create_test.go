package command_test

import (
	"context"
	"testing"

	"gct/internal/context/content/file/application/command"
	"gct/internal/context/content/file/domain"
	"gct/internal/kernel/application"
	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockFileRepo struct {
	saved *domain.File
}

func (m *mockFileRepo) Save(_ context.Context, f *domain.File) error {
	m.saved = f
	return nil
}

type mockEventBus struct{}

func (m *mockEventBus) Publish(_ context.Context, _ ...shared.DomainEvent) error { return nil }
func (m *mockEventBus) Subscribe(_ string, _ application.EventHandler) error     { return nil }

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

func TestCreateFileHandler_Handle(t *testing.T) {
	t.Parallel()

	repo := &mockFileRepo{}
	handler := command.NewCreateFileHandler(repo, &mockEventBus{}, &mockLogger{})

	uploaderID := uuid.New()
	cmd := command.CreateFileCommand{
		Name:         "avatar.png",
		OriginalName: "my-avatar.png",
		MimeType:     "image/png",
		Size:         1024,
		Path:         "/uploads/avatar.png",
		URL:          "https://cdn.example.com/avatar.png",
		UploadedBy:   &uploaderID,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.saved == nil {
		t.Fatal("expected file to be saved")
	}
	if repo.saved.Name() != "avatar.png" {
		t.Fatalf("expected name 'avatar.png', got %s", repo.saved.Name())
	}
	if repo.saved.OriginalName() != "my-avatar.png" {
		t.Fatalf("expected original name 'my-avatar.png', got %s", repo.saved.OriginalName())
	}
	if repo.saved.MimeType() != "image/png" {
		t.Fatalf("expected mimeType 'image/png', got %s", repo.saved.MimeType())
	}
	if repo.saved.Size() != 1024 {
		t.Fatalf("expected size 1024, got %d", repo.saved.Size())
	}
	if repo.saved.Path() != "/uploads/avatar.png" {
		t.Fatalf("expected path '/uploads/avatar.png', got %s", repo.saved.Path())
	}
	if repo.saved.URL() != "https://cdn.example.com/avatar.png" {
		t.Fatalf("expected correct URL, got %s", repo.saved.URL())
	}
	if repo.saved.UploadedBy() == nil || *repo.saved.UploadedBy() != uploaderID {
		t.Fatalf("expected uploadedBy %s", uploaderID)
	}
}

func TestCreateFileHandler_NilUploadedBy(t *testing.T) {
	t.Parallel()

	repo := &mockFileRepo{}
	handler := command.NewCreateFileHandler(repo, &mockEventBus{}, &mockLogger{})

	cmd := command.CreateFileCommand{
		Name:         "doc.pdf",
		OriginalName: "document.pdf",
		MimeType:     "application/pdf",
		Size:         2048,
		Path:         "/uploads/doc.pdf",
		URL:          "https://cdn.example.com/doc.pdf",
		UploadedBy:   nil,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.saved == nil {
		t.Fatal("expected file to be saved")
	}
	if repo.saved.UploadedBy() != nil {
		t.Fatal("expected nil uploadedBy")
	}
}
