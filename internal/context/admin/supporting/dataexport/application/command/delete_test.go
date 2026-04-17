package command_test

import (
	"context"
	"errors"
	"testing"

	"gct/internal/context/admin/supporting/dataexport/application/command"
	exportentity "gct/internal/context/admin/supporting/dataexport/domain/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDeleteDataExportHandler_Success(t *testing.T) {
	t.Parallel()

	repo := &mockWriteRepo{}
	l := &mockLogger{}
	h := command.NewDeleteDataExportHandler(repo, nil, l)

	exportID := exportentity.NewDataExportID()
	cmd := command.DeleteDataExportCommand{ID: exportentity.DataExportID(exportID)}

	err := h.Handle(context.Background(), cmd)
	require.NoError(t, err)
	if repo.deletedID != exportID {
		t.Fatalf("expected deleted ID %s, got %s", exportID, repo.deletedID)
	}
}

func TestDeleteDataExportHandler_RepoError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("delete failed")
	repo := &mockWriteRepo{
		deleteFn: func(_ context.Context, _ exportentity.DataExportID) error {
			return repoErr
		},
	}
	l := &mockLogger{}
	h := command.NewDeleteDataExportHandler(repo, nil, l)

	cmd := command.DeleteDataExportCommand{ID: exportentity.DataExportID(uuid.New())}

	err := h.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
