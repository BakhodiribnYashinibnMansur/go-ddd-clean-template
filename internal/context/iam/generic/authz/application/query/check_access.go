package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"

	authzentity "gct/internal/context/iam/generic/authz/domain/entity"
	authzrepo "gct/internal/context/iam/generic/authz/domain/repository"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// CheckAccessQuery holds the input for checking whether a role has access to a specific endpoint.
type CheckAccessQuery struct {
	RoleID  authzentity.RoleID
	Path    string
	Method  string
	EvalCtx authzentity.EvaluationContext
}

// CheckAccessHandler handles the CheckAccessQuery by delegating to the read repository.
type CheckAccessHandler struct {
	readRepo authzrepo.AuthzReadRepository
	logger   logger.Log
}

// NewCheckAccessHandler creates a new CheckAccessHandler.
func NewCheckAccessHandler(readRepo authzrepo.AuthzReadRepository, l logger.Log) *CheckAccessHandler {
	return &CheckAccessHandler{readRepo: readRepo, logger: l}
}

// Handle executes the CheckAccessQuery and returns true if the role has access.
func (h *CheckAccessHandler) Handle(ctx context.Context, q CheckAccessQuery) (allowed bool, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CheckAccessHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "CheckAccess", "access")()

	allowed, err = h.readRepo.CheckAccess(ctx, q.RoleID, q.Path, q.Method, q.EvalCtx)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "CheckAccess", Entity: "access", EntityID: q.RoleID, Err: err}.KV()...)
		return false, apperrors.MapToServiceError(err)
	}
	return allowed, nil
}
