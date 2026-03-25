package query

import (
	"context"

	appdto "gct/internal/integration/application"
	"gct/internal/integration/domain"
)

// ListQuery holds the input for listing integrations with filtering.
type ListQuery struct {
	Filter domain.IntegrationFilter
}

// ListResult holds the output of the list integrations query.
type ListResult struct {
	Integrations []*appdto.IntegrationView
	Total        int64
}

// ListHandler handles the ListQuery.
type ListHandler struct {
	readRepo domain.IntegrationReadRepository
}

// NewListHandler creates a new ListHandler.
func NewListHandler(readRepo domain.IntegrationReadRepository) *ListHandler {
	return &ListHandler{readRepo: readRepo}
}

// Handle executes the ListQuery and returns a list of IntegrationView with total count.
func (h *ListHandler) Handle(ctx context.Context, q ListQuery) (*ListResult, error) {
	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		return nil, err
	}

	result := make([]*appdto.IntegrationView, len(views))
	for i, v := range views {
		result[i] = &appdto.IntegrationView{
			ID:         v.ID,
			Name:       v.Name,
			Type:       v.Type,
			APIKey:     v.APIKey,
			WebhookURL: v.WebhookURL,
			Enabled:    v.Enabled,
			Config:     v.Config,
			CreatedAt:  v.CreatedAt,
			UpdatedAt:  v.UpdatedAt,
		}
	}

	return &ListResult{
		Integrations: result,
		Total:        total,
	}, nil
}
