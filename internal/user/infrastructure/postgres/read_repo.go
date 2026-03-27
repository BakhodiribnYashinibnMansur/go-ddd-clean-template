package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	shared "gct/internal/shared/domain"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/user/domain"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// readUserColumns are the columns selected for read-model queries.
var readUserColumns = []string{
	"id", "role_id", "username", "email", "phone",
	"attributes", "active", "is_approved",
	"last_seen", "created_at", "updated_at",
}

// UserReadRepo implements domain.UserReadRepository for the CQRS read side.
type UserReadRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewUserReadRepo creates a new UserReadRepo.
func NewUserReadRepo(pool *pgxpool.Pool) *UserReadRepo {
	return &UserReadRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// FindByID returns a UserView for the given user ID.
func (r *UserReadRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.UserView, error) {
	sql, args, err := r.builder.
		Select(readUserColumns...).
		From(consts.TableUsers).
		Where(squirrel.Eq{"id": id}).
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
		attrsJSON  []byte
		active     bool
		isApproved bool
		lastSeen   *time.Time
		createdAt  time.Time
		updatedAt  time.Time
	)

	err = row.Scan(
		&uid, &roleID, &username, &email, &phone,
		&attrsJSON, &active, &isApproved,
		&lastSeen, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableUsers, map[string]any{"id": id})
	}

	var attrs map[string]any
	if len(attrsJSON) > 0 {
		_ = json.Unmarshal(attrsJSON, &attrs)
	}

	return &domain.UserView{
		ID:         uid,
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
func (r *UserReadRepo) List(ctx context.Context, filter domain.UsersFilter) ([]*domain.UserView, int64, error) {
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

	var total int64
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

	var views []*domain.UserView
	for rows.Next() {
		var (
			uid        uuid.UUID
			roleID     *uuid.UUID
			username   *string
			email      *string
			phone      string
			attrsJSON  []byte
			active     bool
			isApproved bool
			lastSeen   *time.Time
			createdAt  time.Time
			updatedAt  time.Time
		)

		if err := rows.Scan(
			&uid, &roleID, &username, &email, &phone,
			&attrsJSON, &active, &isApproved,
			&lastSeen, &createdAt, &updatedAt,
		); err != nil {
			return nil, 0, apperrors.HandlePgError(err, consts.TableUsers, nil)
		}

		var attrs map[string]any
		if len(attrsJSON) > 0 {
			_ = json.Unmarshal(attrsJSON, &attrs)
		}

		views = append(views, &domain.UserView{
			ID:         uid,
			Phone:      phone,
			Email:      email,
			Username:   username,
			RoleID:     roleID,
			Attributes: attrs,
			Active:     active,
			IsApproved: isApproved,
		})
	}

	return views, total, nil
}

// FindSessionByID returns an AuthSession for the given session ID.
func (r *UserReadRepo) FindSessionByID(ctx context.Context, id uuid.UUID) (*shared.AuthSession, error) {
	sql, args, err := r.builder.
		Select("id", "user_id", "device_id", "refresh_token_hash", "expires_at", "revoked", "last_activity").
		From(consts.TableSession).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)

	var s shared.AuthSession
	err = row.Scan(
		&s.ID, &s.UserID, &s.DeviceID, &s.RefreshTokenHash,
		&s.ExpiresAt, &s.Revoked, &s.LastActivity,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableSession, map[string]any{"id": id})
	}

	return &s, nil
}

// FindUserForAuth returns an AuthUser with minimal columns for auth middleware.
func (r *UserReadRepo) FindUserForAuth(ctx context.Context, id uuid.UUID) (*shared.AuthUser, error) {
	sql, args, err := r.builder.
		Select("id", "role_id", "active", "is_approved", "attributes").
		From(consts.TableUsers).
		Where(squirrel.Eq{"id": id}).
		Where(squirrel.Eq{"deleted_at": 0}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)

	var (
		u         shared.AuthUser
		attrsJSON []byte
	)
	err = row.Scan(&u.ID, &u.RoleID, &u.Active, &u.IsApproved, &attrsJSON)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableUsers, map[string]any{"id": id})
	}

	if len(attrsJSON) > 0 {
		_ = json.Unmarshal(attrsJSON, &u.Attributes)
	}

	return &u, nil
}
