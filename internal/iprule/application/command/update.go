package command

import (
	"context"
	"time"

	"gct/internal/iprule/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// UpdateIPRuleCommand holds the input for updating an IP rule.
type UpdateIPRuleCommand struct {
	ID        uuid.UUID
	IPAddress *string
	Action    *string
	Reason    *string
	ExpiresAt *time.Time
}

// UpdateIPRuleHandler handles the UpdateIPRuleCommand.
type UpdateIPRuleHandler struct {
	repo     domain.IPRuleRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateIPRuleHandler creates a new UpdateIPRuleHandler.
func NewUpdateIPRuleHandler(
	repo domain.IPRuleRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *UpdateIPRuleHandler {
	return &UpdateIPRuleHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the UpdateIPRuleCommand.
func (h *UpdateIPRuleHandler) Handle(ctx context.Context, cmd UpdateIPRuleCommand) error {
	r, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	r.Update(cmd.IPAddress, cmd.Action, cmd.Reason, cmd.ExpiresAt)

	if err := h.repo.Update(ctx, r); err != nil {
		h.logger.Errorf("failed to update ip rule: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, r.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
