package command

import (
	"context"

	settingentity "gct/internal/context/iam/generic/usersetting/domain/entity"
	settingrepo "gct/internal/context/iam/generic/usersetting/domain/repository"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// DeleteUserSettingCommand holds the input for deleting a user setting.
type DeleteUserSettingCommand struct {
	ID settingentity.UserSettingID
}

// DeleteUserSettingHandler handles the DeleteUserSettingCommand.
type DeleteUserSettingHandler struct {
	repo   settingrepo.UserSettingRepository
	logger logger.Log
}

// NewDeleteUserSettingHandler creates a new DeleteUserSettingHandler.
func NewDeleteUserSettingHandler(
	repo settingrepo.UserSettingRepository,
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
	defer logger.SlowOp(h.logger, ctx, "DeleteUserSetting", "user_setting")()

	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorc(ctx, "repository delete failed", logger.F{Op: "DeleteUserSetting", Entity: "user_setting", EntityID: cmd.ID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}
	return nil
}
