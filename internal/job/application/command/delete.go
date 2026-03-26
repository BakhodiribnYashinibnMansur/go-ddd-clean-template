package command

import (
	"context"

	"gct/internal/job/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// DeleteJobCommand represents an intent to permanently remove a job record by its unique identifier.
// Deleting a running job does not cancel its in-flight execution — only the database record is removed.
type DeleteJobCommand struct {
	ID uuid.UUID
}

// DeleteJobHandler orchestrates job deletion through the repository layer.
// It enforces a hard-delete strategy — no soft-delete or audit trail is maintained at this level.
// Callers are responsible for ensuring the job is not currently running before deletion.
type DeleteJobHandler struct {
	repo   domain.JobRepository
	logger logger.Log
}

// NewDeleteJobHandler wires up the handler with its required dependencies.
func NewDeleteJobHandler(
	repo domain.JobRepository,
	logger logger.Log,
) *DeleteJobHandler {
	return &DeleteJobHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle performs the deletion of the job identified by cmd.ID.
// Returns nil on success; propagates repository errors (e.g., not found, connection failure) to the caller.
func (h *DeleteJobHandler) Handle(ctx context.Context, cmd DeleteJobCommand) error {
	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete job: %v", err)
		return err
	}
	return nil
}
