package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/context/admin/supporting/dataexport/application/command"
	exportentity "gct/internal/context/admin/supporting/dataexport/domain/entity"
	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUpdateDataExportHandler_StatusProcessing(t *testing.T) {
	t.Parallel()

	exportID := exportentity.NewDataExportID()
	existing := exportentity.ReconstructDataExport(
		exportID.UUID(), time.Now(), time.Now(),
		uuid.New(), "users", "csv", exportentity.ExportStatusPending, nil, nil,
	)

	repo := &mockWriteRepo{
		findByIDFn: func(_ context.Context, id exportentity.DataExportID) (*exportentity.DataExport, error) {
			if id == exportID {
				return existing, nil
			}
			return nil, exportentity.ErrDataExportNotFound
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := command.NewUpdateDataExportHandler(repo, eb, l)

	status := exportentity.ExportStatusProcessing
	cmd := command.UpdateDataExportCommand{
		ID:     exportentity.DataExportID(exportID),
		Status: &status,
	}

	err := h.Handle(context.Background(), cmd)
	require.NoError(t, err)
	if repo.updatedEntity == nil {
		t.Fatal("expected entity to be updated")
	}
	if repo.updatedEntity.Status() != exportentity.ExportStatusProcessing {
		t.Fatalf("expected status PROCESSING, got %s", repo.updatedEntity.Status())
	}
}

func TestUpdateDataExportHandler_StatusCompleted(t *testing.T) {
	t.Parallel()

	exportID := exportentity.NewDataExportID()
	existing := exportentity.ReconstructDataExport(
		exportID.UUID(), time.Now(), time.Now(),
		uuid.New(), "users", "csv", exportentity.ExportStatusProcessing, nil, nil,
	)

	repo := &mockWriteRepo{
		findByIDFn: func(_ context.Context, id exportentity.DataExportID) (*exportentity.DataExport, error) {
			if id == exportID {
				return existing, nil
			}
			return nil, exportentity.ErrDataExportNotFound
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := command.NewUpdateDataExportHandler(repo, eb, l)

	status := exportentity.ExportStatusCompleted
	fileURL := "https://example.com/export.csv"
	cmd := command.UpdateDataExportCommand{
		ID:      exportentity.DataExportID(exportID),
		Status:  &status,
		FileURL: &fileURL,
	}

	err := h.Handle(context.Background(), cmd)
	require.NoError(t, err)
	if repo.updatedEntity.Status() != exportentity.ExportStatusCompleted {
		t.Fatalf("expected status COMPLETED, got %s", repo.updatedEntity.Status())
	}
	if repo.updatedEntity.FileURL() == nil || *repo.updatedEntity.FileURL() != fileURL {
		t.Fatalf("expected file URL %s, got %v", fileURL, repo.updatedEntity.FileURL())
	}
	if len(eb.publishedEvents) == 0 {
		t.Fatal("expected at least one event published")
	}
}

func TestUpdateDataExportHandler_StatusFailed(t *testing.T) {
	t.Parallel()

	exportID := exportentity.NewDataExportID()
	existing := exportentity.ReconstructDataExport(
		exportID.UUID(), time.Now(), time.Now(),
		uuid.New(), "orders", "xlsx", exportentity.ExportStatusProcessing, nil, nil,
	)

	repo := &mockWriteRepo{
		findByIDFn: func(_ context.Context, id exportentity.DataExportID) (*exportentity.DataExport, error) {
			if id == exportID {
				return existing, nil
			}
			return nil, exportentity.ErrDataExportNotFound
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := command.NewUpdateDataExportHandler(repo, eb, l)

	status := exportentity.ExportStatusFailed
	errMsg := "disk full"
	cmd := command.UpdateDataExportCommand{
		ID:     exportentity.DataExportID(exportID),
		Status: &status,
		Error:  &errMsg,
	}

	err := h.Handle(context.Background(), cmd)
	require.NoError(t, err)
	if repo.updatedEntity.Status() != exportentity.ExportStatusFailed {
		t.Fatalf("expected status FAILED, got %s", repo.updatedEntity.Status())
	}
	if repo.updatedEntity.Error() == nil || *repo.updatedEntity.Error() != errMsg {
		t.Fatalf("expected error msg %q, got %v", errMsg, repo.updatedEntity.Error())
	}
}

func TestUpdateDataExportHandler_NotFound(t *testing.T) {
	t.Parallel()

	repo := &mockWriteRepo{}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := command.NewUpdateDataExportHandler(repo, eb, l)

	status := exportentity.ExportStatusProcessing
	cmd := command.UpdateDataExportCommand{
		ID:     exportentity.DataExportID(uuid.New()),
		Status: &status,
	}

	err := h.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, exportentity.ErrDataExportNotFound) {
		t.Fatalf("expected ErrDataExportNotFound, got %v", err)
	}
}

func TestUpdateDataExportHandler_RepoUpdateError(t *testing.T) {
	t.Parallel()

	exportID := exportentity.NewDataExportID()
	existing := exportentity.ReconstructDataExport(
		exportID.UUID(), time.Now(), time.Now(),
		uuid.New(), "users", "csv", exportentity.ExportStatusPending, nil, nil,
	)
	repoErr := errors.New("update failed")

	repo := &mockWriteRepo{
		findByIDFn: func(_ context.Context, id exportentity.DataExportID) (*exportentity.DataExport, error) {
			if id == exportID {
				return existing, nil
			}
			return nil, exportentity.ErrDataExportNotFound
		},
		updateFn: func(_ context.Context, _ *exportentity.DataExport) error {
			return repoErr
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := command.NewUpdateDataExportHandler(repo, eb, l)

	status := exportentity.ExportStatusProcessing
	cmd := command.UpdateDataExportCommand{
		ID:     exportentity.DataExportID(exportID),
		Status: &status,
	}

	err := h.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

func TestUpdateDataExportHandler_NilStatus(t *testing.T) {
	t.Parallel()

	exportID := exportentity.NewDataExportID()
	existing := exportentity.ReconstructDataExport(
		exportID.UUID(), time.Now(), time.Now(),
		uuid.New(), "users", "csv", exportentity.ExportStatusPending, nil, nil,
	)

	repo := &mockWriteRepo{
		findByIDFn: func(_ context.Context, id exportentity.DataExportID) (*exportentity.DataExport, error) {
			if id == exportID {
				return existing, nil
			}
			return nil, exportentity.ErrDataExportNotFound
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := command.NewUpdateDataExportHandler(repo, eb, l)

	cmd := command.UpdateDataExportCommand{ID: exportentity.DataExportID(exportID)}

	err := h.Handle(context.Background(), cmd)
	require.NoError(t, err)
	if repo.updatedEntity.Status() != exportentity.ExportStatusPending {
		t.Fatalf("expected status unchanged (PENDING), got %s", repo.updatedEntity.Status())
	}
}

func TestUpdateDataExportHandler_EventBusError(t *testing.T) {
	t.Parallel()

	exportID := exportentity.NewDataExportID()
	existing := exportentity.ReconstructDataExport(
		exportID.UUID(), time.Now(), time.Now(),
		uuid.New(), "users", "csv", exportentity.ExportStatusPending, nil, nil,
	)

	repo := &mockWriteRepo{
		findByIDFn: func(_ context.Context, id exportentity.DataExportID) (*exportentity.DataExport, error) {
			if id == exportID {
				return existing, nil
			}
			return nil, exportentity.ErrDataExportNotFound
		},
	}
	eb := &mockEventBus{
		publishFn: func(_ context.Context, _ ...shared.DomainEvent) error {
			return errors.New("event bus down")
		},
	}
	l := &mockLogger{}
	h := command.NewUpdateDataExportHandler(repo, eb, l)

	status := exportentity.ExportStatusProcessing
	cmd := command.UpdateDataExportCommand{
		ID:     exportentity.DataExportID(exportID),
		Status: &status,
	}

	err := h.Handle(context.Background(), cmd)
	require.NoError(t, err)
	if repo.updatedEntity == nil {
		t.Fatal("entity should still be updated even if event bus fails")
	}
}
