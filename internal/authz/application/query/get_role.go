package query

import (
	"context"

	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/logger"

	appdto "gct/internal/authz/application"
	"gct/internal/authz/domain"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// GetRoleQuery holds the input for fetching a single role.
type GetRoleQuery struct {
	ID uuid.UUID
}

// GetRoleHandler handles the GetRoleQuery.
type GetRoleHandler struct {
	readRepo domain.AuthzReadRepository
	logger   logger.Log
}

// NewGetRoleHandler creates a new GetRoleHandler.
func NewGetRoleHandler(readRepo domain.AuthzReadRepository, l logger.Log) *GetRoleHandler {
	return &GetRoleHandler{readRepo: readRepo, logger: l}
}

// Handle executes the GetRoleQuery and returns a RoleView.
func (h *GetRoleHandler) Handle(ctx context.Context, q GetRoleQuery) (_ *appdto.RoleView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetRoleHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "GetRole", "role")()

	view, err := h.readRepo.GetRole(ctx, q.ID)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "GetRole", Entity: "access", EntityID: q.ID, Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	return &appdto.RoleView{
		ID:          view.ID,
		Name:        view.Name,
		Description: view.Description,
	}, nil
}
