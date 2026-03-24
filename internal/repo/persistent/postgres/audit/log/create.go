package log

import (
	"context"

	"gct/internal/shared/domain/consts"
	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

func (r *Repo) Create(ctx context.Context, a *domain.AuditLog) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns(
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
		).
		Values(
			a.UserID,
			a.SessionID,
			a.Action,
			a.ResourceType,
			a.ResourceID,
			a.Platform,
			a.IPAddress,
			a.UserAgent,
			a.Permission,
			a.PolicyID,
			a.Decision,
			a.Success,
			a.ErrorMessage,
			a.Metadata,
			a.CreatedAt,
		).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}
