package postgres

import (
	"context"
	"encoding/json"

	"gct/internal/audit/domain"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

// auditLogColumns are the columns for the audit_log table.
var auditLogColumns = []string{
	"id", "user_id", "session_id", "action", "resource_type", "resource_id",
	"platform", "ip_address", "user_agent", "permission", "policy_id",
	"decision", "success", "error_message", "metadata", "created_at",
}

// endpointHistoryColumns are the columns for the endpoint_history table.
var endpointHistoryColumns = []string{
	"id", "user_id", "path", "method", "status_code", "duration_ms",
	"ip_address", "user_agent", "created_at",
}

// AuditLogWriteRepo implements domain.AuditLogRepository using PostgreSQL.
type AuditLogWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewAuditLogWriteRepo creates a new AuditLogWriteRepo.
func NewAuditLogWriteRepo(pool *pgxpool.Pool) *AuditLogWriteRepo {
	return &AuditLogWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new audit log entry. Audit logs are immutable.
func (r *AuditLogWriteRepo) Save(ctx context.Context, auditLog *domain.AuditLog) error {
	metadataJSON, err := json.Marshal(auditLog.Metadata())
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToMarshalJSON)
	}

	sql, args, err := r.builder.
		Insert(consts.TableAuditLog).
		Columns(auditLogColumns...).
		Values(
			auditLog.ID(),
			auditLog.UserID(),
			auditLog.SessionID(),
			string(auditLog.Action()),
			auditLog.ResourceType(),
			auditLog.ResourceID(),
			auditLog.Platform(),
			auditLog.IPAddress(),
			auditLog.UserAgent(),
			auditLog.Permission(),
			auditLog.PolicyID(),
			auditLog.Decision(),
			auditLog.Success(),
			auditLog.ErrorMessage(),
			metadataJSON,
			auditLog.CreatedAt(),
		).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, consts.TableAuditLog, nil)
	}

	return nil
}

// EndpointHistoryWriteRepo implements domain.EndpointHistoryRepository using PostgreSQL.
type EndpointHistoryWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewEndpointHistoryWriteRepo creates a new EndpointHistoryWriteRepo.
func NewEndpointHistoryWriteRepo(pool *pgxpool.Pool) *EndpointHistoryWriteRepo {
	return &EndpointHistoryWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new endpoint history entry. Endpoint history entries are immutable.
func (r *EndpointHistoryWriteRepo) Save(ctx context.Context, entry *domain.EndpointHistory) error {
	sql, args, err := r.builder.
		Insert(consts.TableEndpointHistory).
		Columns(endpointHistoryColumns...).
		Values(
			entry.ID(),
			entry.UserID(),
			entry.Endpoint(),
			entry.Method(),
			entry.StatusCode(),
			entry.Latency(),
			entry.IPAddress(),
			entry.UserAgent(),
			entry.GetCreatedAt(),
		).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, consts.TableEndpointHistory, nil)
	}

	return nil
}
