package command

import (
	"context"

	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/usersetting/domain"

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
func (h *UpsertUserSettingHandler) Handle(ctx context.Context, cmd UpsertUserSettingCommand) error {
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
		h.logger.Errorf("failed to upsert user setting: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, us.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
