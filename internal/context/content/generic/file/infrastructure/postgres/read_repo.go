package postgres

import (
	"context"
	"time"

	"gct/internal/context/content/generic/file/domain"
	"gct/internal/kernel/consts"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var readColumns = []string{
	"id", "stored_name", "original_name", "mime_type", "size",
	"bucket", "url", "uploaded_by", "created_at",
}

// FileReadRepo implements domain.FileReadRepository for the CQRS read side.
type FileReadRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewFileReadRepo creates a new FileReadRepo.
func NewFileReadRepo(pool *pgxpool.Pool) *FileReadRepo {
	return &FileReadRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// FindByID returns a single FileView by its ID.
func (r *FileReadRepo) FindByID(ctx context.Context, id uuid.UUID) (result *domain.FileView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "FileReadRepo.FindByID")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Select(readColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanFileView(row)
}

// List returns a paginated list of FileView with optional filters.
func (r *FileReadRepo) List(ctx context.Context, filter domain.FileFilter) (items []*domain.FileView, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "FileReadRepo.List")
	defer func() { end(err) }()

	conds := squirrel.And{}
	conds = applyFilters(conds, filter)

	// Count total.
	countQB := r.builder.Select("COUNT(*)").From(tableName)
	if len(conds) > 0 {
		countQB = countQB.Where(conds)
	}
	countSQL, countArgs, err := countQB.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	if err = r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}

	// Fetch page.
	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	qb := r.builder.
		Select(readColumns...).
		From(tableName).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(filter.Offset))

	if len(conds) > 0 {
		qb = qb.Where(conds)
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}
	defer rows.Close()

	var views []*domain.FileView
	for rows.Next() {
		v, err := scanFileViewFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		views = append(views, v)
	}

	return views, total, nil
}

func applyFilters(conds squirrel.And, filter domain.FileFilter) squirrel.And {
	if filter.Name != nil {
		conds = append(conds, squirrel.ILike{"original_name": "%" + *filter.Name + "%"})
	}
	if filter.MimeType != nil {
		conds = append(conds, squirrel.Eq{"mime_type": *filter.MimeType})
	}
	return conds
}

func scanFileView(row pgx.Row) (*domain.FileView, error) {
	var (
		id           uuid.UUID
		name         string
		originalName string
		mimeType     string
		size         int64
		path         string
		url          string
		uploadedBy   *uuid.UUID
		createdAt    time.Time
	)

	err := row.Scan(&id, &name, &originalName, &mimeType, &size, &path, &url, &uploadedBy, &createdAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	return &domain.FileView{
		ID:           id,
		Name:         name,
		OriginalName: originalName,
		MimeType:     mimeType,
		Size:         size,
		Path:         path,
		URL:          url,
		UploadedBy:   uploadedBy,
		CreatedAt:    createdAt,
	}, nil
}

func scanFileViewFromRows(rows pgx.Rows) (*domain.FileView, error) {
	var (
		id           uuid.UUID
		name         string
		originalName string
		mimeType     string
		size         int64
		path         string
		url          string
		uploadedBy   *uuid.UUID
		createdAt    time.Time
	)

	err := rows.Scan(&id, &name, &originalName, &mimeType, &size, &path, &url, &uploadedBy, &createdAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	return &domain.FileView{
		ID:           id,
		Name:         name,
		OriginalName: originalName,
		MimeType:     mimeType,
		Size:         size,
		Path:         path,
		URL:          url,
		UploadedBy:   uploadedBy,
		CreatedAt:    createdAt,
	}, nil
}
