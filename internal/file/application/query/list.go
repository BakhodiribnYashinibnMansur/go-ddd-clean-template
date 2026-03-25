package query

import (
	"context"

	appdto "gct/internal/file/application"
	"gct/internal/file/domain"
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
}

// NewListFilesHandler creates a new ListFilesHandler.
func NewListFilesHandler(readRepo domain.FileReadRepository) *ListFilesHandler {
	return &ListFilesHandler{readRepo: readRepo}
}

// Handle executes the ListFilesQuery and returns a list of FileView with total count.
func (h *ListFilesHandler) Handle(ctx context.Context, q ListFilesQuery) (*ListFilesResult, error) {
	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		return nil, err
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
