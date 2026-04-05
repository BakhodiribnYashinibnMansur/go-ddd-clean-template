package command_test

import (
	"context"
	"errors"
	"testing"

	"gct/internal/context/admin/dataexport/application/command"
	"gct/internal/context/admin/dataexport/domain"
	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateDataExportHandler_Success(t *testing.T) {
	t.Parallel()

	repo := &mockWriteRepo{}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := command.NewCreateDataExportHandler(repo, eb, l)

	cmd := command.CreateDataExportCommand{
		UserID:   uuid.New(),
		DataType: "users",
		Format:   "csv",
	}

	err := h.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.savedEntity == nil {
		t.Fatal("expected entity to be saved")
	}
	if repo.savedEntity.UserID() != cmd.UserID {
		t.Fatalf("expected user ID %s, got %s", cmd.UserID, repo.savedEntity.UserID())
	}
	if repo.savedEntity.DataType() != "users" {
		t.Fatalf("expected data type users, got %s", repo.savedEntity.DataType())
	}
	if repo.savedEntity.Format() != "csv" {
		t.Fatalf("expected format csv, got %s", repo.savedEntity.Format())
	}
	if repo.savedEntity.Status() != domain.ExportStatusPending {
		t.Fatalf("expected status PENDING, got %s", repo.savedEntity.Status())
	}
	if len(eb.publishedEvents) != 1 {
		t.Fatalf("expected 1 event published, got %d", len(eb.publishedEvents))
	}
	if eb.publishedEvents[0].EventName() != "dataexport.requested" {
		t.Fatalf("expected dataexport.requested event, got %s", eb.publishedEvents[0].EventName())
	}
}

func TestCreateDataExportHandler_RepoSaveError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("db connection failed")
	repo := &mockWriteRepo{
		saveFn: func(_ context.Context, _ *domain.DataExport) error {
			return repoErr
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := command.NewCreateDataExportHandler(repo, eb, l)

	cmd := command.CreateDataExportCommand{
		UserID:   uuid.New(),
		DataType: "orders",
		Format:   "xlsx",
	}

	err := h.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
	if len(eb.publishedEvents) != 0 {
		t.Fatalf("expected 0 events published on save error, got %d", len(eb.publishedEvents))
	}
}

func TestCreateDataExportHandler_EventBusError(t *testing.T) {
	t.Parallel()

	repo := &mockWriteRepo{}
	eb := &mockEventBus{
		publishFn: func(_ context.Context, _ ...shared.DomainEvent) error {
			return errors.New("event bus down")
		},
	}
	l := &mockLogger{}
	h := command.NewCreateDataExportHandler(repo, eb, l)

	cmd := command.CreateDataExportCommand{
		UserID:   uuid.New(),
		DataType: "logs",
		Format:   "json",
	}

	// Event bus failure is non-fatal — handler returns nil
	err := h.Handle(context.Background(), cmd)
	require.NoError(t, err)
	if repo.savedEntity == nil {
		t.Fatal("entity should still be saved even if event bus fails")
	}
}
