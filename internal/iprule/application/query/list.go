package query

import (
	"context"

	appdto "gct/internal/iprule/application"
	"gct/internal/iprule/domain"
)

// ListIPRulesQuery holds the input for listing IP rules.
type ListIPRulesQuery struct {
	Filter domain.IPRuleFilter
}

// ListIPRulesResult holds the output of the list IP rules query.
type ListIPRulesResult struct {
	IPRules []*appdto.IPRuleView
	Total   int64
}

// ListIPRulesHandler handles the ListIPRulesQuery.
type ListIPRulesHandler struct {
	readRepo domain.IPRuleReadRepository
}

// NewListIPRulesHandler creates a new ListIPRulesHandler.
func NewListIPRulesHandler(readRepo domain.IPRuleReadRepository) *ListIPRulesHandler {
	return &ListIPRulesHandler{readRepo: readRepo}
}

// Handle executes the ListIPRulesQuery and returns a list of IPRuleView with total count.
func (h *ListIPRulesHandler) Handle(ctx context.Context, q ListIPRulesQuery) (*ListIPRulesResult, error) {
	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		return nil, err
	}

	result := make([]*appdto.IPRuleView, len(views))
	for i, v := range views {
		result[i] = &appdto.IPRuleView{
			ID:        v.ID,
			IPAddress: v.IPAddress,
			Action:    v.Action,
			Reason:    v.Reason,
			ExpiresAt: v.ExpiresAt,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
		}
	}

	return &ListIPRulesResult{
		IPRules: result,
		Total:   total,
	}, nil
}
