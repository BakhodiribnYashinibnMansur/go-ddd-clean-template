package command

import (
	"context"

	"gct/internal/context/content/file/domain"
	"gct/internal/platform/application"
	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/logger"
	"gct/internal/platform/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// CreateFileCommand captures metadata for a newly uploaded file to be persisted as a domain record.
// UploadedBy is optional — nil indicates an anonymous or system-initiated upload.
// The caller must ensure the physical file already exists at Path before issuing this command.
type CreateFileCommand struct {
	Name         string
	OriginalName string
	MimeType     string
	Size         int64
	Path         string
	URL          string
	UploadedBy   *uuid.UUID
}

// CreateFileHandler persists file metadata and publishes domain events upon successful creation.
// It does not handle the physical file upload — only the database record and event propagation.
// Callers are responsible for authorization and virus/malware scanning before invoking this handler.
type CreateFileHandler struct {
	repo     domain.FileRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateFileHandler wires up the handler with its required dependencies.
func NewCreateFileHandler(
	repo domain.FileRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *CreateFileHandler {
	return &CreateFileHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle creates a file domain entity and persists it via the repository.
// On success, domain events (e.g., FileCreated) are published; event publish failures are logged but do not fail the operation.
func (h *CreateFileHandler) Handle(ctx context.Context, cmd CreateFileCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateFileHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "CreateFile", "file")()

	f := domain.NewFile(cmd.Name, cmd.OriginalName, cmd.MimeType, cmd.Size, cmd.Path, cmd.URL, cmd.UploadedBy)

	if err := h.repo.Save(ctx, f); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateFile", Entity: "file", Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, f.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "CreateFile", Entity: "file", Err: err}.KV()...)
	}

	return nil
}
