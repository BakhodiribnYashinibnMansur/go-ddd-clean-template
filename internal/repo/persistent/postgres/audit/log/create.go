package log

import (
	"context"

	"gct/consts"
	"gct/internal/domain"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Create(ctx context.Context, a *domain.AuditLog) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns(
			schema.AuditLogUserID,
			schema.AuditLogSessionID,
			schema.AuditLogAction,
			schema.AuditLogResourceType,
			schema.AuditLogResourceID,
			schema.AuditLogPlatform,
			schema.AuditLogIPAddress,
			schema.AuditLogUserAgent,
			schema.AuditLogPermission,
			schema.AuditLogPolicyID,
			schema.AuditLogDecision,
			schema.AuditLogSuccess,
			schema.AuditLogErrorMessage,
			schema.AuditLogMetadata,
			schema.AuditLogCreatedAt,
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
