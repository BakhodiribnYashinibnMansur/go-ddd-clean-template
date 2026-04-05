package query

import (
	"gct/internal/kernel/infrastructure/logger"
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/context/admin/dataexport/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock DataExportReadRepository (inline, shared across query tests)
// ---------------------------------------------------------------------------

type mockDataExportReadRepository struct {
	findByIDView *domain.DataExportView
	findByIDErr  error
	listViews    []*domain.DataExportView
	listTotal    int64
	listErr      error
}

func (m *mockDataExportReadRepository) FindByID(_ context.Context, _ uuid.UUID) (*domain.DataExportView, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	return m.findByIDView, nil
}

func (m *mockDataExportReadRepository) List(_ context.Context, _ domain.DataExportFilter) ([]*domain.DataExportView, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.listViews, m.listTotal, nil
}

// ---------------------------------------------------------------------------
// Tests: GetDataExportHandler
// ---------------------------------------------------------------------------

func TestGetDataExportHandler_Success(t *testing.T) {
	t.Parallel()

	exportID := uuid.New()
	userID := uuid.New()
	now := time.Now()
	fileURL := "https://cdn.example.com/exports/data.csv"

	readRepo := &mockDataExportReadRepository{
		findByIDView: &domain.DataExportView{
			ID:        exportID,
			UserID:    userID,
			DataType:  "users",
			Format:    "csv",
			Status:    "completed",
			FileURL:   &fileURL,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	handler := NewGetDataExportHandler(readRepo, logger.Noop())

	result, err := handler.Handle(context.Background(), GetDataExportQuery{ID: domain.DataExportID(exportID)})
	require.NoError(t, err)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.ID != exportID {
		t.Errorf("expected ID %s, got %s", exportID, result.ID)
	}

	if result.Status != "completed" {
		t.Errorf("expected status 'completed', got '%s'", result.Status)
	}
}

func TestGetDataExportHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &mockDataExportReadRepository{
		findByIDErr: errors.New("not found"),
	}

	handler := NewGetDataExportHandler(readRepo, logger.Noop())

	result, err := handler.Handle(context.Background(), GetDataExportQuery{ID: domain.DataExportID(uuid.New())})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestGetDataExportHandler_MapsAllFields(t *testing.T) {
	t.Parallel()

	exportID := uuid.New()
	userID := uuid.New()
	now := time.Now()
	fileURL := "https://cdn.example.com/exports/data.json"
	errMsg := "timeout"

	readRepo := &mockDataExportReadRepository{
		findByIDView: &domain.DataExportView{
			ID:        exportID,
			UserID:    userID,
			DataType:  "audit_logs",
			Format:    "json",
			Status:    "failed",
			FileURL:   &fileURL,
			Error:     &errMsg,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	handler := NewGetDataExportHandler(readRepo, logger.Noop())

	result, err := handler.Handle(context.Background(), GetDataExportQuery{ID: domain.DataExportID(exportID)})
	require.NoError(t, err)

	v := result

	if v.ID != exportID {
		t.Error("ID mismatch")
	}
	if v.UserID != userID {
		t.Error("userID mismatch")
	}
	if v.DataType != "audit_logs" {
		t.Errorf("expected data type 'audit_logs', got '%s'", v.DataType)
	}
	if v.Format != "json" {
		t.Errorf("expected format 'json', got '%s'", v.Format)
	}
	if v.Status != "failed" {
		t.Errorf("expected status 'failed', got '%s'", v.Status)
	}
	if v.FileURL == nil || *v.FileURL != fileURL {
		t.Error("fileURL mismatch")
	}
	if v.Error == nil || *v.Error != "timeout" {
		t.Error("error mismatch")
	}
	if v.CreatedAt.IsZero() {
		t.Error("createdAt should not be zero")
	}
	if v.UpdatedAt.IsZero() {
		t.Error("updatedAt should not be zero")
	}
}
