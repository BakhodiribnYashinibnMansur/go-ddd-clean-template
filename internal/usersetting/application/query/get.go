package query

import (
	"context"

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
}

// NewGetUserSettingHandler creates a new GetUserSettingHandler.
func NewGetUserSettingHandler(readRepo domain.UserSettingReadRepository) *GetUserSettingHandler {
	return &GetUserSettingHandler{readRepo: readRepo}
}

// Handle executes the GetUserSettingQuery and returns a UserSettingView.
func (h *GetUserSettingHandler) Handle(ctx context.Context, q GetUserSettingQuery) (result *appdto.UserSettingView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetUserSettingHandler.Handle")
	defer func() { end(err) }()

	v, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
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
