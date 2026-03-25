package command

import (
	"context"

	"gct/internal/iprule/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// DeleteIPRuleCommand holds the input for deleting an IP rule.
type DeleteIPRuleCommand struct {
	ID uuid.UUID
}

// DeleteIPRuleHandler handles the DeleteIPRuleCommand.
type DeleteIPRuleHandler struct {
	repo   domain.IPRuleRepository
	logger logger.Log
}

// NewDeleteIPRuleHandler creates a new DeleteIPRuleHandler.
func NewDeleteIPRuleHandler(
	repo domain.IPRuleRepository,
	logger logger.Log,
) *DeleteIPRuleHandler {
	return &DeleteIPRuleHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle executes the DeleteIPRuleCommand.
func (h *DeleteIPRuleHandler) Handle(ctx context.Context, cmd DeleteIPRuleCommand) error {
	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete ip rule: %v", err)
		return err
	}
	return nil
}
