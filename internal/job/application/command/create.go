package command

import (
	"context"
	"time"

	"gct/internal/job/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
)

// CreateJobCommand holds the input for creating a new job.
type CreateJobCommand struct {
	TaskName    string
	Payload     map[string]any
	MaxAttempts int
	ScheduledAt *time.Time
}

// CreateJobHandler handles the CreateJobCommand.
type CreateJobHandler struct {
	repo     domain.JobRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateJobHandler creates a new CreateJobHandler.
func NewCreateJobHandler(
	repo domain.JobRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *CreateJobHandler {
	return &CreateJobHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the CreateJobCommand.
func (h *CreateJobHandler) Handle(ctx context.Context, cmd CreateJobCommand) error {
	j := domain.NewJob(cmd.TaskName, cmd.Payload, cmd.MaxAttempts, cmd.ScheduledAt)

	if err := h.repo.Save(ctx, j); err != nil {
		h.logger.Errorf("failed to save job: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, j.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
