package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	"gct/internal/context/iam/generic/authz/application/dto"
	authzentity "gct/internal/context/iam/generic/authz/domain/entity"
	authzrepo "gct/internal/context/iam/generic/authz/domain/repository"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// GetRoleQuery holds the input for fetching a single role.
type GetRoleQuery struct {
	ID authzentity.RoleID
}

// GetRoleHandler handles the GetRoleQuery.
type GetRoleHandler struct {
	readRepo authzrepo.AuthzReadRepository
	logger   logger.Log
}

// NewGetRoleHandler creates a new GetRoleHandler.
func NewGetRoleHandler(readRepo authzrepo.AuthzReadRepository, l logger.Log) *GetRoleHandler {
	return &GetRoleHandler{readRepo: readRepo, logger: l}
}

// Handle executes the GetRoleQuery and returns a RoleView.
func (h *GetRoleHandler) Handle(ctx context.Context, q GetRoleQuery) (_ *dto.RoleView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetRoleHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "GetRole", "role")()

	view, err := h.readRepo.GetRole(ctx, q.ID)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "GetRole", Entity: "access", EntityID: q.ID, Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	return &dto.RoleView{
		ID:          uuid.UUID(view.ID),
		Name:        view.Name,
		Description: view.Description,
	}, nil
}
