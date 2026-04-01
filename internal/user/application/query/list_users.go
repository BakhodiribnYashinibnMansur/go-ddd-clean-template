package query

import (
	"context"

	"gct/internal/shared/infrastructure/pgxutil"
	appdto "gct/internal/user/application"
	"gct/internal/user/domain"
)

// ListUsersQuery holds the input for listing users with filtering.
type ListUsersQuery struct {
	Filter domain.UsersFilter
}

// ListUsersResult holds the output of the list users query.
type ListUsersResult struct {
	Users []*appdto.UserView
	Total int64
}

// ListUsersHandler handles the ListUsersQuery.
type ListUsersHandler struct {
	readRepo domain.UserReadRepository
}

// NewListUsersHandler creates a new ListUsersHandler.
func NewListUsersHandler(readRepo domain.UserReadRepository) *ListUsersHandler {
	return &ListUsersHandler{readRepo: readRepo}
}

// Handle executes the ListUsersQuery and returns a list of UserView with total count.
func (h *ListUsersHandler) Handle(ctx context.Context, q ListUsersQuery) (_ *ListUsersResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListUsersHandler.Handle")
	defer func() { end(err) }()

	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		return nil, err
	}

	result := make([]*appdto.UserView, len(views))
	for i, v := range views {
		result[i] = &appdto.UserView{
			ID:         v.ID,
			Phone:      v.Phone,
			Email:      v.Email,
			Username:   v.Username,
			RoleID:     v.RoleID,
			Attributes: v.Attributes,
			Active:     v.Active,
			IsApproved: v.IsApproved,
		}
	}

	return &ListUsersResult{
		Users: result,
		Total: total,
	}, nil
}
