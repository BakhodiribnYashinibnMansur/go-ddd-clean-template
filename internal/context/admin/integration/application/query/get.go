package query

import (
	"context"

	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/logger"

	appdto "gct/internal/context/admin/integration/application"
	"gct/internal/context/admin/integration/domain"
	"gct/internal/platform/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// GetQuery holds the input for fetching a single integration.
type GetQuery struct {
	ID uuid.UUID
}

// GetHandler handles the GetQuery.
type GetHandler struct {
	readRepo domain.IntegrationReadRepository
	logger   logger.Log
}

// NewGetHandler creates a new GetHandler.
func NewGetHandler(readRepo domain.IntegrationReadRepository, l logger.Log) *GetHandler {
	return &GetHandler{readRepo: readRepo, logger: l}
}

// Handle executes the GetQuery and returns an IntegrationView.
func (h *GetHandler) Handle(ctx context.Context, q GetQuery) (result *appdto.IntegrationView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "GetIntegration", "integration")()

	view, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "Get", Entity: "integration", EntityID: q.ID, Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	return &appdto.IntegrationView{
		ID:         view.ID,
		Name:       view.Name,
		Type:       view.Type,
		APIKey:     view.APIKey,
		WebhookURL: view.WebhookURL,
		Enabled:    view.Enabled,
		Config:     view.Config,
		CreatedAt:  view.CreatedAt,
		UpdatedAt:  view.UpdatedAt,
	}, nil
}
