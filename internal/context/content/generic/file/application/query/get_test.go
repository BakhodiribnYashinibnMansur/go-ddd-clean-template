package query

import (
	"gct/internal/kernel/infrastructure/logger"
	"context"
	"errors"
	"testing"
	"time"

	fileentity "gct/internal/context/content/generic/file/domain/entity"
	filerepo "gct/internal/context/content/generic/file/domain/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockReadRepo struct {
	view  *filerepo.FileView
	views []*filerepo.FileView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id fileentity.FileID) (*filerepo.FileView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, fileentity.ErrFileNotFound
}

func (m *mockReadRepo) List(_ context.Context, _ filerepo.FileFilter) ([]*filerepo.FileView, int64, error) {
	return m.views, m.total, nil
}

type errorReadRepo struct{ err error }

func (m *errorReadRepo) FindByID(_ context.Context, _ fileentity.FileID) (*filerepo.FileView, error) {
	return nil, m.err
}

func (m *errorReadRepo) List(_ context.Context, _ filerepo.FileFilter) ([]*filerepo.FileView, int64, error) {
	return nil, 0, m.err
}

var errRepo = errors.New("repo failure")

// --- Tests: GetFile ---

func TestGetFileHandler_Handle(t *testing.T) {
	t.Parallel()

	id := fileentity.NewFileID()
	uploaderID := uuid.New()
	now := time.Now()
	readRepo := &mockReadRepo{
		view: &filerepo.FileView{
			ID:           id,
			Name:         "avatar.png",
			OriginalName: "my-avatar.png",
			MimeType:     "image/png",
			Size:         1024,
			Path:         "/uploads/avatar.png",
			URL:          "https://cdn.example.com/avatar.png",
			UploadedBy:   &uploaderID,
			CreatedAt:    now,
		},
	}

	handler := NewGetFileHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetFileQuery{ID: id})
	require.NoError(t, err)
	if result == nil {
		t.Fatal("expected result")
	}
	if result.Name != "avatar.png" {
		t.Errorf("expected name 'avatar.png', got %s", result.Name)
	}
	if result.MimeType != "image/png" {
		t.Errorf("expected mimeType 'image/png', got %s", result.MimeType)
	}
	if result.Size != 1024 {
		t.Errorf("expected size 1024, got %d", result.Size)
	}
	if result.UploadedBy == nil || *result.UploadedBy != uploaderID {
		t.Error("uploadedBy not mapped correctly")
	}
}

func TestGetFileHandler_NotFound(t *testing.T) {
	t.Parallel()

	readRepo := &mockReadRepo{}
	handler := NewGetFileHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetFileQuery{ID: fileentity.NewFileID()})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetFileHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &errorReadRepo{err: errRepo}
	handler := NewGetFileHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetFileQuery{ID: fileentity.NewFileID()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestGetFileHandler_AllFieldsMapped(t *testing.T) {
	t.Parallel()

	id := fileentity.NewFileID()
	uploaderID := uuid.New()
	now := time.Now()

	readRepo := &mockReadRepo{
		view: &filerepo.FileView{
			ID:           id,
			Name:         "report.pdf",
			OriginalName: "annual-report.pdf",
			MimeType:     "application/pdf",
			Size:         50000,
			Path:         "/uploads/report.pdf",
			URL:          "https://cdn.example.com/report.pdf",
			UploadedBy:   &uploaderID,
			CreatedAt:    now,
		},
	}

	handler := NewGetFileHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetFileQuery{ID: id})
	require.NoError(t, err)
	if result.OriginalName != "annual-report.pdf" {
		t.Errorf("expected originalName 'annual-report.pdf', got %s", result.OriginalName)
	}
	if result.Path != "/uploads/report.pdf" {
		t.Errorf("expected correct path, got %s", result.Path)
	}
	if result.URL != "https://cdn.example.com/report.pdf" {
		t.Errorf("expected correct URL, got %s", result.URL)
	}
}
