package dataexport

import (
	"context"
	"testing"

	"gct/internal/context/admin/supporting/dataexport"
	"gct/internal/context/admin/supporting/dataexport/application/command"
	"gct/internal/context/admin/supporting/dataexport/application/query"
	"gct/internal/context/admin/supporting/dataexport/domain"
	"gct/internal/kernel/infrastructure/eventbus"
	"gct/internal/kernel/infrastructure/logger"
	"gct/test/integration/common/setup"

	"github.com/google/uuid"
)

func newTestBC(t *testing.T) *dataexport.BoundedContext {
	t.Helper()
	eb := eventbus.NewInMemoryEventBus()
	l := logger.New("error")
	return dataexport.NewBoundedContext(setup.TestPG.Pool, eb, l)
}

func TestIntegration_CreateAndGetDataExport(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	err := bc.CreateDataExport.Handle(ctx, command.CreateDataExportCommand{
		UserID:   userID,
		DataType: "users",
		Format:   "csv",
	})
	if err != nil {
		t.Fatalf("CreateDataExport: %v", err)
	}

	result, err := bc.ListDataExports.Handle(ctx, query.ListDataExportsQuery{
		Filter: domain.DataExportFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListDataExports: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 data export, got %d", result.Total)
	}

	de := result.Exports[0]
	if de.DataType != "users" {
		t.Errorf("expected data type users, got %s", de.DataType)
	}
	if de.Status != "PENDING" {
		t.Errorf("expected status PENDING, got %s", de.Status)
	}

	view, err := bc.GetDataExport.Handle(ctx, query.GetDataExportQuery{ID: domain.DataExportID(de.ID)})
	if err != nil {
		t.Fatalf("GetDataExport: %v", err)
	}
	if view.ID != de.ID {
		t.Errorf("ID mismatch: %s vs %s", view.ID, de.ID)
	}
}

func TestIntegration_UpdateDataExport(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	err := bc.CreateDataExport.Handle(ctx, command.CreateDataExportCommand{
		UserID:   userID,
		DataType: "audit_logs",
		Format:   "json",
	})
	if err != nil {
		t.Fatalf("CreateDataExport: %v", err)
	}

	list, _ := bc.ListDataExports.Handle(ctx, query.ListDataExportsQuery{
		Filter: domain.DataExportFilter{Limit: 10},
	})
	deID := domain.DataExportID(list.Exports[0].ID)

	processing := domain.ExportStatusProcessing
	err = bc.UpdateDataExport.Handle(ctx, command.UpdateDataExportCommand{
		ID:     domain.DataExportID(deID),
		Status: &processing,
	})
	if err != nil {
		t.Fatalf("UpdateDataExport (processing): %v", err)
	}

	view, _ := bc.GetDataExport.Handle(ctx, query.GetDataExportQuery{ID: domain.DataExportID(deID)})
	if view.Status != "PROCESSING" {
		t.Errorf("expected status PROCESSING, got %s", view.Status)
	}

	completed := domain.ExportStatusCompleted
	fileURL := "https://storage.example.com/exports/audit_logs.json"
	err = bc.UpdateDataExport.Handle(ctx, command.UpdateDataExportCommand{
		ID:      domain.DataExportID(deID),
		Status:  &completed,
		FileURL: &fileURL,
	})
	if err != nil {
		t.Fatalf("UpdateDataExport (completed): %v", err)
	}

	view, _ = bc.GetDataExport.Handle(ctx, query.GetDataExportQuery{ID: domain.DataExportID(deID)})
	if view.Status != "COMPLETED" {
		t.Errorf("expected status COMPLETED, got %s", view.Status)
	}
}

func TestIntegration_DeleteDataExport(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	err := bc.CreateDataExport.Handle(ctx, command.CreateDataExportCommand{
		UserID:   userID,
		DataType: "transactions",
		Format:   "csv",
	})
	if err != nil {
		t.Fatalf("CreateDataExport: %v", err)
	}

	list, _ := bc.ListDataExports.Handle(ctx, query.ListDataExportsQuery{
		Filter: domain.DataExportFilter{Limit: 10},
	})
	deID := domain.DataExportID(list.Exports[0].ID)

	err = bc.DeleteDataExport.Handle(ctx, command.DeleteDataExportCommand{ID: domain.DataExportID(deID)})
	if err != nil {
		t.Fatalf("DeleteDataExport: %v", err)
	}

	list2, _ := bc.ListDataExports.Handle(ctx, query.ListDataExportsQuery{
		Filter: domain.DataExportFilter{Limit: 10},
	})
	if list2.Total != 0 {
		t.Errorf("expected 0 data exports after delete, got %d", list2.Total)
	}
}

func TestIntegration_DataExportStateMachine(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	// Step 1: Create export (PENDING)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	err := bc.CreateDataExport.Handle(ctx, command.CreateDataExportCommand{
		UserID:   userID,
		DataType: "users",
		Format:   "csv",
	})
	if err != nil {
		t.Fatalf("CreateDataExport: %v", err)
	}

	list, err := bc.ListDataExports.Handle(ctx, query.ListDataExportsQuery{
		Filter: domain.DataExportFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListDataExports: %v", err)
	}
	if list.Total != 1 {
		t.Fatalf("expected 1 data export, got %d", list.Total)
	}
	deID := domain.DataExportID(list.Exports[0].ID)

	view, err := bc.GetDataExport.Handle(ctx, query.GetDataExportQuery{ID: deID})
	if err != nil {
		t.Fatalf("GetDataExport (pending): %v", err)
	}
	if view.Status != "PENDING" {
		t.Errorf("expected status PENDING, got %s", view.Status)
	}

	// Step 2: Update to PROCESSING
	processing := domain.ExportStatusProcessing
	err = bc.UpdateDataExport.Handle(ctx, command.UpdateDataExportCommand{
		ID:     deID,
		Status: &processing,
	})
	if err != nil {
		t.Fatalf("UpdateDataExport (processing): %v", err)
	}

	view, err = bc.GetDataExport.Handle(ctx, query.GetDataExportQuery{ID: deID})
	if err != nil {
		t.Fatalf("GetDataExport (processing): %v", err)
	}
	if view.Status != "PROCESSING" {
		t.Errorf("expected status PROCESSING, got %s", view.Status)
	}

	// Step 3: Update to COMPLETED with fileURL
	completed := domain.ExportStatusCompleted
	fileURL := "https://storage.example.com/exports/users.csv"
	err = bc.UpdateDataExport.Handle(ctx, command.UpdateDataExportCommand{
		ID:      deID,
		Status:  &completed,
		FileURL: &fileURL,
	})
	if err != nil {
		t.Fatalf("UpdateDataExport (completed): %v", err)
	}

	// Step 4: Verify fileURL in GetDataExport
	view, err = bc.GetDataExport.Handle(ctx, query.GetDataExportQuery{ID: deID})
	if err != nil {
		t.Fatalf("GetDataExport (completed): %v", err)
	}
	if view.Status != "COMPLETED" {
		t.Errorf("expected status COMPLETED, got %s", view.Status)
	}
	if view.FileURL == nil {
		t.Fatal("expected FileURL to be set, got nil")
	}
	if *view.FileURL != fileURL {
		t.Errorf("expected FileURL %q, got %q", fileURL, *view.FileURL)
	}
}

