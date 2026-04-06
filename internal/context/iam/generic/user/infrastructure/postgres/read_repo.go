package postgres

import (
	"context"
	"fmt"
	"time"

	"gct/internal/context/iam/generic/user/domain"
	"gct/internal/kernel/consts"
	shared "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/metadata"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// readUserColumns are the columns selected for read-model queries.
var readUserColumns = []string{
	"id", "role_id", "username", "email", "phone",
	"active", "is_approved",
	"last_seen", "created_at", "updated_at",
}

// UserReadRepo implements domain.UserReadRepository for the CQRS read side.
type UserReadRepo struct {
	pool     *pgxpool.Pool
	builder  squirrel.StatementBuilderType
	metadata *metadata.GenericMetadataRepo
}

// NewUserReadRepo creates a new UserReadRepo.
func NewUserReadRepo(pool *pgxpool.Pool) *UserReadRepo {
	return &UserReadRepo{
		pool:     pool,
		builder:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		metadata: metadata.NewGenericMetadataRepo(pool),
	}
}

// FindByID returns a UserView for the given user ID.
func (r *UserReadRepo) FindByID(ctx context.Context, id domain.UserID) (result *domain.UserView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "UserReadRepo.FindByID")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Select(readUserColumns...).
		From(consts.TableUsers).
		Where(squirrel.Eq{"id": id.UUID()}).
		Where(squirrel.Eq{"deleted_at": 0}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)

	var (
		uid        uuid.UUID
		roleID     *uuid.UUID
		username   *string
		email      *string
		phone      string
		active     bool
		isApproved bool
		lastSeen   *time.Time
		createdAt  time.Time
		updatedAt  time.Time
	)

	err = row.Scan(
		&uid, &roleID, &username, &email, &phone,
		&active, &isApproved,
		&lastSeen, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableUsers, map[string]any{"id": id})
	}

	attrs, err := r.metadata.GetAll(ctx, metadata.EntityTypeUserAttributes, uid)
	if err != nil {
		return nil, err
	}

	return &domain.UserView{
		ID:         domain.UserID(uid),
		Phone:      phone,
		Email:      email,
		Username:   username,
		RoleID:     roleID,
		Attributes: attrs,
		Active:     active,
		IsApproved: isApproved,
	}, nil
}

// List returns a paginated list of UserView with optional filters.
func (r *UserReadRepo) List(ctx context.Context, filter domain.UsersFilter) (items []*domain.UserView, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "UserReadRepo.List")
	defer func() { end(err) }()

	// Build WHERE conditions.
	conds := squirrel.And{squirrel.Eq{"deleted_at": 0}}
	if filter.Phone != nil {
		conds = append(conds, squirrel.Eq{"phone": *filter.Phone})
	}
	if filter.Email != nil {
		conds = append(conds, squirrel.Eq{"email": *filter.Email})
	}
	if filter.Active != nil {
		conds = append(conds, squirrel.Eq{"active": *filter.Active})
	}
	if filter.IsApproved != nil {
		conds = append(conds, squirrel.Eq{"is_approved": *filter.IsApproved})
	}

	// Count total.
	countSQL, countArgs, err := r.builder.
		Select("COUNT(*)").
		From(consts.TableUsers).
		Where(conds).
		ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	if err = r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.HandlePgError(err, consts.TableUsers, nil)
	}

	// Fetch page.
	qb := r.builder.
		Select(readUserColumns...).
		From(consts.TableUsers).
		Where(conds)

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

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, consts.TableUsers, nil)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			uid        uuid.UUID
			roleID     *uuid.UUID
			username   *string
			email      *string
			phone      string
			active     bool
			isApproved bool
			lastSeen   *time.Time
			createdAt  time.Time
			updatedAt  time.Time
		)

		if err := rows.Scan(
			&uid, &roleID, &username, &email, &phone,
			&active, &isApproved,
			&lastSeen, &createdAt, &updatedAt,
		); err != nil {
			return nil, 0, apperrors.HandlePgError(err, consts.TableUsers, nil)
		}

		attrs, err := r.metadata.GetAll(ctx, metadata.EntityTypeUserAttributes, uid)
		if err != nil {
			return nil, 0, err
		}

		items = append(items, &domain.UserView{
			ID:         domain.UserID(uid),
			Phone:      phone,
			Email:      email,
			Username:   username,
			RoleID:     roleID,
			Attributes: attrs,
			Active:     active,
			IsApproved: isApproved,
		})
	}

	return items, total, nil
}

// FindSessionByID returns an AuthSession for the given session ID.
func (r *UserReadRepo) FindSessionByID(ctx context.Context, id domain.SessionID) (result *shared.AuthSession, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "UserReadRepo.FindSessionByID")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Select("id", "user_id", "device_id", "refresh_token_hash", "expires_at", "revoked", "last_activity", "integration_name", "previous_refresh_hash", "device_fingerprint").
		From(consts.TableSession).
		Where(squirrel.Eq{"id": id.UUID()}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)

	var s shared.AuthSession
	var prevHash *string
	var devFP *string
	err = row.Scan(
		&s.ID, &s.UserID, &s.DeviceID, &s.RefreshTokenHash,
		&s.ExpiresAt, &s.Revoked, &s.LastActivity, &s.IntegrationName,
		&prevHash, &devFP,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableSession, map[string]any{"id": id})
	}
	if prevHash != nil {
		s.PreviousRefreshHash = *prevHash
	}
	if devFP != nil {
		s.DeviceFingerprint = *devFP
	}

	return &s, nil
}

// FindUserForAuth returns an AuthUser with minimal columns for auth middleware.
func (r *UserReadRepo) FindUserForAuth(ctx context.Context, id domain.UserID) (result *shared.AuthUser, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "UserReadRepo.FindUserForAuth")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Select("id", "role_id", "active", "is_approved").
		From(consts.TableUsers).
		Where(squirrel.Eq{"id": id.UUID()}).
		Where(squirrel.Eq{"deleted_at": 0}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)

	var u shared.AuthUser
	err = row.Scan(&u.ID, &u.RoleID, &u.Active, &u.IsApproved)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableUsers, map[string]any{"id": id})
	}

	attrs, err := r.metadata.GetAll(ctx, metadata.EntityTypeUserAttributes, u.ID)
	if err != nil {
		return nil, err
	}
	u.Attributes = attrs

	return &u, nil
}
