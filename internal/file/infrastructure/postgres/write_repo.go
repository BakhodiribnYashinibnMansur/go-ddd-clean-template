package postgres

import (
	"context"

	"gct/internal/file/domain"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = consts.TableFileMetadata

var writeColumns = []string{
	"id", "name", "original_name", "mime_type", "size",
	"path", "url", "uploaded_by", "created_at",
}

// FileWriteRepo implements domain.FileRepository using PostgreSQL.
type FileWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewFileWriteRepo creates a new FileWriteRepo.
func NewFileWriteRepo(pool *pgxpool.Pool) *FileWriteRepo {
	return &FileWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new File aggregate into the database.
func (r *FileWriteRepo) Save(ctx context.Context, f *domain.File) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns(writeColumns...).
		Values(
			f.ID(),
			f.Name(),
			f.OriginalName(),
			f.MimeType(),
			f.Size(),
			f.Path(),
			f.URL(),
			f.UploadedBy(),
			f.CreatedAt(),
		).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}
