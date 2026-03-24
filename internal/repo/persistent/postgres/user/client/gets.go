package client

import (
	"context"

	"gct/internal/shared/domain/consts"
	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Gets(ctx context.Context, filter *domain.UsersFilter) ([]*domain.User, int, error) {
	// Base query
	baseQb := r.builder.Select(
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
			qb = qb.OrderBy("created_at" + " DESC")
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

	// Fetch attributes for all users (batch)
	// This block is no longer needed as attributes are fetched directly in the main query
	// if len(users) > 0 {
	// 	userIDs := make([]any, len(users))
	// 	userMap := make(map[string]*domain.User)
	// 	for i, u := range users {
	// 		userIDs[i] = u.ID
	// 		userMap[u.ID.String()] = u
	// 		u.Attributes = make(map[string]any)
	// 	}

	// 	attrSql, attrArgs, err := r.builder.Select("user_id", "key", "value").
	// 		From("user_attributes").
	// 		Where(squirrel.Eq{"user_id": userIDs}).
	// 		ToSql()

	// 	if err == nil {
	// 		rows, err := r.pool.Query(ctx, attrSql, attrArgs...)
	// 		if err == nil {
	// 			defer rows.Close()
	// 			for rows.Next() {
	// 				var uid, key, value string
	// 				if err := rows.Scan(&uid, &key, &value); err == nil {
	// 					if u, ok := userMap[uid]; ok {
	// 						u.Attributes[key] = value
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}
	// }

	if err := rows.Err(); err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}

	// Count Query
	// Start fresh from builder but reuse filter logic
	countBaseQb := r.builder.Select("COUNT(*)").From(tableName).Where("deleted_at" + " = 0")
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
	return qb
}
