package query

import (
	"context"

	appdto "gct/internal/authz/application"
	"gct/internal/authz/domain"
	shared "gct/internal/shared/domain"
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
}

// NewListPoliciesHandler creates a new ListPoliciesHandler.
func NewListPoliciesHandler(readRepo domain.AuthzReadRepository) *ListPoliciesHandler {
	return &ListPoliciesHandler{readRepo: readRepo}
}

// Handle executes the ListPoliciesQuery and returns a list of PolicyView.
func (h *ListPoliciesHandler) Handle(ctx context.Context, q ListPoliciesQuery) (*ListPoliciesResult, error) {
	views, total, err := h.readRepo.ListPolicies(ctx, q.Pagination)
	if err != nil {
		return nil, err
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
