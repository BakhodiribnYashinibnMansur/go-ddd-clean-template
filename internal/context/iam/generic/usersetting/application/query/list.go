package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	"gct/internal/context/iam/generic/usersetting/application/dto"
	settingrepo "gct/internal/context/iam/generic/usersetting/domain/repository"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// ListUserSettingsQuery holds the input for listing user settings with filtering.
type ListUserSettingsQuery struct {
	Filter settingrepo.UserSettingFilter
}

// ListUserSettingsResult holds the output of the list user settings query.
type ListUserSettingsResult struct {
	Settings []*dto.UserSettingView
	Total    int64
}

// ListUserSettingsHandler handles the ListUserSettingsQuery.
type ListUserSettingsHandler struct {
	readRepo settingrepo.UserSettingReadRepository
	logger   logger.Log
}

// NewListUserSettingsHandler creates a new ListUserSettingsHandler.
func NewListUserSettingsHandler(readRepo settingrepo.UserSettingReadRepository, l logger.Log) *ListUserSettingsHandler {
	return &ListUserSettingsHandler{readRepo: readRepo, logger: l}
}

// Handle executes the ListUserSettingsQuery and returns a list of UserSettingView with total count.
func (h *ListUserSettingsHandler) Handle(ctx context.Context, q ListUserSettingsQuery) (_ *ListUserSettingsResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListUserSettingsHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ListUserSettings", "user_setting")()

	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "ListUserSettings", Entity: "user_setting", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	result := make([]*dto.UserSettingView, len(views))
	for i, v := range views {
		result[i] = &dto.UserSettingView{
			ID:        uuid.UUID(v.ID),
			UserID:    v.UserID,
			Key:       v.Key,
			Value:     v.Value,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
		}
	}

	return &ListUserSettingsResult{
		Settings: result,
		Total:    total,
	}, nil
}
