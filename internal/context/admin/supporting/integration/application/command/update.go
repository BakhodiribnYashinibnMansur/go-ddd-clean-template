package command

import (
	"context"

	integentity "gct/internal/context/admin/supporting/integration/domain/entity"
	integrepo "gct/internal/context/admin/supporting/integration/domain/repository"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/outbox"
)

// UpdateCommand represents a partial update to an existing integration identified by ID.
// Pointer fields implement patch semantics — nil means "leave unchanged," non-nil means "overwrite."
// Callers must provide at least one non-nil field for the update to be meaningful.
type UpdateCommand struct {
	ID         integentity.IntegrationID
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
	repo      integrepo.IntegrationRepository
	committer *outbox.EventCommitter
	logger    logger.Log
}

// NewUpdateHandler wires up the handler with its required dependencies.
func NewUpdateHandler(
	repo integrepo.IntegrationRepository,
	committer *outbox.EventCommitter,
	logger logger.Log,
) *UpdateHandler {
	return &UpdateHandler{
		repo:      repo,
		committer: committer,
		logger:    logger,
	}
}

// Handle fetches the integration by ID, applies the patch via domain logic, and persists the result.
// Returns a repository error if the integration is not found.
func (h *UpdateHandler) Handle(ctx context.Context, cmd UpdateCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "UpdateHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "UpdateIntegration", "integration")()

	i, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	i.UpdateDetails(cmd.Name, cmd.Type, cmd.APIKey, cmd.WebhookURL, cmd.Enabled, cmd.Config)

	return h.committer.Commit(ctx, func(ctx context.Context) error {
		if err := h.repo.Update(ctx, i); err != nil {
			h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "UpdateIntegration", Entity: "integration", EntityID: cmd.ID, Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, i.Events)
}
