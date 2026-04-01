package postgres

import (
	"context"
	"fmt"
	"time"

	"gct/internal/audit/domain"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/metadata"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// readAuditLogColumns are the columns selected for audit log read-model queries.
var readAuditLogColumns = []string{
	"id", "user_id", "session_id", "action", "resource_type", "resource_id",
	"platform", "ip_address::text", "user_agent", "permission", "policy_id",
	"decision", "success", "error_message", "created_at",
}

// readEndpointHistoryColumns are the columns selected for endpoint history read-model queries.
var readEndpointHistoryColumns = []string{
	"id", "user_id", "path", "method", "status_code", "duration_ms",
	"ip_address::text", "user_agent", "created_at",
}

// AuditReadRepo implements domain.AuditReadRepository for the CQRS read side.
type AuditReadRepo struct {
	pool     *pgxpool.Pool
	builder  squirrel.StatementBuilderType
	metadata *metadata.GenericMetadataRepo
}

// NewAuditReadRepo creates a new AuditReadRepo.
func NewAuditReadRepo(pool *pgxpool.Pool) *AuditReadRepo {
	return &AuditReadRepo{
		pool:     pool,
		builder:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		metadata: metadata.NewGenericMetadataRepo(pool),
	}
}

// ListAuditLogs returns a paginated list of audit log views with optional filters.
func (r *AuditReadRepo) ListAuditLogs(ctx context.Context, filter domain.AuditLogFilter) (items []*domain.AuditLogView, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "AuditReadRepo.ListAuditLogs")
	defer func() { end(err) }()

	conds := squirrel.And{}

	if filter.UserID != nil {
		conds = append(conds, squirrel.Eq{"user_id": *filter.UserID})
	}
	if filter.Action != nil {
		conds = append(conds, squirrel.Eq{"action": string(*filter.Action)})
	}
	if filter.ResourceType != nil {
		conds = append(conds, squirrel.Eq{"resource_type": *filter.ResourceType})
	}
	if filter.ResourceID != nil {
		conds = append(conds, squirrel.Eq{"resource_id": *filter.ResourceID})
	}
	if filter.Success != nil {
		conds = append(conds, squirrel.Eq{"success": *filter.Success})
	}
	if filter.FromDate != nil {
		conds = append(conds, squirrel.GtOrEq{"created_at": *filter.FromDate})
	}
	if filter.ToDate != nil {
		conds = append(conds, squirrel.LtOrEq{"created_at": *filter.ToDate})
	}

	// Count total.
	countQB := r.builder.Select("COUNT(*)").From(consts.TableAuditLog)
	if len(conds) > 0 {
		countQB = countQB.Where(conds)
	}

	countSQL, countArgs, err := countQB.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	if err = r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.HandlePgError(err, consts.TableAuditLog, nil)
	}

	// Fetch page.
	qb := r.builder.Select(readAuditLogColumns...).From(consts.TableAuditLog)
	if len(conds) > 0 {
		qb = qb.Where(conds)
	}

	if filter.Pagination != nil {
		qb = qb.Limit(uint64(filter.Pagination.Limit)).
			Offset(uint64(filter.Pagination.Offset))

		if filter.Pagination.SortBy != "" {
			order := "ASC"
			if filter.Pagination.SortOrder == "DESC" {
				order = "DESC"
			}
			qb = qb.OrderBy(fmt.Sprintf("%s %s", filter.Pagination.SortBy, order))
		}
	}

	// Default ordering by created_at DESC if no sort specified.
	if filter.Pagination == nil || filter.Pagination.SortBy == "" {
		qb = qb.OrderBy("created_at DESC")
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, consts.TableAuditLog, nil)
	}
	defer rows.Close()

	var views []*domain.AuditLogView
	for rows.Next() {
		var (
			id           uuid.UUID
			userID       *uuid.UUID
			sessionID    *uuid.UUID
			action       string
			resourceType *string
			resourceID   *uuid.UUID
			platform     *string
			ipAddress    *string
			userAgent    *string
			permission   *string
			policyID     *uuid.UUID
			decision     *string
			success      bool
			errorMessage *string
			createdAt    time.Time
		)

		if err := rows.Scan(
			&id, &userID, &sessionID, &action, &resourceType, &resourceID,
			&platform, &ipAddress, &userAgent, &permission, &policyID,
			&decision, &success, &errorMessage, &createdAt,
		); err != nil {
			return nil, 0, apperrors.HandlePgError(err, consts.TableAuditLog, nil)
		}

		meta, err := r.metadata.GetAll(ctx, metadata.EntityTypeAuditLogMetadata, id)
		if err != nil {
			return nil, 0, err
		}

		views = append(views, &domain.AuditLogView{
			ID:           id,
			UserID:       userID,
			SessionID:    sessionID,
			Action:       domain.AuditAction(action),
			ResourceType: resourceType,
			ResourceID:   resourceID,
			Platform:     platform,
			IPAddress:    ipAddress,
			UserAgent:    userAgent,
			Permission:   permission,
			PolicyID:     policyID,
			Decision:     decision,
			Success:      success,
			ErrorMessage: errorMessage,
			Metadata:     meta,
			CreatedAt:    createdAt,
		})
	}

	return views, total, nil
}

