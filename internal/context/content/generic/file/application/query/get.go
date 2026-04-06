package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	"gct/internal/context/content/generic/file/application/dto"
	fileentity "gct/internal/context/content/generic/file/domain/entity"
	filerepo "gct/internal/context/content/generic/file/domain/repository"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// GetFileQuery holds the input for getting a single file.
type GetFileQuery struct {
	ID fileentity.FileID
}

// GetFileHandler handles the GetFileQuery.
type GetFileHandler struct {
	readRepo filerepo.FileReadRepository
	logger   logger.Log
}

// NewGetFileHandler creates a new GetFileHandler.
func NewGetFileHandler(readRepo filerepo.FileReadRepository, l logger.Log) *GetFileHandler {
	return &GetFileHandler{readRepo: readRepo, logger: l}
}

// Handle executes the GetFileQuery and returns a FileView.
func (h *GetFileHandler) Handle(ctx context.Context, q GetFileQuery) (result *dto.FileView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "GetFileHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "GetFile", "file")()

	v, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "GetFile", Entity: "file", EntityID: q.ID.String(), Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	return &dto.FileView{
		ID:           uuid.UUID(v.ID),
		Name:         v.Name,
		OriginalName: v.OriginalName,
		MimeType:     v.MimeType,
		Size:         v.Size,
		Path:         v.Path,
		URL:          v.URL,
		UploadedBy:   v.UploadedBy,
		CreatedAt:    v.CreatedAt,
	}, nil
}
