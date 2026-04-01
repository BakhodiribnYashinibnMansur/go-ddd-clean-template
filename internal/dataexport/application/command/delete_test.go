package command_test

import (
	"context"
	"errors"
	"testing"

	"gct/internal/dataexport/application/command"

	"github.com/google/uuid"
)

func TestDeleteDataExportHandler_Success(t *testing.T) {
	repo := &mockWriteRepo{}
	l := &mockLogger{}
	h := command.NewDeleteDataExportHandler(repo, l)

	exportID := uuid.New()
	cmd := command.DeleteDataExportCommand{ID: exportID}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.deletedID != exportID {
		t.Fatalf("expected deleted ID %s, got %s", exportID, repo.deletedID)
	}
}

func TestDeleteDataExportHandler_RepoError(t *testing.T) {
	repoErr := errors.New("delete failed")
	repo := &mockWriteRepo{
		deleteFn: func(_ context.Context, _ uuid.UUID) error {
			return repoErr
		},
	}
	l := &mockLogger{}
	h := command.NewDeleteDataExportHandler(repo, l)

	cmd := command.DeleteDataExportCommand{ID: uuid.New()}

	err := h.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
