package postgres

import (
	"context"

	"gct/internal/audit/domain"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/metadata"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

// auditLogColumns are the columns for the audit_log table.
var auditLogColumns = []string{
	"id", "user_id", "session_id", "action", "resource_type", "resource_id",
	"platform", "ip_address", "user_agent", "permission", "policy_id",
	"decision", "success", "error_message", "created_at",
}

// endpointHistoryColumns are the columns for the endpoint_history table.
var endpointHistoryColumns = []string{
	"id", "user_id", "path", "method", "status_code", "duration_ms",
	"ip_address", "user_agent", "created_at",
}

// AuditLogWriteRepo implements domain.AuditLogRepository using PostgreSQL.
type AuditLogWriteRepo struct {
	pool     *pgxpool.Pool
	builder  squirrel.StatementBuilderType
	metadata *metadata.GenericMetadataRepo
}

// NewAuditLogWriteRepo creates a new AuditLogWriteRepo.
func NewAuditLogWriteRepo(pool *pgxpool.Pool) *AuditLogWriteRepo {
	return &AuditLogWriteRepo{
		pool:     pool,
		builder:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		metadata: metadata.NewGenericMetadataRepo(pool),
	}
}

// Save inserts a new audit log entry. Audit logs are immutable.
func (r *AuditLogWriteRepo) Save(ctx context.Context, auditLog *domain.AuditLog) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "AuditLogWriteRepo.Save")
	defer func() { end(err) }()

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
			auditLog.CreatedAt(),
		).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, consts.TableAuditLog, nil)
	}

	return r.metadata.SetMany(ctx, metadata.EntityTypeAuditLogMetadata, auditLog.ID(), auditLog.Metadata())
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
func (r *EndpointHistoryWriteRepo) Save(ctx context.Context, entry *domain.EndpointHistory) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "EndpointHistoryWriteRepo.Save")
	defer func() { end(err) }()

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
