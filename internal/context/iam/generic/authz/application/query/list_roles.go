package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	appdto "gct/internal/context/iam/generic/authz/application"
	"gct/internal/context/iam/generic/authz/domain"
	shared "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/pgxutil"
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
	logger   logger.Log
}

// NewListRolesHandler creates a new ListRolesHandler.
func NewListRolesHandler(readRepo domain.AuthzReadRepository, l logger.Log) *ListRolesHandler {
	return &ListRolesHandler{readRepo: readRepo, logger: l}
}

// Handle executes the ListRolesQuery and returns a list of RoleView.
func (h *ListRolesHandler) Handle(ctx context.Context, q ListRolesQuery) (_ *ListRolesResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListRolesHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ListRoles", "role")()

	views, total, err := h.readRepo.ListRoles(ctx, q.Pagination)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "ListRoles", Entity: "access", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
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
