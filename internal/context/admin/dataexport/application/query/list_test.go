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

func TestListDataExportsHandler_Success(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	now := time.Now()

	readRepo := &mockDataExportReadRepository{
		listViews: []*domain.DataExportView{
			{
				ID:        uuid.New(),
				UserID:    userID,
				DataType:  "users",
				Format:    "csv",
				Status:    "completed",
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				ID:        uuid.New(),
				UserID:    userID,
				DataType:  "audit_logs",
				Format:    "json",
				Status:    "pending",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		listTotal: 2,
	}

	handler := NewListDataExportsHandler(readRepo, logger.Noop())

	q := ListDataExportsQuery{
		Filter: domain.DataExportFilter{},
	}

	result, err := handler.Handle(context.Background(), q)
	require.NoError(t, err)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}

	if len(result.Exports) != 2 {
		t.Fatalf("expected 2 exports, got %d", len(result.Exports))
	}

	if result.Exports[0].DataType != "users" {
		t.Errorf("expected data type 'users', got '%s'", result.Exports[0].DataType)
	}

	if result.Exports[1].Status != "pending" {
		t.Errorf("expected status 'pending', got '%s'", result.Exports[1].Status)
	}
}

func TestListDataExportsHandler_Empty(t *testing.T) {
	t.Parallel()

	readRepo := &mockDataExportReadRepository{
		listViews: []*domain.DataExportView{},
		listTotal: 0,
	}

	handler := NewListDataExportsHandler(readRepo, logger.Noop())

	q := ListDataExportsQuery{
		Filter: domain.DataExportFilter{},
	}

	result, err := handler.Handle(context.Background(), q)
	require.NoError(t, err)

	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}

	if len(result.Exports) != 0 {
		t.Errorf("expected 0 exports, got %d", len(result.Exports))
	}
}

func TestListDataExportsHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &mockDataExportReadRepository{
		listErr: errors.New("database unavailable"),
	}

	handler := NewListDataExportsHandler(readRepo, logger.Noop())

	q := ListDataExportsQuery{
		Filter: domain.DataExportFilter{},
	}

	result, err := handler.Handle(context.Background(), q)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if result != nil {
		t.Error("expected nil result on error")
	}
}
