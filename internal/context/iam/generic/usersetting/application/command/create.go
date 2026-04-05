package command

import (
	"context"

	"gct/internal/context/iam/generic/usersetting/domain"
	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// UpsertUserSettingCommand holds the input for creating or updating a user setting.
type UpsertUserSettingCommand struct {
	UserID uuid.UUID
	Key    string
	Value  string
}

// UpsertUserSettingHandler handles the UpsertUserSettingCommand.
type UpsertUserSettingHandler struct {
	repo     domain.UserSettingRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpsertUserSettingHandler creates a new UpsertUserSettingHandler.
func NewUpsertUserSettingHandler(
	repo domain.UserSettingRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *UpsertUserSettingHandler {
	return &UpsertUserSettingHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the UpsertUserSettingCommand.
func (h *UpsertUserSettingHandler) Handle(ctx context.Context, cmd UpsertUserSettingCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "UpsertUserSettingHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "CreateUserSetting", "user_setting")()

	// Try to find existing setting by user+key.
	existing, _ := h.repo.FindByUserIDAndKey(ctx, cmd.UserID, cmd.Key)

	var us *domain.UserSetting
	if existing != nil {
		existing.ChangeValue(cmd.Value)
		us = existing
	} else {
		us = domain.NewUserSetting(cmd.UserID, cmd.Key, cmd.Value)
	}

	if err := h.repo.Upsert(ctx, us); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateUserSetting", Entity: "user_setting", Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, us.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "CreateUserSetting", Entity: "user_setting", Err: err}.KV()...)
	}

	return nil
}
