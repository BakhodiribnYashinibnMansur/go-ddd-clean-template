package command

import (
	"context"

	"gct/internal/integration/domain"
	"gct/internal/shared/application"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// UpdateCommand represents a partial update to an existing integration identified by ID.
// Pointer fields implement patch semantics — nil means "leave unchanged," non-nil means "overwrite."
// Callers must provide at least one non-nil field for the update to be meaningful.
type UpdateCommand struct {
	ID         uuid.UUID
	Name       *string
	Type       *string
	APIKey     *string
	WebhookURL *string
	Enabled    *bool
	Config     *map[string]string
}

// UpdateHandler applies partial modifications to an existing integration via fetch-then-update.
// Callers are responsible for authorization; this handler only enforces repository-level constraints.
type UpdateHandler struct {
	repo     domain.IntegrationRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateHandler wires up the handler with its required dependencies.
func NewUpdateHandler(
	repo domain.IntegrationRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *UpdateHandler {
	return &UpdateHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle fetches the integration by ID, applies the patch via domain logic, and persists the result.
// Returns a repository error if the integration is not found. Event publish failures are logged but non-fatal.
func (h *UpdateHandler) Handle(ctx context.Context, cmd UpdateCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "UpdateHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "UpdateIntegration", "integration")()

	i, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	i.UpdateDetails(cmd.Name, cmd.Type, cmd.APIKey, cmd.WebhookURL, cmd.Enabled, cmd.Config)

	if err := h.repo.Update(ctx, i); err != nil {
		h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "UpdateIntegration", Entity: "integration", EntityID: cmd.ID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, i.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "UpdateIntegration", Entity: "integration", EntityID: cmd.ID, Err: err}.KV()...)
	}

	return nil
}
