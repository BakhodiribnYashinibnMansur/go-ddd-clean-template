package command

import (
	"context"

	"gct/internal/job/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// DeleteJobCommand holds the input for deleting a job.
type DeleteJobCommand struct {
	ID uuid.UUID
}

// DeleteJobHandler handles the DeleteJobCommand.
type DeleteJobHandler struct {
	repo   domain.JobRepository
	logger logger.Log
}

// NewDeleteJobHandler creates a new DeleteJobHandler.
func NewDeleteJobHandler(
	repo domain.JobRepository,
	logger logger.Log,
) *DeleteJobHandler {
	return &DeleteJobHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle executes the DeleteJobCommand.
func (h *DeleteJobHandler) Handle(ctx context.Context, cmd DeleteJobCommand) error {
	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete job: %v", err)
		return err
	}
	return nil
}
