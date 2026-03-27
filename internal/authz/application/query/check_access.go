package query

import (
	"context"

	"gct/internal/authz/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// CheckAccessQuery holds the input for checking whether a role has access to a specific endpoint.
type CheckAccessQuery struct {
	RoleID uuid.UUID
	Path   string
	Method string
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
func (h *CheckAccessHandler) Handle(ctx context.Context, q CheckAccessQuery) (bool, error) {
	allowed, err := h.readRepo.CheckAccess(ctx, q.RoleID, q.Path, q.Method)
	if err != nil {
		h.logger.Errorf("check access failed for role %s on %s %s: %v", q.RoleID, q.Method, q.Path, err)
		return false, err
	}
	return allowed, nil
}
