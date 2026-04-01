package metadata

import (
	"context"
	"time"

	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GenericMetadataRepo provides CRUD for the entity_metadata EAV table.
// It is not a domain interface — BCs compose it as a private field in their infra repos.
type GenericMetadataRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewGenericMetadataRepo creates a new GenericMetadataRepo.
func NewGenericMetadataRepo(pool *pgxpool.Pool) *GenericMetadataRepo {
	return &GenericMetadataRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// SetMany upserts multiple key-value pairs for a given entity.
func (r *GenericMetadataRepo) SetMany(ctx context.Context, entityType string, entityID uuid.UUID, entries map[string]string) error {
	if len(entries) == 0 {
		return nil
	}

	now := time.Now()
	qb := r.builder.
		Insert(consts.TableEntityMetadata).
		Columns("entity_type", "entity_id", "key", "value", "created_at", "updated_at")

	for k, v := range entries {
		qb = qb.Values(entityType, entityID, k, v, now, now)
	}

	qb = qb.Suffix("ON CONFLICT (entity_type, entity_id, key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at")

	sql, args, err := qb.ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, consts.TableEntityMetadata, nil)
	}

	return nil
}

// SetManyTx upserts multiple key-value pairs within an existing transaction.
func (r *GenericMetadataRepo) SetManyTx(ctx context.Context, tx pgx.Tx, entityType string, entityID uuid.UUID, entries map[string]string) error {
	if len(entries) == 0 {
		return nil
	}

	now := time.Now()
	qb := r.builder.
		Insert(consts.TableEntityMetadata).
		Columns("entity_type", "entity_id", "key", "value", "created_at", "updated_at")

	for k, v := range entries {
		qb = qb.Values(entityType, entityID, k, v, now, now)
	}

	qb = qb.Suffix("ON CONFLICT (entity_type, entity_id, key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at")

	sql, args, err := qb.ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = tx.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, consts.TableEntityMetadata, nil)
	}

	return nil
}

// GetAll retrieves all key-value pairs for an entity.
func (r *GenericMetadataRepo) GetAll(ctx context.Context, entityType string, entityID uuid.UUID) (map[string]string, error) {
	sql, args, err := r.builder.
		Select("key", "value").
		From(consts.TableEntityMetadata).
		Where(squirrel.Eq{"entity_type": entityType, "entity_id": entityID}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableEntityMetadata, nil)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, apperrors.HandlePgError(err, consts.TableEntityMetadata, nil)
		}
		result[k] = v
	}

	return result, nil
}

// GetAllTx retrieves all key-value pairs within an existing transaction.
func (r *GenericMetadataRepo) GetAllTx(ctx context.Context, tx pgx.Tx, entityType string, entityID uuid.UUID) (map[string]string, error) {
	sql, args, err := r.builder.
		Select("key", "value").
		From(consts.TableEntityMetadata).
		Where(squirrel.Eq{"entity_type": entityType, "entity_id": entityID}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableEntityMetadata, nil)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, apperrors.HandlePgError(err, consts.TableEntityMetadata, nil)
		}
		result[k] = v
	}

	return result, nil
}

// DeleteAll removes all metadata for an entity.
func (r *GenericMetadataRepo) DeleteAll(ctx context.Context, entityType string, entityID uuid.UUID) error {
	sql, args, err := r.builder.
		Delete(consts.TableEntityMetadata).
		Where(squirrel.Eq{"entity_type": entityType, "entity_id": entityID}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, consts.TableEntityMetadata, nil)
	}

	return nil
}

// DeleteAllTx removes all metadata within an existing transaction.
func (r *GenericMetadataRepo) DeleteAllTx(ctx context.Context, tx pgx.Tx, entityType string, entityID uuid.UUID) error {
	sql, args, err := r.builder.
		Delete(consts.TableEntityMetadata).
		Where(squirrel.Eq{"entity_type": entityType, "entity_id": entityID}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
	}

	if _, err = tx.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, consts.TableEntityMetadata, nil)
	}

	return nil
}
