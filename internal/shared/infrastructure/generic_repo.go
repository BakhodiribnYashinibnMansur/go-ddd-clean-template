package infrastructure

import (
	"context"
	"fmt"

	"gct/internal/shared/domain"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RowScanner converts a pgx.Row into an entity of type T.
type RowScanner[T any] func(row pgx.Row) (*T, error)

// BaseRepository provides generic CRUD operations using Squirrel.
type BaseRepository[T any] struct {
	pool      *pgxpool.Pool
	builder   squirrel.StatementBuilderType
	tableName string
	columns   []string
	scanner   RowScanner[T]
}

func NewBaseRepository[T any](
	pool *pgxpool.Pool,
	tableName string,
	columns []string,
	scanner RowScanner[T],
) *BaseRepository[T] {
	return &BaseRepository[T]{
		pool:      pool,
		builder:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		tableName: tableName,
		columns:   columns,
		scanner:   scanner,
	}
}

// Getters for subclasses
func (r *BaseRepository[T]) TableName() string                      { return r.tableName }
func (r *BaseRepository[T]) Columns() []string                      { return r.columns }
func (r *BaseRepository[T]) Pool() *pgxpool.Pool                    { return r.pool }
func (r *BaseRepository[T]) Builder() squirrel.StatementBuilderType { return r.builder }

// FindByID retrieves an entity by UUID, excluding soft-deleted records.
func (r *BaseRepository[T]) FindByID(ctx context.Context, id uuid.UUID) (*T, error) {
	query, args, err := r.builder.
		Select(r.columns...).
		From(r.tableName).
		Where(squirrel.Eq{"id": id}).
		Where(squirrel.Eq{"deleted_at": 0}).
		ToSql()
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrRepoDatabase, "failed to build query")
	}
	row := r.pool.QueryRow(ctx, query, args...)
	entity, err := r.scanner(row)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrRepoNotFound,
			fmt.Sprintf("%s not found: %s", r.tableName, id))
	}
	return entity, nil
}

// Delete performs a soft delete by setting deleted_at to the current unix timestamp.
func (r *BaseRepository[T]) Delete(ctx context.Context, id uuid.UUID) error {
	query, args, err := r.builder.
		Update(r.tableName).
		Set("deleted_at", squirrel.Expr("EXTRACT(EPOCH FROM NOW())::bigint")).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrRepoDatabase, "failed to build delete query")
	}
	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrRepoDatabase, "failed to execute soft delete")
	}
	return nil
}

// List retrieves paginated entities, excluding soft-deleted.
func (r *BaseRepository[T]) List(ctx context.Context, filter domain.Pagination) ([]*T, int64, error) {
	// Count
	countQuery, countArgs, err := r.builder.
		Select("COUNT(*)").
		From(r.tableName).
		Where(squirrel.Eq{"deleted_at": 0}).
		ToSql()
	if err != nil {
		return nil, 0, apperrors.Wrap(err, apperrors.ErrRepoDatabase, "failed to build count query")
	}
	var total int64
	err = r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, apperrors.ErrRepoDatabase, "failed to count")
	}

	// Data
	qb := r.builder.
		Select(r.columns...).
		From(r.tableName).
		Where(squirrel.Eq{"deleted_at": 0}).
		Limit(uint64(filter.Limit)).
		Offset(uint64(filter.Offset))

	if filter.SortBy != "" {
		order := "ASC"
		if filter.SortOrder == "DESC" {
			order = "DESC"
		}
		qb = qb.OrderBy(fmt.Sprintf("%s %s", filter.SortBy, order))
	}

	query, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.Wrap(err, apperrors.ErrRepoDatabase, "failed to build list query")
	}
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, apperrors.ErrRepoDatabase, "failed to query list")
	}
	defer rows.Close()

	var entities []*T
	for rows.Next() {
		entity, err := r.scanner(rows)
		if err != nil {
			return nil, 0, apperrors.Wrap(err, apperrors.ErrRepoDatabase, "failed to scan row")
		}
		entities = append(entities, entity)
	}
	return entities, total, nil
}
