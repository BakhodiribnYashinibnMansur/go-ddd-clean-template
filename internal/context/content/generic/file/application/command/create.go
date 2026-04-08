package command

import (
	"context"

	fileentity "gct/internal/context/content/generic/file/domain/entity"
	filerepo "gct/internal/context/content/generic/file/domain/repository"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/outbox"

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
	repo      filerepo.FileRepository
	committer *outbox.EventCommitter
	logger    logger.Log
}

// NewCreateFileHandler wires up the handler with its required dependencies.
func NewCreateFileHandler(
	repo filerepo.FileRepository,
	committer *outbox.EventCommitter,
	logger logger.Log,
) *CreateFileHandler {
	return &CreateFileHandler{
		repo:      repo,
		committer: committer,
		logger:    logger,
	}
}

// Handle creates a file domain entity and persists it via the repository.
// On success, domain events (e.g., FileCreated) are published; event publish failures are logged but do not fail the operation.
func (h *CreateFileHandler) Handle(ctx context.Context, cmd CreateFileCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateFileHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "CreateFile", "file")()

	f := fileentity.NewFile(cmd.Name, cmd.OriginalName, cmd.MimeType, cmd.Size, cmd.Path, cmd.URL, cmd.UploadedBy)

	return h.committer.Commit(ctx, func(ctx context.Context) error {
		if err := h.repo.Save(ctx, f); err != nil {
			h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateFile", Entity: "file", Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, f.Events)
}
