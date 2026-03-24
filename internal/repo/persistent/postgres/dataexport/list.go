package dataexport

import (
	"context"

	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/domain"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) List(ctx context.Context, filter domain.DataExportFilter) ([]domain.DataExport, int64, error) {
	q := r.builder.
		Select("id", "type", "status", "file_url", "filters", "created_by", "created_at", "completed_at").
		From(table)

	if filter.Type != "" {
		q = q.Where(squirrel.Eq{"type": filter.Type})
	}
	if filter.Status != "" {
		q = q.Where(squirrel.Eq{"status": filter.Status})
	}

	countSQL, countArgs, _ := r.builder.Select("COUNT(*)").FromSelect(q, "sub").ToSql()
	var total int64
	if err := r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.HandlePgError(err, table, nil)
	}

	if filter.Limit > 0 {
		q = q.Limit(uint64(filter.Limit))
	}
	if filter.Offset > 0 {
		q = q.Offset(uint64(filter.Offset))
	}
	sql, args, err := q.OrderBy("created_at DESC").ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build list")
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, table, nil)
	}
	defer rows.Close()

	var items []domain.DataExport
	for rows.Next() {
		var e domain.DataExport
		if err := rows.Scan(&e.ID, &e.Type, &e.Status, &e.FileURL, &e.Filters, &e.CreatedBy, &e.CreatedAt, &e.CompletedAt); err != nil {
			return nil, 0, apperrors.HandlePgError(err, table, nil)
		}
		items = append(items, e)
	}
	if items == nil {
		items = []domain.DataExport{}
	}
	return items, total, nil
}
