package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	"gct/internal/context/admin/supporting/integration/application/dto"
	integentity "gct/internal/context/admin/supporting/integration/domain/entity"
	integrepo "gct/internal/context/admin/supporting/integration/domain/repository"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// GetQuery holds the input for fetching a single integration.
type GetQuery struct {
	ID integentity.IntegrationID
}

// GetHandler handles the GetQuery.
type GetHandler struct {
	readRepo integrepo.IntegrationReadRepository
	logger   logger.Log
}

// NewGetHandler creates a new GetHandler.
func NewGetHandler(readRepo integrepo.IntegrationReadRepository, l logger.Log) *GetHandler {
	return &GetHandler{readRepo: readRepo, logger: l}
}

// Handle executes the GetQuery and returns an IntegrationView.
func (h *GetHandler) Handle(ctx context.Context, q GetQuery) (result *dto.IntegrationView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "GetIntegration", "integration")()

	view, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "Get", Entity: "integration", EntityID: q.ID, Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	return &dto.IntegrationView{
		ID:         view.ID.UUID(),
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
