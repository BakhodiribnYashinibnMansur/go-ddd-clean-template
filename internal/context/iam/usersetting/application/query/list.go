package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	appdto "gct/internal/context/iam/usersetting/application"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/context/iam/usersetting/domain"
)

// ListUserSettingsQuery holds the input for listing user settings with filtering.
type ListUserSettingsQuery struct {
	Filter domain.UserSettingFilter
}

// ListUserSettingsResult holds the output of the list user settings query.
type ListUserSettingsResult struct {
	Settings []*appdto.UserSettingView
	Total    int64
}

// ListUserSettingsHandler handles the ListUserSettingsQuery.
type ListUserSettingsHandler struct {
	readRepo domain.UserSettingReadRepository
	logger   logger.Log
}

// NewListUserSettingsHandler creates a new ListUserSettingsHandler.
func NewListUserSettingsHandler(readRepo domain.UserSettingReadRepository, l logger.Log) *ListUserSettingsHandler {
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

	result := make([]*appdto.UserSettingView, len(views))
	for i, v := range views {
		result[i] = &appdto.UserSettingView{
			ID:        v.ID,
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
