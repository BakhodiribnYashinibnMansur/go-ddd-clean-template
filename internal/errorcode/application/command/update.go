package command

import (
	"context"

	"gct/internal/errorcode/domain"
	"gct/internal/shared/application"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// UpdateErrorCodeCommand represents a full replacement of an error code's mutable fields.
// The Code field is immutable after creation — only Message, HTTPStatus, and behavioral hints can be changed.
// All fields are required (non-pointer), so every update is a full overwrite of the mutable attributes.
type UpdateErrorCodeCommand struct {
	ID         uuid.UUID
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

// UpdateErrorCodeHandler applies a full replacement of an error code's mutable fields using a fetch-mutate-persist pattern.
// Event publish failures are logged but do not roll back the persisted changes.
type UpdateErrorCodeHandler struct {
	repo     domain.ErrorCodeRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateErrorCodeHandler wires dependencies for error code updates.
func NewUpdateErrorCodeHandler(
	repo domain.ErrorCodeRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *UpdateErrorCodeHandler {
	return &UpdateErrorCodeHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle fetches the error code by ID, overwrites all mutable fields, persists, and publishes events.
// Returns a repository error if the error code is not found or the update fails.
func (h *UpdateErrorCodeHandler) Handle(ctx context.Context, cmd UpdateErrorCodeCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "UpdateErrorCodeHandler.Handle")
	defer func() { end(err) }()

	ec, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	ec.Update(
		cmd.Message, cmd.MessageUz, cmd.MessageRu, cmd.HTTPStatus,
		cmd.Category, cmd.Severity,
		cmd.Retryable, cmd.RetryAfter, cmd.Suggestion,
	)

	if err := h.repo.Update(ctx, ec); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "UpdateErrorCode", Entity: "error_code", EntityID: cmd.ID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, ec.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "UpdateErrorCode", Entity: "error_code", Err: err}.KV()...)
	}

	return nil
}
