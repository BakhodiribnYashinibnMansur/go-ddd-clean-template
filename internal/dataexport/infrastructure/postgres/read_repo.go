package postgres

import (
	"context"
	"time"

	"gct/internal/dataexport/domain"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var readColumns = []string{
	"id", "user_id", "data_type", "format", "status",
	"file_url", "error", "created_at", "updated_at",
}

// DataExportReadRepo implements domain.DataExportReadRepository for the CQRS read side.
type DataExportReadRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewDataExportReadRepo creates a new DataExportReadRepo.
func NewDataExportReadRepo(pool *pgxpool.Pool) *DataExportReadRepo {
	return &DataExportReadRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// FindByID returns a single DataExportView by its ID.
func (r *DataExportReadRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.DataExportView, error) {
	sql, args, err := r.builder.
		Select(readColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanDataExportView(row)
}

// List returns a paginated list of DataExportView with optional filters.
func (r *DataExportReadRepo) List(ctx context.Context, filter domain.DataExportFilter) ([]*domain.DataExportView, int64, error) {
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

	var total int64
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

	var views []*domain.DataExportView
	for rows.Next() {
		v, err := scanDataExportViewFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		views = append(views, v)
	}

	return views, total, nil
}

func applyFilters(conds squirrel.And, filter domain.DataExportFilter) squirrel.And {
	if filter.UserID != nil {
		conds = append(conds, squirrel.Eq{"user_id": *filter.UserID})
	}
	if filter.DataType != nil {
		conds = append(conds, squirrel.Eq{"data_type": *filter.DataType})
	}
	if filter.Status != nil {
		conds = append(conds, squirrel.Eq{"status": *filter.Status})
	}
	return conds
}

func scanDataExportView(row pgx.Row) (*domain.DataExportView, error) {
	var (
		id        uuid.UUID
		userID    uuid.UUID
		dataType  string
		format    string
		status    string
		fileURL   *string
		errorMsg  *string
		createdAt time.Time
		updatedAt time.Time
	)

	err := row.Scan(&id, &userID, &dataType, &format, &status, &fileURL, &errorMsg, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	return &domain.DataExportView{
		ID:        id,
		UserID:    userID,
		DataType:  dataType,
		Format:    format,
		Status:    status,
		FileURL:   fileURL,
		Error:     errorMsg,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func scanDataExportViewFromRows(rows pgx.Rows) (*domain.DataExportView, error) {
	var (
		id        uuid.UUID
		userID    uuid.UUID
		dataType  string
		format    string
		status    string
		fileURL   *string
		errorMsg  *string
		createdAt time.Time
		updatedAt time.Time
	)

	err := rows.Scan(&id, &userID, &dataType, &format, &status, &fileURL, &errorMsg, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	return &domain.DataExportView{
		ID:        id,
		UserID:    userID,
		DataType:  dataType,
		Format:    format,
		Status:    status,
		FileURL:   fileURL,
		Error:     errorMsg,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}
