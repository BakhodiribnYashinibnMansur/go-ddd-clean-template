package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	"gct/internal/context/content/generic/file/application/dto"
	filerepo "gct/internal/context/content/generic/file/domain/repository"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// ListFilesQuery holds the input for listing files with filtering.
type ListFilesQuery struct {
	Filter filerepo.FileFilter
}

// ListFilesResult holds the output of the list files query.
type ListFilesResult struct {
	Files []*dto.FileView
	Total int64
}

// ListFilesHandler handles the ListFilesQuery.
type ListFilesHandler struct {
	readRepo filerepo.FileReadRepository
	logger   logger.Log
}

// NewListFilesHandler creates a new ListFilesHandler.
func NewListFilesHandler(readRepo filerepo.FileReadRepository, l logger.Log) *ListFilesHandler {
	return &ListFilesHandler{readRepo: readRepo, logger: l}
}

// Handle executes the ListFilesQuery and returns a list of FileView with total count.
func (h *ListFilesHandler) Handle(ctx context.Context, q ListFilesQuery) (_ *ListFilesResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListFilesHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ListFiles", "file")()

	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "ListFiles", Entity: "file", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	result := make([]*dto.FileView, len(views))
	for i, v := range views {
		result[i] = &dto.FileView{
			ID:           uuid.UUID(v.ID),
			Name:         v.Name,
			OriginalName: v.OriginalName,
			MimeType:     v.MimeType,
			Size:         v.Size,
			Path:         v.Path,
			URL:          v.URL,
			UploadedBy:   v.UploadedBy,
			CreatedAt:    v.CreatedAt,
		}
	}

	return &ListFilesResult{
		Files: result,
		Total: total,
	}, nil
}
