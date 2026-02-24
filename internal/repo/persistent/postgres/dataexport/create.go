package dataexport

import (
	"context"

	apperrors "gct/pkg/errors"
	"gct/internal/domain"
)

func (r *Repo) Create(ctx context.Context, e *domain.DataExport) error {
	filters := string(e.Filters)
	if filters == "" {
		filters = "{}"
	}
	sql, args, err := r.builder.
		Insert(table).
		Columns("id", "type", "status", "file_url", "filters", "created_by", "completed_at").
		Values(e.ID, e.Type, e.Status, e.FileURL, filters, e.CreatedBy, e.CompletedAt).
		Suffix("RETURNING created_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build insert")
	}
	return r.pool.QueryRow(ctx, sql, args...).Scan(&e.CreatedAt)
}
