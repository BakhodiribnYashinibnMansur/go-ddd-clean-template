package permission

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) GetScopes(ctx context.Context, permID uuid.UUID) ([]*domain.Scope, error) {
	sql, args, err := r.builder.
		Select("s."+"path", "s."+"method", "s."+"created_at").
		From("scope"+" s").
		Join("permission_scope"+" ps ON s."+"path"+" = ps."+"path"+" AND s."+"method"+" = ps."+"method").
		Where(squirrel.Eq{"ps." + "permission_id": permID}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build select query")
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, "scope", nil)
	}
	defer rows.Close()

	var scopes []*domain.Scope
	for rows.Next() {
		var s domain.Scope
		if err := rows.Scan(&s.Path, &s.Method, &s.CreatedAt); err != nil {
			return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to scan row")
		}
		scopes = append(scopes, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.HandlePgError(err, "scope", nil)
	}

	return scopes, nil
}
