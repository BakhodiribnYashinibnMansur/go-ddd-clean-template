package postgres

import (
	"context"
	"time"

	"gct/internal/context/iam/generic/session/application/dto"
	"gct/internal/kernel/consts"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// readSessionColumns are the columns selected for read-model queries.
var readSessionColumns = []string{
	"id", "user_id", "device_id", "device_name", "device_type",
	"ip_address::text", "user_agent", "expires_at", "last_activity",
	"revoked", "created_at",
}

// SessionReadRepo implements query.SessionReadRepository for the CQRS read side.
type SessionReadRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewSessionReadRepo creates a new SessionReadRepo.
func NewSessionReadRepo(pool *pgxpool.Pool) *SessionReadRepo {
	return &SessionReadRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// FindByID returns a SessionView for the given session ID.
func (r *SessionReadRepo) FindByID(ctx context.Context, id uuid.UUID) (result *dto.SessionView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "SessionReadRepo.FindByID")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Select(readSessionColumns...).
		From(consts.TableSession).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)

	var (
		sid          uuid.UUID
		userID       uuid.UUID
		deviceID     string
		deviceName   string
		deviceType   string
		ipAddress    string
		userAgent    string
		expiresAt    time.Time
		lastActivity time.Time
		revoked      bool
		createdAt    time.Time
	)

	err = row.Scan(
		&sid, &userID, &deviceID, &deviceName, &deviceType,
		&ipAddress, &userAgent, &expiresAt, &lastActivity,
		&revoked, &createdAt,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableSession, map[string]any{"id": id})
	}

	return &dto.SessionView{
		ID:           sid,
		UserID:       userID,
		DeviceID:     deviceID,
		DeviceName:   deviceName,
		DeviceType:   deviceType,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		ExpiresAt:    expiresAt,
		LastActivity: lastActivity,
		Revoked:      revoked,
		CreatedAt:    createdAt,
	}, nil
}

// List returns a paginated list of SessionView with optional filters.
func (r *SessionReadRepo) List(ctx context.Context, filter dto.SessionsFilter) (items []*dto.SessionView, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "SessionReadRepo.List")
	defer func() { end(err) }()

	// Build WHERE conditions.
	conds := squirrel.And{}
	if filter.UserID != nil {
		conds = append(conds, squirrel.Eq{"user_id": *filter.UserID})
	}
	if filter.Revoked != nil {
		conds = append(conds, squirrel.Eq{"revoked": *filter.Revoked})
	}

	// Count total.
	countQB := r.builder.Select("COUNT(*)").From(consts.TableSession)
	if len(conds) > 0 {
		countQB = countQB.Where(conds)
	}
	countSQL, countArgs, err := countQB.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	if err = r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.HandlePgError(err, consts.TableSession, nil)
	}

	// Fetch page.
	qb := r.builder.
		Select(readSessionColumns...).
		From(consts.TableSession).
		OrderBy("created_at DESC")

	if len(conds) > 0 {
		qb = qb.Where(conds)
	}

	if filter.Limit > 0 {
		qb = qb.Limit(uint64(filter.Limit))
	}
	if filter.Offset > 0 {
		qb = qb.Offset(uint64(filter.Offset))
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, consts.TableSession, nil)
	}
	defer rows.Close()

	var views []*dto.SessionView
	for rows.Next() {
		var (
			sid          uuid.UUID
			userID       uuid.UUID
			deviceID     string
			deviceName   string
			deviceType   string
			ipAddress    string
			userAgent    string
			expiresAt    time.Time
			lastActivity time.Time
			revoked      bool
			createdAt    time.Time
		)

		if err := rows.Scan(
			&sid, &userID, &deviceID, &deviceName, &deviceType,
			&ipAddress, &userAgent, &expiresAt, &lastActivity,
			&revoked, &createdAt,
		); err != nil {
			return nil, 0, apperrors.HandlePgError(err, consts.TableSession, nil)
		}

		views = append(views, &dto.SessionView{
			ID:           sid,
			UserID:       userID,
			DeviceID:     deviceID,
			DeviceName:   deviceName,
			DeviceType:   deviceType,
			IPAddress:    ipAddress,
			UserAgent:    userAgent,
			ExpiresAt:    expiresAt,
			LastActivity: lastActivity,
			Revoked:      revoked,
			CreatedAt:    createdAt,
		})
	}

	return views, total, nil
}
