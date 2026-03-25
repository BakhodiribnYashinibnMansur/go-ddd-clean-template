package command

import (
	"context"

	"gct/internal/job/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// UpdateJobCommand holds the input for updating an existing job.
type UpdateJobCommand struct {
	ID     uuid.UUID
	Status *string
	Result map[string]any
	Error  *string
}

// UpdateJobHandler handles the UpdateJobCommand.
type UpdateJobHandler struct {
	repo     domain.JobRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateJobHandler creates a new UpdateJobHandler.
func NewUpdateJobHandler(
	repo domain.JobRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *UpdateJobHandler {
	return &UpdateJobHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the UpdateJobCommand.
func (h *UpdateJobHandler) Handle(ctx context.Context, cmd UpdateJobCommand) error {
	j, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if cmd.Status != nil {
		switch *cmd.Status {
		case domain.JobStatusRunning:
			j.Start()
		case domain.JobStatusCompleted:
			j.Complete(cmd.Result)
		case domain.JobStatusFailed:
			errMsg := ""
			if cmd.Error != nil {
				errMsg = *cmd.Error
			}
			j.Fail(errMsg)
		}
	}

	if err := h.repo.Update(ctx, j); err != nil {
		h.logger.Errorf("failed to update job: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, j.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
