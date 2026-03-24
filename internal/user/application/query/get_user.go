package query

import (
	"context"

	appdto "gct/internal/user/application"
	"gct/internal/user/domain"

	"github.com/google/uuid"
)

// GetUserQuery holds the input for fetching a single user.
type GetUserQuery struct {
	ID uuid.UUID
}

// GetUserHandler handles the GetUserQuery.
type GetUserHandler struct {
	readRepo domain.UserReadRepository
}

// NewGetUserHandler creates a new GetUserHandler.
func NewGetUserHandler(readRepo domain.UserReadRepository) *GetUserHandler {
	return &GetUserHandler{readRepo: readRepo}
}

// Handle executes the GetUserQuery and returns a UserView.
func (h *GetUserHandler) Handle(ctx context.Context, q GetUserQuery) (*appdto.UserView, error) {
	view, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	// Map domain UserView to application UserView.
	return &appdto.UserView{
		ID:         view.ID,
		Phone:      view.Phone,
		Email:      view.Email,
		Username:   view.Username,
		RoleID:     view.RoleID,
		Attributes: view.Attributes,
		Active:     view.Active,
		IsApproved: view.IsApproved,
	}, nil
}
