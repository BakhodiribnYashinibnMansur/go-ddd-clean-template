package query

import (
	"context"

	appdto "gct/internal/authz/application"
	"gct/internal/authz/domain"
	shared "gct/internal/shared/domain"
)

// ListPermissionsQuery holds the input for listing permissions.
type ListPermissionsQuery struct {
	Pagination shared.Pagination
}

// ListPermissionsResult holds the output of the list permissions query.
type ListPermissionsResult struct {
	Permissions []*appdto.PermissionView
	Total       int64
}

// ListPermissionsHandler handles the ListPermissionsQuery.
type ListPermissionsHandler struct {
	readRepo domain.AuthzReadRepository
}

// NewListPermissionsHandler creates a new ListPermissionsHandler.
func NewListPermissionsHandler(readRepo domain.AuthzReadRepository) *ListPermissionsHandler {
	return &ListPermissionsHandler{readRepo: readRepo}
}

// Handle executes the ListPermissionsQuery and returns a list of PermissionView.
func (h *ListPermissionsHandler) Handle(ctx context.Context, q ListPermissionsQuery) (*ListPermissionsResult, error) {
	views, total, err := h.readRepo.ListPermissions(ctx, q.Pagination)
	if err != nil {
		return nil, err
	}

	result := make([]*appdto.PermissionView, len(views))
	for i, v := range views {
		result[i] = &appdto.PermissionView{
			ID:          v.ID,
			ParentID:    v.ParentID,
			Name:        v.Name,
			Description: v.Description,
		}
	}

	return &ListPermissionsResult{
		Permissions: result,
		Total:       total,
	}, nil
}
