package query

import (
	"context"

	appdto "gct/internal/emailtemplate/application"
	"gct/internal/emailtemplate/domain"
)

// ListQuery holds the input for listing email templates with filtering.
type ListQuery struct {
	Filter domain.EmailTemplateFilter
}

// ListResult holds the output of the list email templates query.
type ListResult struct {
	Templates []*appdto.EmailTemplateView
	Total     int64
}

// ListHandler handles the ListQuery.
type ListHandler struct {
	readRepo domain.EmailTemplateReadRepository
}

// NewListHandler creates a new ListHandler.
func NewListHandler(readRepo domain.EmailTemplateReadRepository) *ListHandler {
	return &ListHandler{readRepo: readRepo}
}

// Handle executes the ListQuery and returns a list of EmailTemplateView with total count.
func (h *ListHandler) Handle(ctx context.Context, q ListQuery) (*ListResult, error) {
	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		return nil, err
	}

	result := make([]*appdto.EmailTemplateView, len(views))
	for i, v := range views {
		result[i] = &appdto.EmailTemplateView{
			ID:        v.ID,
			Name:      v.Name,
			Subject:   v.Subject,
			HTMLBody:  v.HTMLBody,
			TextBody:  v.TextBody,
			Variables: v.Variables,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
		}
	}

	return &ListResult{
		Templates: result,
		Total:     total,
	}, nil
}
