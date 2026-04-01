package command

import (
	"context"

	"gct/internal/integration/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"
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
	repo     domain.IntegrationRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateHandler wires up the handler with its required dependencies.
func NewCreateHandler(
	repo domain.IntegrationRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *CreateHandler {
	return &CreateHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle persists the new integration and publishes domain events.
// Event publish failures are logged but do not roll back the save — eventual consistency is acceptable here.
func (h *CreateHandler) Handle(ctx context.Context, cmd CreateCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateHandler.Handle")
	defer func() { end(err) }()

	i := domain.NewIntegration(cmd.Name, cmd.Type, cmd.APIKey, cmd.WebhookURL, cmd.Enabled, cmd.Config)

	if err := h.repo.Save(ctx, i); err != nil {
		h.logger.Errorf("failed to save integration: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, i.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
