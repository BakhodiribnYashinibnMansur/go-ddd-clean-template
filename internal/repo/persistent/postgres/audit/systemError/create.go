package systemerror

import (
	"context"

	"gct/consts"
	"gct/internal/domain"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Create(ctx context.Context, e *domain.SystemError) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns(
			schema.SystemErrorCode,
			schema.SystemErrorMessage,
			schema.SystemErrorStackTrace,
			schema.SystemErrorMetadata,
			schema.SystemErrorSeverity,
			schema.SystemErrorServiceName,
			schema.SystemErrorRequestID,
			schema.SystemErrorUserID,
			schema.SystemErrorIPAddress,
			schema.SystemErrorPath,
			schema.SystemErrorMethod,
			schema.SystemErrorCreatedAt,
		).
		Values(
			e.Code,
			e.Message,
			e.StackTrace,
			e.Metadata,
			e.Severity,
			e.ServiceName,
			e.RequestID,
			e.UserID,
			e.IPAddress,
			e.Path,
			e.Method,
			e.CreatedAt,
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
