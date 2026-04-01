package command

import (
	"context"

	"gct/internal/errorcode/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"
)

// CreateErrorCodeCommand represents an intent to register a new standardized error code in the system.
// Code is the stable identifier returned to API consumers (e.g., "AUTH_001"). Retryable and RetryAfter
// instruct clients whether and when to retry; Suggestion provides human-readable remediation guidance.
type CreateErrorCodeCommand struct {
	Code       string
	Message    string
	HTTPStatus int
	Category   string
	Severity   string
	Retryable  bool
	RetryAfter int
	Suggestion string
}

// CreateErrorCodeHandler orchestrates error code creation and emits domain events for downstream consumers.
// Event publish failures are logged but do not roll back the persisted error code.
type CreateErrorCodeHandler struct {
	repo     domain.ErrorCodeRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateErrorCodeHandler wires dependencies for error code creation.
func NewCreateErrorCodeHandler(
	repo domain.ErrorCodeRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *CreateErrorCodeHandler {
	return &CreateErrorCodeHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle persists a new error code and publishes its domain events.
// Returns nil on success; propagates repository errors (e.g., duplicate code) to the caller.
func (h *CreateErrorCodeHandler) Handle(ctx context.Context, cmd CreateErrorCodeCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateErrorCodeHandler.Handle")
	defer func() { end(err) }()

	ec := domain.NewErrorCode(
		cmd.Code, cmd.Message, cmd.HTTPStatus,
		cmd.Category, cmd.Severity,
		cmd.Retryable, cmd.RetryAfter, cmd.Suggestion,
	)

	if err := h.repo.Save(ctx, ec); err != nil {
		h.logger.Errorf("failed to save error code: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, ec.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
