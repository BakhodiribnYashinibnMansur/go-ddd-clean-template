package session

import (
	"context"

	"gct/consts"
	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

func (r *Repo) Gets(ctx context.Context, filter *domain.SessionsFilter) ([]*domain.Session, int, error) {
	qb := r.buildSelectSessionsQuery(filter)
	countQb := r.buildCountSessionsQuery(filter)

	count, err := r.getTotalCount(ctx, countQb)
	if err != nil {
		return nil, 0, err
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, map[string]any{"operation": "get_sessions"})
	}
	defer rows.Close()

	sessions, err := r.scanSessionRows(rows)
	if err != nil {
		return nil, 0, err
	}

	return sessions, count, nil
}

func (r *Repo) buildSelectSessionsQuery(filter *domain.SessionsFilter) squirrel.SelectBuilder {
	qb := r.builder.
		Select(
			"id",
			"device_id",
			"device_name",
			"device_type",
			"ip_address"+"::text",
			"user_agent",
			"fcm_token",
			"refresh_token_hash",
			// "data",
			"user_id",
			"expires_at",
			"last_activity",
			"revoked",
			"created_at",
			"updated_at",
		).
		From(tableName)

	if !filter.IsIDNull() {
		qb = qb.Where(squirrel.Eq{"id": *filter.ID})
	}
	if !filter.IsUserIDNull() {
		qb = qb.Where(squirrel.Eq{"user_id": *filter.UserID})
	}
	if !filter.IsRevokedNull() {
		qb = qb.Where(squirrel.Eq{"revoked": *filter.Revoked})
	}

	if filter.IsValidLimit() {
		qb = qb.Limit(uint64(filter.Pagination.Limit))
	}
	if filter.IsValidOffset() {
		qb = qb.Offset(uint64(filter.Pagination.Offset))
	}

	// Default sort by created_at DESC if not specified (or always for now)
	if filter.IsPaginationNull() || filter.Pagination.SortBy == "" {
		qb = qb.OrderBy("created_at" + " DESC")
	} else {
		// Handle dynamic sort if needed, but for now fallback/default to created_at DESC
		// to ensure consistent latest-first view
		qb = qb.OrderBy(filter.Pagination.SortBy + " " + filter.Pagination.SortOrder)
	}

	return qb
}

func (r *Repo) buildCountSessionsQuery(filter *domain.SessionsFilter) squirrel.SelectBuilder {
	countQb := r.builder.Select("COUNT(*)").From(tableName)
	if !filter.IsIDNull() {
		countQb = countQb.Where(squirrel.Eq{"id": *filter.ID})
	}
	if !filter.IsUserIDNull() {
		countQb = countQb.Where(squirrel.Eq{"user_id": *filter.UserID})
	}
	if !filter.IsRevokedNull() {
		countQb = countQb.Where(squirrel.Eq{"revoked": *filter.Revoked})
	}
	return countQb
}

func (r *Repo) getTotalCount(ctx context.Context, qb squirrel.SelectBuilder) (int, error) {
	sql, args, err := qb.ToSql()
	if err != nil {
		return 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	var count int
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, apperrors.HandlePgError(err, tableName, map[string]any{"operation": "count"})
	}
	return count, nil
}

func (r *Repo) scanSessionRows(rows pgx.Rows) ([]*domain.Session, error) {
	var sessions []*domain.Session
	for rows.Next() {
		var s domain.Session
		err := rows.Scan(
			&s.ID, &s.DeviceID, &s.DeviceName, &s.DeviceType, &s.IPAddress, &s.UserAgent,
			&s.FCMToken, &s.RefreshTokenHash,
			// &s.Data,
			&s.UserID, &s.ExpiresAt,
			&s.LastActivity, &s.Revoked, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, apperrors.HandlePgError(err, tableName, map[string]any{"operation": "scan_row"})
		}
		sessions = append(sessions, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.HandlePgError(err, tableName, map[string]any{"operation": "scan_rows"})
	}

	return sessions, nil
}
