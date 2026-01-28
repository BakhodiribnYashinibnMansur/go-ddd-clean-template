package client

import (
	"context"

	"gct/consts"
	"gct/internal/domain"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Gets(ctx context.Context, filter *domain.UsersFilter) ([]*domain.User, int, error) {
	// Base query
	baseQb := r.builder.Select(
		schema.UsersID,
		schema.UsersRoleID,
		schema.UsersUsername,
		schema.UsersEmail,
		schema.UsersPhone,
		schema.UsersPasswordHash,
		schema.UsersSalt,
		schema.UsersAttributes,
		schema.UsersActive,
		schema.UsersIsApproved,
		schema.UsersCreatedAt,
		schema.UsersUpdatedAt,
		schema.UsersDeletedAt,
		schema.UsersLastSeen,
	).
		From(tableName).
		Where(schema.UsersDeletedAt + " = 0")

	// Apply dynamic filters
	qb := r.applyFilters(baseQb, &filter.UserFilter)

	// Apply Sorting & Pagination
	if filter.Pagination != nil {
		if filter.Pagination.SortBy != "" {
			order := consts.SQLOrderAsc
			if filter.Pagination.SortOrder == consts.SQLOrderDesc {
				order = consts.SQLOrderDesc
			}
			qb = qb.OrderBy(filter.Pagination.SortBy + " " + order)
		} else {
			qb = qb.OrderBy(schema.UsersCreatedAt + " DESC")
		}
		qb = qb.Limit(uint64(filter.Pagination.Limit)).Offset(uint64(filter.Pagination.Offset))
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

	var users []*domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.ID, &u.RoleID, &u.Username, &u.Email, &u.Phone, &u.PasswordHash, &u.Salt,
			&u.Attributes, &u.Active, &u.IsApproved,
			&u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &u.LastSeen,
		); err != nil {
			return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToScanRow)
		}
		users = append(users, &u)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}

	// Count Query
	// Start fresh from builder but reuse filter logic
	countBaseQb := r.builder.Select("COUNT(*)").From(tableName).Where(schema.UsersDeletedAt + " = 0")
	countQb := r.applyFilters(countBaseQb, &filter.UserFilter)

	countSql, countArgs, err := countQb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}
	var count int
	if err := r.pool.QueryRow(ctx, countSql, countArgs...).Scan(&count); err != nil {
		return nil, 0, apperrors.HandlePgError(err, consts.TableUsers, nil)
	}

	return users, count, nil
}

func (r *Repo) applyFilters(qb squirrel.SelectBuilder, filter *domain.UserFilter) squirrel.SelectBuilder {
	if filter.ID != nil {
		qb = qb.Where(squirrel.Eq{schema.UsersID: *filter.ID})
	}
	if filter.RoleID != nil {
		qb = qb.Where(squirrel.Eq{schema.UsersRoleID: *filter.RoleID})
	}
	if filter.Username != nil {
		qb = qb.Where(squirrel.Eq{schema.UsersUsername: *filter.Username})
	}
	if filter.Phone != nil {
		qb = qb.Where(squirrel.Eq{schema.UsersPhone: *filter.Phone})
	}
	if filter.Email != nil {
		qb = qb.Where(squirrel.Eq{schema.UsersEmail: *filter.Email})
	}
	if filter.Active != nil {
		qb = qb.Where(squirrel.Eq{schema.UsersActive: *filter.Active})
	}
	if filter.IsApproved != nil {
		qb = qb.Where(squirrel.Eq{schema.UsersIsApproved: *filter.IsApproved})
	}
	return qb
}
