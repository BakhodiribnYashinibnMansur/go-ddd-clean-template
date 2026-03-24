package query

import (
	"context"

	appdto "gct/internal/authz/application"
	"gct/internal/authz/domain"

	"github.com/google/uuid"
)

// GetRoleQuery holds the input for fetching a single role.
type GetRoleQuery struct {
	ID uuid.UUID
}

// GetRoleHandler handles the GetRoleQuery.
type GetRoleHandler struct {
	readRepo domain.AuthzReadRepository
}

// NewGetRoleHandler creates a new GetRoleHandler.
func NewGetRoleHandler(readRepo domain.AuthzReadRepository) *GetRoleHandler {
	return &GetRoleHandler{readRepo: readRepo}
}

// Handle executes the GetRoleQuery and returns a RoleView.
func (h *GetRoleHandler) Handle(ctx context.Context, q GetRoleQuery) (*appdto.RoleView, error) {
	view, err := h.readRepo.GetRole(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	return &appdto.RoleView{
		ID:          view.ID,
		Name:        view.Name,
		Description: view.Description,
	}, nil
}
