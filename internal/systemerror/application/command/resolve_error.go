package command

import (
	"context"

	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/systemerror/domain"

	"github.com/google/uuid"
)

// ResolveErrorCommand holds the input for resolving a system error.
type ResolveErrorCommand struct {
	ID         uuid.UUID
	ResolvedBy uuid.UUID
}

// ResolveErrorHandler handles the ResolveErrorCommand.
type ResolveErrorHandler struct {
	repo     domain.SystemErrorRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewResolveErrorHandler creates a new ResolveErrorHandler.
func NewResolveErrorHandler(
	repo domain.SystemErrorRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *ResolveErrorHandler {
	return &ResolveErrorHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the ResolveErrorCommand.
func (h *ResolveErrorHandler) Handle(ctx context.Context, cmd ResolveErrorCommand) error {
	se, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	se.Resolve(cmd.ResolvedBy)

	if err := h.repo.Update(ctx, se); err != nil {
		h.logger.Errorf("failed to update system error: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, se.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
