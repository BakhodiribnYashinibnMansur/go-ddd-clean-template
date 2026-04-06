package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"

	"gct/internal/context/admin/supporting/integration/application/dto"
	integentity "gct/internal/context/admin/supporting/integration/domain/entity"
	integrepo "gct/internal/context/admin/supporting/integration/domain/repository"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// ValidateAPIKeyQuery holds the input for validating an API key.
type ValidateAPIKeyQuery struct {
	APIKey string
}

// ValidateAPIKeyHandler handles the ValidateAPIKeyQuery.
type ValidateAPIKeyHandler struct {
	readRepo integrepo.IntegrationReadRepository
	logger   logger.Log
	l        logger.Log
}

// NewValidateAPIKeyHandler creates a new ValidateAPIKeyHandler.
func NewValidateAPIKeyHandler(readRepo integrepo.IntegrationReadRepository, l logger.Log) *ValidateAPIKeyHandler {
	return &ValidateAPIKeyHandler{readRepo: readRepo, l: l}
}

// Handle executes the ValidateAPIKeyQuery and returns an APIKeyView if the key is valid and active.
func (h *ValidateAPIKeyHandler) Handle(ctx context.Context, q ValidateAPIKeyQuery) (result *dto.APIKeyView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ValidateAPIKeyHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ValidateApiKey", "integration")()

	view, err := h.readRepo.FindByAPIKey(ctx, q.APIKey)
	if err != nil {
		h.l.Warnw("api key validation failed", "error", err)
		return nil, apperrors.MapToServiceError(err)
	}

	if !view.Active {
		h.l.Warnw("api key is inactive", "api_key_id", view.ID)
		return nil, integentity.ErrAPIKeyInactive
	}

	return &dto.APIKeyView{
		ID:            view.ID,
		IntegrationID: view.IntegrationID.UUID(),
		Key:           view.Key,
		Active:        view.Active,
	}, nil
}
