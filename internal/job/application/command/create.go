package command

import (
	"context"
	"time"

	"gct/internal/job/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
)

// CreateJobCommand represents an intent to enqueue a new background job for deferred execution.
// Payload carries task-specific arguments as schemaless JSON; its structure is defined by the TaskName consumer.
// ScheduledAt is optional — nil means "execute as soon as a worker picks it up."
type CreateJobCommand struct {
	TaskName    string
	Payload     map[string]any
	MaxAttempts int
	ScheduledAt *time.Time
}

// CreateJobHandler persists a new job record and publishes domain events for worker pickup.
// The job starts in a "pending" state; actual execution is handled by a separate worker process.
type CreateJobHandler struct {
	repo     domain.JobRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateJobHandler wires up the handler with its required dependencies.
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

// Handle creates the job domain entity, persists it, and publishes domain events (e.g., JobCreated).
// Event publish failures are logged but do not fail the operation — the job is already saved.
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
