package query

import (
	"context"

	appdto "gct/internal/integration/application"
	"gct/internal/integration/domain"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"
)

// ValidateAPIKeyQuery holds the input for validating an API key.
type ValidateAPIKeyQuery struct {
	APIKey string
}

// ValidateAPIKeyHandler handles the ValidateAPIKeyQuery.
type ValidateAPIKeyHandler struct {
	readRepo domain.IntegrationReadRepository
	l        logger.Log
}

// NewValidateAPIKeyHandler creates a new ValidateAPIKeyHandler.
func NewValidateAPIKeyHandler(readRepo domain.IntegrationReadRepository, l logger.Log) *ValidateAPIKeyHandler {
	return &ValidateAPIKeyHandler{readRepo: readRepo, l: l}
}

// Handle executes the ValidateAPIKeyQuery and returns an APIKeyView if the key is valid and active.
func (h *ValidateAPIKeyHandler) Handle(ctx context.Context, q ValidateAPIKeyQuery) (result *appdto.APIKeyView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ValidateAPIKeyHandler.Handle")
	defer func() { end(err) }()

	view, err := h.readRepo.FindByAPIKey(ctx, q.APIKey)
	if err != nil {
		h.l.Warnw("api key validation failed", "error", err)
		return nil, err
	}

	if !view.Active {
		h.l.Warnw("api key is inactive", "api_key_id", view.ID)
		return nil, domain.ErrAPIKeyInactive
	}

	return &appdto.APIKeyView{
		ID:            view.ID,
		IntegrationID: view.IntegrationID,
		Key:           view.Key,
		Active:        view.Active,
	}, nil
}
