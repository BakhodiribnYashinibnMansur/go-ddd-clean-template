package command

import (
	"context"

	"gct/internal/job/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// UpdateJobCommand represents a state transition for an existing job, typically issued by a worker.
// Status drives the domain state machine (pending -> running -> completed/failed).
// Result captures the output on success; Error captures the failure message on failure.
type UpdateJobCommand struct {
	ID     uuid.UUID
	Status *string
	Result map[string]any
	Error  *string
}

// UpdateJobHandler transitions a job through its lifecycle states and persists the outcome.
// State transitions are delegated to domain methods (Start, Complete, Fail) which enforce valid transitions.
type UpdateJobHandler struct {
	repo     domain.JobRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateJobHandler wires up the handler with its required dependencies.
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

// Handle fetches the job by ID, applies the requested state transition, and persists the updated entity.
// Unrecognized status values are silently ignored — only Running, Completed, and Failed trigger transitions.
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
