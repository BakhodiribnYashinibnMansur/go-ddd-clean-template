package command

import (
	"context"

	"gct/internal/context/admin/supporting/errorcode/domain"
	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// CreateErrorCodeCommand represents an intent to register a new standardized error code in the system.
// Code is the stable identifier returned to API consumers (e.g., "AUTH_001"). Retryable and RetryAfter
// instruct clients whether and when to retry; Suggestion provides human-readable remediation guidance.
type CreateErrorCodeCommand struct {
	Code       string
	Message    string
	MessageUz  string
	MessageRu  string
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
	defer logger.SlowOp(h.logger, ctx, "CreateErrorCode", "error_code")()

	ec := domain.NewErrorCode(
		cmd.Code, cmd.Message, cmd.HTTPStatus,
		cmd.Category, cmd.Severity,
		cmd.Retryable, cmd.RetryAfter, cmd.Suggestion,
	)
	ec.SetTranslations(cmd.MessageUz, cmd.MessageRu)

	if err := h.repo.Save(ctx, ec); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateErrorCode", Entity: "error_code", EntityID: cmd.Code, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, ec.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "CreateErrorCode", Entity: "error_code", Err: err}.KV()...)
	}

	return nil
}
