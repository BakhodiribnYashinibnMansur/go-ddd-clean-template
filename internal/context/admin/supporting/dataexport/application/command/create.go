package command

import (
	"context"

	exportentity "gct/internal/context/admin/supporting/dataexport/domain/entity"
	exportrepo "gct/internal/context/admin/supporting/dataexport/domain/repository"
	shareddomain "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/outbox"

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
type CreateDataExportHandler struct {
	repo      exportrepo.DataExportRepository
	committer *outbox.EventCommitter
	logger    logger.Log
}

// NewCreateDataExportHandler wires dependencies for data export creation.
func NewCreateDataExportHandler(
	repo exportrepo.DataExportRepository,
	committer *outbox.EventCommitter,
	logger logger.Log,
) *CreateDataExportHandler {
	return &CreateDataExportHandler{
		repo:      repo,
		committer: committer,
		logger:    logger,
	}
}

// Handle persists a new data export record in pending state and publishes domain events.
// Returns nil on success; propagates repository errors to the caller.
func (h *CreateDataExportHandler) Handle(ctx context.Context, cmd CreateDataExportCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateDataExportHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "CreateDataExport", "data_export")()

	de := exportentity.NewDataExport(cmd.UserID, cmd.DataType, cmd.Format)

	return h.committer.Commit(ctx, func(ctx context.Context, q shareddomain.Querier) error {
		if err := h.repo.Save(ctx, q, de); err != nil {
			h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateDataExport", Entity: "data_export", Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, de.Events)
}
