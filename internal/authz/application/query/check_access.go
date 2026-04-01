package query

import (
	"context"

	"gct/internal/authz/domain"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// CheckAccessQuery holds the input for checking whether a role has access to a specific endpoint.
type CheckAccessQuery struct {
	RoleID  uuid.UUID
	Path    string
	Method  string
	EvalCtx domain.EvaluationContext
}

// CheckAccessHandler handles the CheckAccessQuery by delegating to the read repository.
type CheckAccessHandler struct {
	readRepo domain.AuthzReadRepository
	logger   logger.Log
}

// NewCheckAccessHandler creates a new CheckAccessHandler.
func NewCheckAccessHandler(readRepo domain.AuthzReadRepository, l logger.Log) *CheckAccessHandler {
	return &CheckAccessHandler{readRepo: readRepo, logger: l}
}

// Handle executes the CheckAccessQuery and returns true if the role has access.
func (h *CheckAccessHandler) Handle(ctx context.Context, q CheckAccessQuery) (allowed bool, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CheckAccessHandler.Handle")
	defer func() { end(err) }()

	allowed, err = h.readRepo.CheckAccess(ctx, q.RoleID, q.Path, q.Method, q.EvalCtx)
	if err != nil {
		h.logger.Errorf("check access failed for role %s on %s %s: %v", q.RoleID, q.Method, q.Path, err)
		return false, err
	}
	return allowed, nil
}
