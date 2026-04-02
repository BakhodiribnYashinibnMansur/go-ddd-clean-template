package query

import (
	"context"

	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/logger"

	appdto "gct/internal/file/application"
	"gct/internal/file/domain"
	"gct/internal/shared/infrastructure/pgxutil"
)

// ListFilesQuery holds the input for listing files with filtering.
type ListFilesQuery struct {
	Filter domain.FileFilter
}

// ListFilesResult holds the output of the list files query.
type ListFilesResult struct {
	Files []*appdto.FileView
	Total int64
}

// ListFilesHandler handles the ListFilesQuery.
type ListFilesHandler struct {
	readRepo domain.FileReadRepository
	logger   logger.Log
}

// NewListFilesHandler creates a new ListFilesHandler.
func NewListFilesHandler(readRepo domain.FileReadRepository, l logger.Log) *ListFilesHandler {
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

	result := make([]*appdto.FileView, len(views))
	for i, v := range views {
		result[i] = &appdto.FileView{
			ID:           v.ID,
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
