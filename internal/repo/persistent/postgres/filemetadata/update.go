package filemetadata

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
)

// Update patches the original_name of a file_metadata record and returns the updated row.
func (r *Repo) Update(ctx context.Context, id string, req domain.UpdateFileMetadataRequest) (*domain.FileMetadata, error) {
	q := r.builder.Update(table)

	if req.OriginalName != nil {
		q = q.Set("original_name", *req.OriginalName)
	}

	q = q.Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": id}).
		Suffix("RETURNING id, original_name, stored_name, bucket, url, size, mime_type, uploaded_by, created_at, updated_at")

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build update")
	}

	var f domain.FileMetadata
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&f.ID, &f.OriginalName, &f.StoredName, &f.Bucket,
		&f.URL, &f.Size, &f.MimeType, &f.UploadedBy,
		&f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, table, nil)
	}
	return &f, nil
}
