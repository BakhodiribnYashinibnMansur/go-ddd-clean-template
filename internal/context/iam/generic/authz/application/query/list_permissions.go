package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	"gct/internal/context/iam/generic/authz/application/dto"
	authzrepo "gct/internal/context/iam/generic/authz/domain/repository"
	shared "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// ListPermissionsQuery holds the input for listing permissions.
type ListPermissionsQuery struct {
	Pagination shared.Pagination
}

// ListPermissionsResult holds the output of the list permissions query.
type ListPermissionsResult struct {
	Permissions []*dto.PermissionView
	Total       int64
}

// ListPermissionsHandler handles the ListPermissionsQuery.
type ListPermissionsHandler struct {
	readRepo authzrepo.AuthzReadRepository
	logger   logger.Log
}

// NewListPermissionsHandler creates a new ListPermissionsHandler.
func NewListPermissionsHandler(readRepo authzrepo.AuthzReadRepository, l logger.Log) *ListPermissionsHandler {
	return &ListPermissionsHandler{readRepo: readRepo, logger: l}
}

// Handle executes the ListPermissionsQuery and returns a list of PermissionView.
func (h *ListPermissionsHandler) Handle(ctx context.Context, q ListPermissionsQuery) (_ *ListPermissionsResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListPermissionsHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ListPermissions", "permission")()

	views, total, err := h.readRepo.ListPermissions(ctx, q.Pagination)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "ListPermissions", Entity: "access", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	result := make([]*dto.PermissionView, len(views))
	for i, v := range views {
		var parentID *uuid.UUID
		if v.ParentID != nil {
			id := uuid.UUID(*v.ParentID)
			parentID = &id
		}
		result[i] = &dto.PermissionView{
			ID:          uuid.UUID(v.ID),
			ParentID:    parentID,
			Name:        v.Name,
			Description: v.Description,
		}
	}

	return &ListPermissionsResult{
		Permissions: result,
		Total:       total,
	}, nil
}
