package command

import (
	"context"

	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"
	"gct/internal/usersetting/domain"

	"github.com/google/uuid"
)

// DeleteUserSettingCommand holds the input for deleting a user setting.
type DeleteUserSettingCommand struct {
	ID uuid.UUID
}

// DeleteUserSettingHandler handles the DeleteUserSettingCommand.
type DeleteUserSettingHandler struct {
	repo   domain.UserSettingRepository
	logger logger.Log
}

// NewDeleteUserSettingHandler creates a new DeleteUserSettingHandler.
func NewDeleteUserSettingHandler(
	repo domain.UserSettingRepository,
	logger logger.Log,
) *DeleteUserSettingHandler {
	return &DeleteUserSettingHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle executes the DeleteUserSettingCommand.
func (h *DeleteUserSettingHandler) Handle(ctx context.Context, cmd DeleteUserSettingCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "DeleteUserSettingHandler.Handle")
	defer func() { end(err) }()

	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete user setting: %v", err)
		return err
	}
	return nil
}
