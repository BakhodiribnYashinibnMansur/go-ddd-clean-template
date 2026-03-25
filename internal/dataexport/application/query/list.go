package query

import (
	"context"

	appdto "gct/internal/dataexport/application"
	"gct/internal/dataexport/domain"
)

// ListDataExportsQuery holds the input for listing data exports with filtering.
type ListDataExportsQuery struct {
	Filter domain.DataExportFilter
}

// ListDataExportsResult holds the output of the list data exports query.
type ListDataExportsResult struct {
	Exports []*appdto.DataExportView
	Total   int64
}

// ListDataExportsHandler handles the ListDataExportsQuery.
type ListDataExportsHandler struct {
	readRepo domain.DataExportReadRepository
}

// NewListDataExportsHandler creates a new ListDataExportsHandler.
func NewListDataExportsHandler(readRepo domain.DataExportReadRepository) *ListDataExportsHandler {
	return &ListDataExportsHandler{readRepo: readRepo}
}

// Handle executes the ListDataExportsQuery and returns a list of DataExportView with total count.
func (h *ListDataExportsHandler) Handle(ctx context.Context, q ListDataExportsQuery) (*ListDataExportsResult, error) {
	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		return nil, err
	}

	result := make([]*appdto.DataExportView, len(views))
	for i, v := range views {
		result[i] = &appdto.DataExportView{
			ID:        v.ID,
			UserID:    v.UserID,
			DataType:  v.DataType,
			Format:    v.Format,
			Status:    v.Status,
			FileURL:   v.FileURL,
			Error:     v.Error,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
		}
	}

	return &ListDataExportsResult{
		Exports: result,
		Total:   total,
	}, nil
}