func TestIntegration_DataExportFail(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	// Step 1: Create export (PENDING)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	err := bc.CreateDataExport.Handle(ctx, command.CreateDataExportCommand{
		UserID:   userID,
		DataType: "audit_logs",
		Format:   "json",
	})
	if err != nil {
		t.Fatalf("CreateDataExport: %v", err)
	}

	list, err := bc.ListDataExports.Handle(ctx, query.ListDataExportsQuery{
		Filter: domain.DataExportFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListDataExports: %v", err)
	}
	deID := domain.DataExportID(list.Exports[0].ID)

	// Step 2: Update to PROCESSING
	processing := domain.ExportStatusProcessing
	err = bc.UpdateDataExport.Handle(ctx, command.UpdateDataExportCommand{
		ID:     deID,
		Status: &processing,
	})
	if err != nil {
		t.Fatalf("UpdateDataExport (processing): %v", err)
	}

	// Step 3: Update to FAILED with error message
	failed := domain.ExportStatusFailed
	errMsg := "disk quota exceeded"
	err = bc.UpdateDataExport.Handle(ctx, command.UpdateDataExportCommand{
		ID:     deID,
		Status: &failed,
		Error:  &errMsg,
	})
	if err != nil {
		t.Fatalf("UpdateDataExport (failed): %v", err)
	}

	// Step 4: Verify status is FAILED
	view, err := bc.GetDataExport.Handle(ctx, query.GetDataExportQuery{ID: deID})
	if err != nil {
		t.Fatalf("GetDataExport (failed): %v", err)
	}
	if view.Status != "FAILED" {
		t.Errorf("expected status FAILED, got %s", view.Status)
	}
	// Note: error_message column doesn't exist in DB schema yet,
	// so view.Error is always nil from the read repo.
}
