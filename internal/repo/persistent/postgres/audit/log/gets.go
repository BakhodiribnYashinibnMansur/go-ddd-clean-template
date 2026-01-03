package log

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Gets(ctx context.Context, filter *domain.AuditLogsFilter) ([]*domain.AuditLog, int, error) {
	query := r.builder.Select(
		"id",
		"user_id",
		"session_id",
		"action",
		"resource_type",
		"resource_id",
		"platform",
		"ip_address",
		"user_agent",
		"permission",
		"policy_id",
		"decision",
		"success",
		"error_message",
		"metadata",
		"created_at",
	).From(tableName)

	if filter.UserID != nil {
		query = query.Where(squirrel.Eq{"user_id": filter.UserID})
	}
	if filter.Action != nil {
		query = query.Where(squirrel.Eq{"action": filter.Action})
	}
	if filter.ResourceType != nil {
		query = query.Where(squirrel.Eq{"resource_type": filter.ResourceType})
	}
	if filter.ResourceID != nil {
		query = query.Where(squirrel.Eq{"resource_id": filter.ResourceID})
	}
	if filter.Success != nil {
		query = query.Where(squirrel.Eq{"success": filter.Success})
	}
	if filter.FromDate != nil {
		query = query.Where(squirrel.GtOrEq{"created_at": filter.FromDate})
	}
	if filter.ToDate != nil {
		query = query.Where(squirrel.LtOrEq{"created_at": filter.ToDate})
	}

	// Count total
	countQuery := r.builder.Select("COUNT(*)").From(tableName)
	// Apply same filters to count
	if filter.UserID != nil {
		countQuery = countQuery.Where(squirrel.Eq{"user_id": filter.UserID})
	}
	if filter.Action != nil {
		countQuery = countQuery.Where(squirrel.Eq{"action": filter.Action})
	}
	if filter.ResourceType != nil {
		countQuery = countQuery.Where(squirrel.Eq{"resource_type": filter.ResourceType})
	}
	if filter.ResourceID != nil {
		countQuery = countQuery.Where(squirrel.Eq{"resource_id": filter.ResourceID})
	}
	if filter.Success != nil {
		countQuery = countQuery.Where(squirrel.Eq{"success": filter.Success})
	}
	if filter.FromDate != nil {
		countQuery = countQuery.Where(squirrel.GtOrEq{"created_at": filter.FromDate})
	}
	if filter.ToDate != nil {
		countQuery = countQuery.Where(squirrel.LtOrEq{"created_at": filter.ToDate})
	}

	sql, args, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to build count query")
	}

	var count int
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(ctx, err, tableName, nil)
	}

	// Apply pagination
	if filter.Pagination != nil {
		if filter.Pagination.Limit > 0 {
			query = query.Limit(uint64(filter.Pagination.Limit))
		}
		if filter.Pagination.Offset > 0 {
			query = query.Offset(uint64(filter.Pagination.Offset))
		}
	}

	// Order by created_at desc
	query = query.OrderBy("created_at DESC")

	sql, args, err = query.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to build select query")
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(ctx, err, tableName, nil)
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		var l domain.AuditLog
		// Scan nullable fields properly.
		// Actually, let's scan directly into the struct pointers if they are nullable in DB AND the struct field is a pointer.
		// My domain model uses pointers for nullable fields, which pgx supports.
		err = rows.Scan(
			&l.ID,
			&l.UserID,
			&l.SessionID,
			&l.Action,
			&l.ResourceType,
			&l.ResourceID,
			&l.Platform,
			&l.IPAddress,
			&l.UserAgent,
			&l.Permission,
			&l.PolicyID,
			&l.Decision,
			&l.Success,
			&l.ErrorMessage,
			&l.Metadata,
			&l.CreatedAt,
		)
		if err != nil {
			return nil, 0, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, fmt.Sprintf("failed to scan row: %s", err))
		}
		logs = append(logs, &l)
	}

	return logs, count, nil
}