// ListEndpointHistory returns a paginated list of endpoint history views with optional filters.
func (r *AuditReadRepo) ListEndpointHistory(ctx context.Context, filter domain.EndpointHistoryFilter) (items []*domain.EndpointHistoryView, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "AuditReadRepo.ListEndpointHistory")
	defer func() { end(err) }()

	conds := squirrel.And{}

	if filter.UserID != nil {
		conds = append(conds, squirrel.Eq{"user_id": *filter.UserID})
	}
	if filter.Method != nil {
		conds = append(conds, squirrel.Eq{"method": *filter.Method})
	}
	if filter.Endpoint != nil {
		conds = append(conds, squirrel.Eq{"path": *filter.Endpoint})
	}
	if filter.StatusCode != nil {
		conds = append(conds, squirrel.Eq{"status_code": *filter.StatusCode})
	}
	if filter.FromDate != nil {
		conds = append(conds, squirrel.GtOrEq{"created_at": *filter.FromDate})
	}
	if filter.ToDate != nil {
		conds = append(conds, squirrel.LtOrEq{"created_at": *filter.ToDate})
	}

	// Count total.
	countQB := r.builder.Select("COUNT(*)").From(consts.TableEndpointHistory)
	if len(conds) > 0 {
		countQB = countQB.Where(conds)
	}

	countSQL, countArgs, err := countQB.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	if err = r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.HandlePgError(err, consts.TableEndpointHistory, nil)
	}

	// Fetch page.
	qb := r.builder.Select(readEndpointHistoryColumns...).From(consts.TableEndpointHistory)
	if len(conds) > 0 {
		qb = qb.Where(conds)
	}

	if filter.Pagination != nil {
		qb = qb.Limit(uint64(filter.Pagination.Limit)).
			Offset(uint64(filter.Pagination.Offset))

		if filter.Pagination.SortBy != "" {
			order := "ASC"
			if filter.Pagination.SortOrder == "DESC" {
				order = "DESC"
			}
			qb = qb.OrderBy(fmt.Sprintf("%s %s", filter.Pagination.SortBy, order))
		}
	}

	// Default ordering by created_at DESC if no sort specified.
	if filter.Pagination == nil || filter.Pagination.SortBy == "" {
		qb = qb.OrderBy("created_at DESC")
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, consts.TableEndpointHistory, nil)
	}
	defer rows.Close()

	var views []*domain.EndpointHistoryView
	for rows.Next() {
		var (
			id         uuid.UUID
			userID     *uuid.UUID
			endpoint   string
			method     string
			statusCode int
			latency    int
			ipAddress  *string
			userAgent  *string
			createdAt  time.Time
		)

		if err := rows.Scan(
			&id, &userID, &endpoint, &method, &statusCode, &latency,
			&ipAddress, &userAgent, &createdAt,
		); err != nil {
			return nil, 0, apperrors.HandlePgError(err, consts.TableEndpointHistory, nil)
		}

		views = append(views, &domain.EndpointHistoryView{
			ID:         id,
			UserID:     userID,
			Endpoint:   endpoint,
			Method:     method,
			StatusCode: statusCode,
			Latency:    latency,
			IPAddress:  ipAddress,
			UserAgent:  userAgent,
			CreatedAt:  createdAt,
		})
	}

	return views, total, nil
}
