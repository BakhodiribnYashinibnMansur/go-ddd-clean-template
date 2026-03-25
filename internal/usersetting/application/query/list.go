package query

import (
	"context"

	appdto "gct/internal/usersetting/application"
	"gct/internal/usersetting/domain"
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
}

// NewListUserSettingsHandler creates a new ListUserSettingsHandler.
func NewListUserSettingsHandler(readRepo domain.UserSettingReadRepository) *ListUserSettingsHandler {
	return &ListUserSettingsHandler{readRepo: readRepo}
}

// Handle executes the ListUserSettingsQuery and returns a list of UserSettingView with total count.
func (h *ListUserSettingsHandler) Handle(ctx context.Context, q ListUserSettingsQuery) (*ListUserSettingsResult, error) {
	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		return nil, err
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
