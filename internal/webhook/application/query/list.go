package query

import (
	"context"

	appdto "gct/internal/webhook/application"
	"gct/internal/webhook/domain"
)

// ListQuery holds the input for listing webhooks with filtering.
type ListQuery struct {
	Filter domain.WebhookFilter
}

// ListResult holds the output of the list webhooks query.
type ListResult struct {
	Webhooks []*appdto.WebhookView
	Total    int64
}

// ListHandler handles the ListQuery.
type ListHandler struct {
	readRepo domain.WebhookReadRepository
}

// NewListHandler creates a new ListHandler.
func NewListHandler(readRepo domain.WebhookReadRepository) *ListHandler {
	return &ListHandler{readRepo: readRepo}
}

// Handle executes the ListQuery and returns a list of WebhookView with total count.
func (h *ListHandler) Handle(ctx context.Context, q ListQuery) (*ListResult, error) {
	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		return nil, err
	}

	result := make([]*appdto.WebhookView, len(views))
	for i, v := range views {
		result[i] = &appdto.WebhookView{
			ID:        v.ID,
			Name:      v.Name,
			URL:       v.URL,
			Secret:    v.Secret,
			Events:    v.Events,
			Enabled:   v.Enabled,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
		}
	}

	return &ListResult{
		Webhooks: result,
		Total:    total,
	}, nil
}
