package dataexport

import (
	"context"
	"encoding/json"
	"time"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) Create(ctx context.Context, req domain.CreateDataExportRequest, userID string) (*domain.DataExport, error) {
	id := uuid.New().String()
	filters := req.Filters
	if len(filters) == 0 {
		filters = json.RawMessage("{}")
	}
	now := time.Now()
	export := &domain.DataExport{
		ID:          id,
		Type:        req.Type,
		Status:      "completed",
		Filters:     filters,
		CreatedBy:   &userID,
		CompletedAt: &now,
	}
	if err := uc.repo.Create(ctx, export); err != nil {
		return nil, err
	}
	return export, nil
}
