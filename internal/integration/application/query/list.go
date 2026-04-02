package query

import (
	"context"

	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/logger"

	appdto "gct/internal/integration/application"
	"gct/internal/integration/domain"
	"gct/internal/shared/infrastructure/pgxutil"
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
	logger   logger.Log
}

// NewListHandler creates a new ListHandler.
func NewListHandler(readRepo domain.IntegrationReadRepository, l logger.Log) *ListHandler {
	return &ListHandler{readRepo: readRepo, logger: l}
}

// Handle executes the ListQuery and returns a list of IntegrationView with total count.
func (h *ListHandler) Handle(ctx context.Context, q ListQuery) (_ *ListResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListHandler.Handle")
	defer func() { end(err) }()

	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "List", Entity: "integration", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
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
