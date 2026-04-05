package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/context/admin/dataexport/application/command"
	"gct/internal/context/admin/dataexport/domain"
	shared "gct/internal/platform/domain"

	"github.com/google/uuid"
)

func TestUpdateDataExportHandler_StatusProcessing(t *testing.T) {
	exportID := uuid.New()
	existing := domain.ReconstructDataExport(
		exportID, time.Now(), time.Now(),
		uuid.New(), "users", "csv", domain.ExportStatusPending, nil, nil,
	)

	repo := &mockWriteRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.DataExport, error) {
			if id == exportID {
				return existing, nil
			}
			return nil, domain.ErrDataExportNotFound
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := command.NewUpdateDataExportHandler(repo, eb, l)

	status := domain.ExportStatusProcessing
	cmd := command.UpdateDataExportCommand{
		ID:     exportID,
		Status: &status,
	}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.updatedEntity == nil {
		t.Fatal("expected entity to be updated")
	}
	if repo.updatedEntity.Status() != domain.ExportStatusProcessing {
		t.Fatalf("expected status PROCESSING, got %s", repo.updatedEntity.Status())
	}
}

func TestUpdateDataExportHandler_StatusCompleted(t *testing.T) {
	exportID := uuid.New()
	existing := domain.ReconstructDataExport(
		exportID, time.Now(), time.Now(),
		uuid.New(), "users", "csv", domain.ExportStatusProcessing, nil, nil,
	)

	repo := &mockWriteRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.DataExport, error) {
			if id == exportID {
				return existing, nil
			}
			return nil, domain.ErrDataExportNotFound
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := command.NewUpdateDataExportHandler(repo, eb, l)

	status := domain.ExportStatusCompleted
	fileURL := "https://example.com/export.csv"
	cmd := command.UpdateDataExportCommand{
		ID:      exportID,
		Status:  &status,
		FileURL: &fileURL,
	}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.updatedEntity.Status() != domain.ExportStatusCompleted {
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
	exportID := uuid.New()
	existing := domain.ReconstructDataExport(
		exportID, time.Now(), time.Now(),
		uuid.New(), "orders", "xlsx", domain.ExportStatusProcessing, nil, nil,
	)

	repo := &mockWriteRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.DataExport, error) {
			if id == exportID {
				return existing, nil
			}
			return nil, domain.ErrDataExportNotFound
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := command.NewUpdateDataExportHandler(repo, eb, l)

	status := domain.ExportStatusFailed
	errMsg := "disk full"
	cmd := command.UpdateDataExportCommand{
		ID:     exportID,
		Status: &status,
		Error:  &errMsg,
	}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.updatedEntity.Status() != domain.ExportStatusFailed {
		t.Fatalf("expected status FAILED, got %s", repo.updatedEntity.Status())
	}
	if repo.updatedEntity.Error() == nil || *repo.updatedEntity.Error() != errMsg {
		t.Fatalf("expected error msg %q, got %v", errMsg, repo.updatedEntity.Error())
	}
}

func TestUpdateDataExportHandler_NotFound(t *testing.T) {
	repo := &mockWriteRepo{}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := command.NewUpdateDataExportHandler(repo, eb, l)

	status := domain.ExportStatusProcessing
	cmd := command.UpdateDataExportCommand{
		ID:     uuid.New(),
		Status: &status,
	}

	err := h.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrDataExportNotFound) {
		t.Fatalf("expected ErrDataExportNotFound, got %v", err)
	}
}

func TestUpdateDataExportHandler_RepoUpdateError(t *testing.T) {
	exportID := uuid.New()
	existing := domain.ReconstructDataExport(
		exportID, time.Now(), time.Now(),
		uuid.New(), "users", "csv", domain.ExportStatusPending, nil, nil,
	)
	repoErr := errors.New("update failed")

	repo := &mockWriteRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.DataExport, error) {
			if id == exportID {
				return existing, nil
			}
			return nil, domain.ErrDataExportNotFound
		},
		updateFn: func(_ context.Context, _ *domain.DataExport) error {
			return repoErr
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := command.NewUpdateDataExportHandler(repo, eb, l)

	status := domain.ExportStatusProcessing
	cmd := command.UpdateDataExportCommand{
		ID:     exportID,
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
	exportID := uuid.New()
	existing := domain.ReconstructDataExport(
		exportID, time.Now(), time.Now(),
		uuid.New(), "users", "csv", domain.ExportStatusPending, nil, nil,
	)

	repo := &mockWriteRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.DataExport, error) {
			if id == exportID {
				return existing, nil
			}
			return nil, domain.ErrDataExportNotFound
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := command.NewUpdateDataExportHandler(repo, eb, l)

	cmd := command.UpdateDataExportCommand{ID: exportID}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.updatedEntity.Status() != domain.ExportStatusPending {
		t.Fatalf("expected status unchanged (PENDING), got %s", repo.updatedEntity.Status())
	}
}

func TestUpdateDataExportHandler_EventBusError(t *testing.T) {
	exportID := uuid.New()
	existing := domain.ReconstructDataExport(
		exportID, time.Now(), time.Now(),
		uuid.New(), "users", "csv", domain.ExportStatusPending, nil, nil,
	)

	repo := &mockWriteRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.DataExport, error) {
			if id == exportID {
				return existing, nil
			}
			return nil, domain.ErrDataExportNotFound
		},
	}
	eb := &mockEventBus{
		publishFn: func(_ context.Context, _ ...shared.DomainEvent) error {
			return errors.New("event bus down")
		},
	}
	l := &mockLogger{}
	h := command.NewUpdateDataExportHandler(repo, eb, l)

	status := domain.ExportStatusProcessing
	cmd := command.UpdateDataExportCommand{
		ID:     exportID,
		Status: &status,
	}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error (event bus failure is non-fatal), got %v", err)
	}
	if repo.updatedEntity == nil {
		t.Fatal("entity should still be updated even if event bus fails")
	}
}
