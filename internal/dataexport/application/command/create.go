package command

import (
	"context"

	"gct/internal/dataexport/domain"
	"gct/internal/shared/application"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// CreateDataExportCommand represents an intent to initiate a data export job for a specific user.
// DataType identifies the data category (e.g., "users", "audit_logs") and Format sets the output encoding (e.g., "csv", "json").
// The export starts in a pending state; a background worker is expected to pick it up via domain events.
type CreateDataExportCommand struct {
	UserID   uuid.UUID
	DataType string
	Format   string
}

// CreateDataExportHandler orchestrates data export creation and emits domain events for async processing.
// Event publish failures are logged but do not roll back the persisted export record.
type CreateDataExportHandler struct {
	repo     domain.DataExportRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateDataExportHandler wires dependencies for data export creation.
func NewCreateDataExportHandler(
	repo domain.DataExportRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *CreateDataExportHandler {
	return &CreateDataExportHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle persists a new data export record in pending state and publishes domain events.
// Returns nil on success; propagates repository errors to the caller.
func (h *CreateDataExportHandler) Handle(ctx context.Context, cmd CreateDataExportCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateDataExportHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "CreateDataExport", "data_export")()

	de := domain.NewDataExport(cmd.UserID, cmd.DataType, cmd.Format)

	if err := h.repo.Save(ctx, de); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateDataExport", Entity: "data_export", Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, de.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "CreateDataExport", Entity: "data_export", Err: err}.KV()...)
	}

	return nil
}
