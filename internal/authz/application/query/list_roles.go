package query

import (
	"context"

	appdto "gct/internal/authz/application"
	"gct/internal/authz/domain"
	shared "gct/internal/shared/domain"
	"gct/internal/shared/infrastructure/pgxutil"
)

// ListRolesQuery holds the input for listing roles.
type ListRolesQuery struct {
	Pagination shared.Pagination
}

// ListRolesResult holds the output of the list roles query.
type ListRolesResult struct {
	Roles []*appdto.RoleView
	Total int64
}

// ListRolesHandler handles the ListRolesQuery.
type ListRolesHandler struct {
	readRepo domain.AuthzReadRepository
}

// NewListRolesHandler creates a new ListRolesHandler.
func NewListRolesHandler(readRepo domain.AuthzReadRepository) *ListRolesHandler {
	return &ListRolesHandler{readRepo: readRepo}
}

// Handle executes the ListRolesQuery and returns a list of RoleView.
func (h *ListRolesHandler) Handle(ctx context.Context, q ListRolesQuery) (_ *ListRolesResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListRolesHandler.Handle")
	defer func() { end(err) }()

	views, total, err := h.readRepo.ListRoles(ctx, q.Pagination)
	if err != nil {
		return nil, err
	}

	result := make([]*appdto.RoleView, len(views))
	for i, v := range views {
		result[i] = &appdto.RoleView{
			ID:          v.ID,
			Name:        v.Name,
			Description: v.Description,
		}
	}

	return &ListRolesResult{
		Roles: result,
		Total: total,
	}, nil
}
