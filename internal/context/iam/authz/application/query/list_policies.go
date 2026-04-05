package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	appdto "gct/internal/context/iam/authz/application"
	"gct/internal/context/iam/authz/domain"
	shared "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// ListPoliciesQuery holds the input for listing policies.
type ListPoliciesQuery struct {
	Pagination shared.Pagination
}

// ListPoliciesResult holds the output of the list policies query.
type ListPoliciesResult struct {
	Policies []*appdto.PolicyView
	Total    int64
}

// ListPoliciesHandler handles the ListPoliciesQuery.
type ListPoliciesHandler struct {
	readRepo domain.AuthzReadRepository
	logger   logger.Log
}

// NewListPoliciesHandler creates a new ListPoliciesHandler.
func NewListPoliciesHandler(readRepo domain.AuthzReadRepository, l logger.Log) *ListPoliciesHandler {
	return &ListPoliciesHandler{readRepo: readRepo, logger: l}
}

// Handle executes the ListPoliciesQuery and returns a list of PolicyView.
func (h *ListPoliciesHandler) Handle(ctx context.Context, q ListPoliciesQuery) (_ *ListPoliciesResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListPoliciesHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ListPolicies", "policy")()

	views, total, err := h.readRepo.ListPolicies(ctx, q.Pagination)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "ListPolicies", Entity: "access", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	result := make([]*appdto.PolicyView, len(views))
	for i, v := range views {
		result[i] = &appdto.PolicyView{
			ID:           v.ID,
			PermissionID: v.PermissionID,
			Effect:       v.Effect,
			Priority:     v.Priority,
			Active:       v.Active,
			Conditions:   v.Conditions,
		}
	}

	return &ListPoliciesResult{
		Policies: result,
		Total:    total,
	}, nil
}
