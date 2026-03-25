package command

import (
	"context"

	"gct/internal/ratelimit/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
)

// CreateRateLimitCommand holds the input for creating a new rate limit.
type CreateRateLimitCommand struct {
	Name              string
	Rule              string
	RequestsPerWindow int
	WindowDuration    int
	Enabled           bool
}

// CreateRateLimitHandler handles the CreateRateLimitCommand.
type CreateRateLimitHandler struct {
	repo     domain.RateLimitRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateRateLimitHandler creates a new CreateRateLimitHandler.
func NewCreateRateLimitHandler(
	repo domain.RateLimitRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *CreateRateLimitHandler {
	return &CreateRateLimitHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the CreateRateLimitCommand.
func (h *CreateRateLimitHandler) Handle(ctx context.Context, cmd CreateRateLimitCommand) error {
	rl := domain.NewRateLimit(cmd.Name, cmd.Rule, cmd.RequestsPerWindow, cmd.WindowDuration, cmd.Enabled)

	if err := h.repo.Save(ctx, rl); err != nil {
		h.logger.Errorf("failed to save rate limit: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, rl.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
