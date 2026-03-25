package command

import (
	"context"
	"time"

	"gct/internal/iprule/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
)

// CreateIPRuleCommand holds the input for creating a new IP rule.
type CreateIPRuleCommand struct {
	IPAddress string
	Action    string
	Reason    string
	ExpiresAt *time.Time
}

// CreateIPRuleHandler handles the CreateIPRuleCommand.
type CreateIPRuleHandler struct {
	repo     domain.IPRuleRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateIPRuleHandler creates a new CreateIPRuleHandler.
func NewCreateIPRuleHandler(
	repo domain.IPRuleRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *CreateIPRuleHandler {
	return &CreateIPRuleHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the CreateIPRuleCommand.
func (h *CreateIPRuleHandler) Handle(ctx context.Context, cmd CreateIPRuleCommand) error {
	r := domain.NewIPRule(cmd.IPAddress, cmd.Action, cmd.Reason, cmd.ExpiresAt)

	if err := h.repo.Save(ctx, r); err != nil {
		h.logger.Errorf("failed to save ip rule: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, r.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
