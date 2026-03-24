package query

import (
	"context"

	appdto "gct/internal/authz/application"
	"gct/internal/authz/domain"
	shared "gct/internal/shared/domain"
)

// ListScopesQuery holds the input for listing scopes.
type ListScopesQuery struct {
	Pagination shared.Pagination
}

// ListScopesResult holds the output of the list scopes query.
type ListScopesResult struct {
	Scopes []*appdto.ScopeView
	Total  int64
}

// ListScopesHandler handles the ListScopesQuery.
type ListScopesHandler struct {
	readRepo domain.AuthzReadRepository
}

// NewListScopesHandler creates a new ListScopesHandler.
func NewListScopesHandler(readRepo domain.AuthzReadRepository) *ListScopesHandler {
	return &ListScopesHandler{readRepo: readRepo}
}

// Handle executes the ListScopesQuery and returns a list of ScopeView.
func (h *ListScopesHandler) Handle(ctx context.Context, q ListScopesQuery) (*ListScopesResult, error) {
	views, total, err := h.readRepo.ListScopes(ctx, q.Pagination)
	if err != nil {
		return nil, err
	}

	result := make([]*appdto.ScopeView, len(views))
	for i, v := range views {
		result[i] = &appdto.ScopeView{
			Path:   v.Path,
			Method: v.Method,
		}
	}

	return &ListScopesResult{
		Scopes: result,
		Total:  total,
	}, nil
}
