package client

import (
	"context"

	"gct/internal/shared/domain/consts"
	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Get(ctx context.Context, filter *domain.UserFilter) (*domain.User, error) {
	qb := r.builder.
		Select(
			"id",
			"role_id",
			"username",
			"email",
			"phone",
			"password_hash",
			"salt",
			"attributes",
			"active",
			"is_approved",
			"created_at",
			"updated_at",
			"deleted_at",
			"last_seen",
		).
		From(tableName).
		Where("deleted_at" + " = 0")

	if filter.ID != nil {
		qb = qb.Where(squirrel.Eq{"id": *filter.ID})
	}

	if filter.RoleID != nil {
		qb = qb.Where(squirrel.Eq{"role_id": *filter.RoleID})
	}

	if filter.Username != nil {
		qb = qb.Where(squirrel.Eq{"username": *filter.Username})
	}

	if filter.Phone != nil {
		qb = qb.Where(squirrel.Eq{"phone": *filter.Phone})
	}

	if filter.Email != nil {
		qb = qb.Where(squirrel.Eq{"email": *filter.Email})
	}

	if filter.Active != nil {
		qb = qb.Where(squirrel.Eq{"active": *filter.Active})
	}

	if filter.IsApproved != nil {
		qb = qb.Where(squirrel.Eq{"is_approved": *filter.IsApproved})
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase,
			consts.ErrMsgFailedToBuildQuery)
	}

	var u domain.User
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&u.ID, &u.RoleID, &u.Username, &u.Email, &u.Phone, &u.PasswordHash, &u.Salt,
		&u.Attributes, &u.Active, &u.IsApproved,
		&u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &u.LastSeen,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, map[string]any{
			"filter": filter,
		})
	}

	// Fetch relations
	relSql, relArgs, err := r.builder.
		Select("r."+"id", "r."+"type", "r."+"name", "r."+"created_at").
		From("relation"+" r").
		Join("user_relation"+" ur ON r."+"id"+" = ur."+"relation_id").
		Where(squirrel.Eq{"ur." + "user_id": u.ID}).
		ToSql()
	if err == nil {
		rows, err := r.pool.Query(ctx, relSql, relArgs...)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var rel domain.Relation
				if err := rows.Scan(&rel.ID, &rel.Type, &rel.Name, &rel.CreatedAt); err == nil {
					u.Relations = append(u.Relations, rel)
				}
			}
		}
	}

	return &u, nil
}
