package filemetadata

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

// List returns a paginated slice of FileMetadata records and the total count.
func (r *Repo) List(ctx context.Context, filter domain.FileMetadataFilter) ([]domain.FileMetadata, int64, error) {
	q := r.builder.
		Select("id", "original_name", "stored_name", "bucket", "url", "size", "mime_type", "uploaded_by", "created_at", "updated_at").
		From(table)

	if filter.Search != "" {
		q = q.Where(squirrel.ILike{"original_name": "%" + filter.Search + "%"})
	}
	if filter.MimeType != "" {
		q = q.Where(squirrel.Eq{"mime_type": filter.MimeType})
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

	listSQL, args, err := q.OrderBy("created_at DESC").ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build list")
	}

	rows, err := r.pool.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, table, nil)
	}
	defer rows.Close()

	var items []domain.FileMetadata
	for rows.Next() {
		var f domain.FileMetadata
		if err := rows.Scan(
			&f.ID, &f.OriginalName, &f.StoredName, &f.Bucket,
			&f.URL, &f.Size, &f.MimeType, &f.UploadedBy,
			&f.CreatedAt, &f.UpdatedAt,
		); err != nil {
			return nil, 0, apperrors.HandlePgError(err, table, nil)
		}
		items = append(items, f)
	}
	if items == nil {
		items = []domain.FileMetadata{}
	}
	return items, total, nil
}
