package command

import (
	"context"
	"fmt"

	integentity "gct/internal/context/admin/supporting/integration/domain/entity"
	integrepo "gct/internal/context/admin/supporting/integration/domain/repository"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/outbox"
)

// CreateCommand represents an intent to register a new third-party integration.
// Config carries provider-specific settings (e.g., Slack channel, SMTP host) and is stored as schemaless JSON.
// The APIKey is persisted as-is — callers should encrypt sensitive credentials before constructing this command.
type CreateCommand struct {
	Name       string
	Type       string
	APIKey     string
	WebhookURL string
	Enabled    bool
	Config     map[string]string
}

// CreateHandler orchestrates integration creation through the repository and event bus.
// Domain events are emitted on success so downstream listeners can initialize provider connections.
type CreateHandler struct {
	repo      integrepo.IntegrationRepository
	committer *outbox.EventCommitter
	logger    logger.Log
}

// NewCreateHandler wires up the handler with its required dependencies.
func NewCreateHandler(
	repo integrepo.IntegrationRepository,
	committer *outbox.EventCommitter,
	logger logger.Log,
) *CreateHandler {
	return &CreateHandler{
		repo:      repo,
		committer: committer,
		logger:    logger,
	}
}

// Handle persists the new integration and publishes domain events.
func (h *CreateHandler) Handle(ctx context.Context, cmd CreateCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "CreateIntegration", "integration")()

	i, err := integentity.NewIntegration(cmd.Name, cmd.Type, cmd.APIKey, cmd.WebhookURL, cmd.Enabled, cmd.Config)
	if err != nil {
		return fmt.Errorf("create_integration: %w", err)
	}

	return h.committer.Commit(ctx, func(ctx context.Context) error {
		if err := h.repo.Save(ctx, i); err != nil {
			h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateIntegration", Entity: "integration", Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, i.Events)
}
