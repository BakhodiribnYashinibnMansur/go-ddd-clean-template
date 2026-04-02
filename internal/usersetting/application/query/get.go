package query

import (
	"context"

	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/logger"

	appdto "gct/internal/usersetting/application"
	"gct/internal/shared/infrastructure/pgxutil"
	"gct/internal/usersetting/domain"

	"github.com/google/uuid"
)

// GetUserSettingQuery holds the input for getting a single user setting.
type GetUserSettingQuery struct {
	ID uuid.UUID
}

// GetUserSettingHandler handles the GetUserSettingQuery.
type GetUserSettingHandler struct {
	readRepo domain.UserSettingReadRepository
	logger   logger.Log
}

// NewGetUserSettingHandler creates a new GetUserSettingHandler.
func NewGetUserSettingHandler(readRepo domain.UserSettingReadRepository, l logger.Log) *GetUserSettingHandler {
	return &GetUserSettingHandler{readRepo: readRepo, logger: l}
}

// Handle executes the GetUserSettingQuery and returns a UserSettingView.
func (h *GetUserSettingHandler) Handle(ctx context.Context, q GetUserSettingQuery) (result *appdto.UserSettingView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetUserSettingHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "GetUserSetting", "user_setting")()

	v, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "GetUserSetting", Entity: "user_setting", EntityID: q.ID, Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	return &appdto.UserSettingView{
		ID:        v.ID,
		UserID:    v.UserID,
		Key:       v.Key,
		Value:     v.Value,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}, nil
}
