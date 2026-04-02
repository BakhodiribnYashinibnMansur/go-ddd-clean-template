package query

import (
	"context"

	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/logger"

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
	logger   logger.Log
}

// NewListUsersHandler creates a new ListUsersHandler.
func NewListUsersHandler(readRepo domain.UserReadRepository, l logger.Log) *ListUsersHandler {
	return &ListUsersHandler{readRepo: readRepo, logger: l}
}

// Handle executes the ListUsersQuery and returns a list of UserView with total count.
func (h *ListUsersHandler) Handle(ctx context.Context, q ListUsersQuery) (_ *ListUsersResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListUsersHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ListUsers", "user")()

	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "ListUsers", Entity: "user", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
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
