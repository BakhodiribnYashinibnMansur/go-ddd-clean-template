package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gct/internal/shared/domain/consts"
	shared "gct/internal/shared/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/pgxutil"
	"gct/internal/user/domain"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	usersTable   = consts.TableUsers
	sessionTable = consts.TableSession
)

// userColumns are the columns for the users table.
var userColumns = []string{
	"id", "role_id", "username", "email", "phone",
	"password_hash", "salt", "attributes",
	"active", "is_approved",
	"created_at", "updated_at", "deleted_at", "last_seen",
}

// sessionSelectColumns are the columns for SELECT queries (ip_address cast to text for pgx).
var sessionSelectColumns = []string{
	"id", "user_id", "device_id", "device_name", "device_type",
	"ip_address::text", "user_agent", "refresh_token_hash",
	"expires_at", "last_activity", "revoked",
	"created_at", "updated_at",
}

// sessionInsertColumns are the columns for INSERT queries (no cast).
var sessionInsertColumns = []string{
	"id", "user_id", "device_id", "device_name", "device_type",
	"ip_address", "user_agent", "refresh_token_hash",
	"expires_at", "last_activity", "revoked",
	"created_at", "updated_at",
}

// UserWriteRepo implements domain.UserRepository using PostgreSQL.
type UserWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewUserWriteRepo creates a new UserWriteRepo.
func NewUserWriteRepo(pool *pgxpool.Pool) *UserWriteRepo {
	return &UserWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new User aggregate (and its sessions) into the database.
func (r *UserWriteRepo) Save(ctx context.Context, user *domain.User) error {
	return pgxutil.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		attrsJSON, err := json.Marshal(user.Attributes())
		if err != nil {
			return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToMarshalJSON)
		}

		var emailVal *string
		if user.Email() != nil {
			v := user.Email().Value()
			emailVal = &v
		}

		var deletedAtVal int64
		if user.DeletedAt() != nil {
			deletedAtVal = user.DeletedAt().Unix()
		}

		sql, args, err := r.builder.
			Insert(usersTable).
			Columns(userColumns...).
			Values(
				user.ID(),
				user.RoleID(),
				user.Username(),
				emailVal,
				user.Phone().Value(),
				user.Password().Hash(),
				"", // salt — bcrypt includes salt in hash
				attrsJSON,
				user.IsActive(),
				user.IsApproved(),
				user.CreatedAt(),
				user.UpdatedAt(),
				deletedAtVal,
				user.LastSeen(),
			).
			ToSql()
		if err != nil {
			return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
		}

		if _, err = tx.Exec(ctx, sql, args...); err != nil {
			return apperrors.HandlePgError(err, usersTable, nil)
		}

		// Insert sessions if any.
		for _, s := range user.Sessions() {
			if err := r.insertSession(ctx, tx, &s); err != nil {
				return err
			}
		}

		return nil
	})
}

// insertSession inserts a single session row within an existing transaction.
func (r *UserWriteRepo) insertSession(ctx context.Context, tx pgx.Tx, s *domain.Session) error {
	sql, args, err := r.builder.
		Insert(sessionTable).
		Columns(sessionInsertColumns...).
		Values(
			s.ID(),
			s.UserID(),
			s.DeviceID(),
			s.DeviceName(),
			string(s.DeviceType()),
			s.IPAddress(),
			s.UserAgent(),
			s.RefreshTokenHash(),
			s.ExpiresAt(),
			s.LastActivity(),
			s.IsRevoked(),
			s.CreatedAt(),
			s.UpdatedAt(),
		).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = tx.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, sessionTable, nil)
	}
	return nil
}

// FindByID retrieves a User aggregate by ID, including its sessions.
func (r *UserWriteRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	// Fetch user row.
	sql, args, err := r.builder.
		Select(userColumns...).
		From(usersTable).
		Where(squirrel.Eq{"id": id}).
		Where(squirrel.Eq{"deleted_at": 0}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	user, err := scanUser(row)
	if err != nil {
		return nil, apperrors.HandlePgError(err, usersTable, map[string]any{"id": id})
	}

	// Fetch sessions for this user.
	sessions, err := r.findSessionsByUserID(ctx, user.ID())
	if err != nil {
		return nil, err
	}

	// Reconstruct with sessions.
	return domain.ReconstructUser(
		user.ID(),
		user.CreatedAt(),
		user.UpdatedAt(),
		user.DeletedAt(),
		user.Phone(),
		user.Email(),
		user.Username(),
		user.Password(),
		user.RoleID(),
		user.Attributes(),
		user.IsActive(),
		user.IsApproved(),
		user.LastSeen(),
		sessions,
	), nil
}

// Update updates the User aggregate in the database.
func (r *UserWriteRepo) Update(ctx context.Context, user *domain.User) error {
	return pgxutil.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		attrsJSON, err := json.Marshal(user.Attributes())
		if err != nil {
			return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToMarshalJSON)
		}

		var emailVal *string
		if user.Email() != nil {
			v := user.Email().Value()
			emailVal = &v
		}

		var deletedAtVal int64
		if user.DeletedAt() != nil {
			deletedAtVal = user.DeletedAt().Unix()
		}

		sql, args, err := r.builder.
			Update(usersTable).
			Set("role_id", user.RoleID()).
			Set("username", user.Username()).
			Set("email", emailVal).
			Set("phone", user.Phone().Value()).
			Set("password_hash", user.Password().Hash()).
			Set("attributes", attrsJSON).
			Set("active", user.IsActive()).
			Set("is_approved", user.IsApproved()).
			Set("updated_at", user.UpdatedAt()).
			Set("deleted_at", deletedAtVal).
			Set("last_seen", user.LastSeen()).
			Where(squirrel.Eq{"id": user.ID()}).
			ToSql()
		if err != nil {
			return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
		}

		if _, err = tx.Exec(ctx, sql, args...); err != nil {
			return apperrors.HandlePgError(err, usersTable, nil)
		}

		// Upsert sessions: INSERT new ones, UPDATE existing ones (avoid FK violations).
		for _, s := range user.Sessions() {
			upsertSQL, upsertArgs, upsertErr := r.builder.
				Insert(sessionTable).
				Columns(sessionInsertColumns...).
				Values(
					s.ID(), s.UserID(), s.DeviceID(), s.DeviceName(), string(s.DeviceType()),
					s.IPAddress(), s.UserAgent(), s.RefreshTokenHash(),
					s.ExpiresAt(), s.LastActivity(), s.IsRevoked(),
					s.CreatedAt(), s.UpdatedAt(),
				).
				Suffix("ON CONFLICT (id) DO UPDATE SET refresh_token_hash = EXCLUDED.refresh_token_hash, last_activity = EXCLUDED.last_activity, revoked = EXCLUDED.revoked, updated_at = EXCLUDED.updated_at").
				ToSql()
			if upsertErr != nil {
				return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
			}
			if _, err = tx.Exec(ctx, upsertSQL, upsertArgs...); err != nil {
				return apperrors.HandlePgError(err, sessionTable, nil)
			}
		}

		return nil
	})
}

// Delete performs a soft delete on the user by setting deleted_at to the current unix timestamp.
func (r *UserWriteRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := r.builder.
		Update(usersTable).
		Set("deleted_at", squirrel.Expr("EXTRACT(EPOCH FROM NOW())::bigint")).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, usersTable, nil)
	}
	return nil
}

// List retrieves a paginated list of users (without sessions).
func (r *UserWriteRepo) List(ctx context.Context, filter shared.Pagination) ([]*domain.User, int64, error) {
	// Count total.
	countSQL, countArgs, err := r.builder.
		Select("COUNT(*)").
		From(usersTable).
		Where(squirrel.Eq{"deleted_at": 0}).
		ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	var total int64
	if err = r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.HandlePgError(err, usersTable, nil)
	}

	// Fetch page.
	qb := r.builder.
		Select(userColumns...).
		From(usersTable).
		Where(squirrel.Eq{"deleted_at": 0}).
		Limit(uint64(filter.Limit)).
		Offset(uint64(filter.Offset))

	if filter.SortBy != "" {
		order := "ASC"
		if filter.SortOrder == "DESC" {
			order = "DESC"
		}
		qb = qb.OrderBy(fmt.Sprintf("%s %s", filter.SortBy, order))
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, usersTable, nil)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u, err := scanUserFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, usersTable, nil)
		}
		users = append(users, u)
	}

	return users, total, nil
}

// FindByPhone finds a user by phone number.
func (r *UserWriteRepo) FindByPhone(ctx context.Context, phone domain.Phone) (*domain.User, error) {
	sql, args, err := r.builder.
		Select(userColumns...).
		From(usersTable).
		Where(squirrel.Eq{"phone": phone.Value()}).
		Where(squirrel.Eq{"deleted_at": 0}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	user, err := scanUser(row)
	if err != nil {
		return nil, apperrors.HandlePgError(err, usersTable, map[string]any{"phone": phone.Value()})
	}

	sessions, err := r.findSessionsByUserID(ctx, user.ID())
	if err != nil {
		return nil, err
	}

	return domain.ReconstructUser(
		user.ID(),
		user.CreatedAt(),
		user.UpdatedAt(),
		user.DeletedAt(),
		user.Phone(),
		user.Email(),
		user.Username(),
		user.Password(),
		user.RoleID(),
		user.Attributes(),
		user.IsActive(),
		user.IsApproved(),
		user.LastSeen(),
		sessions,
	), nil
}

// FindByEmail finds a user by email address.
func (r *UserWriteRepo) FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	sql, args, err := r.builder.
		Select(userColumns...).
		From(usersTable).
		Where(squirrel.Eq{"email": email.Value()}).
		Where(squirrel.Eq{"deleted_at": 0}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	user, err := scanUser(row)
	if err != nil {
		return nil, apperrors.HandlePgError(err, usersTable, map[string]any{"email": email.Value()})
	}

	sessions, err := r.findSessionsByUserID(ctx, user.ID())
	if err != nil {
		return nil, err
	}

	return domain.ReconstructUser(
		user.ID(),
		user.CreatedAt(),
		user.UpdatedAt(),
		user.DeletedAt(),
		user.Phone(),
		user.Email(),
		user.Username(),
		user.Password(),
		user.RoleID(),
		user.Attributes(),
		user.IsActive(),
		user.IsApproved(),
		user.LastSeen(),
		sessions,
	), nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// findSessionsByUserID returns all sessions for a given user ID.
func (r *UserWriteRepo) findSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Session, error) {
	sql, args, err := r.builder.
		Select(sessionSelectColumns...).
		From(sessionTable).
		Where(squirrel.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, sessionTable, nil)
	}
	defer rows.Close()

	var sessions []domain.Session
	for rows.Next() {
		s, err := scanSessionFromRows(rows)
		if err != nil {
			return nil, apperrors.HandlePgError(err, sessionTable, nil)
		}
		sessions = append(sessions, *s)
	}

	return sessions, nil
}

// scanUser scans a single user row (pgx.Row) and returns a User aggregate without sessions.
func scanUser(row pgx.Row) (*domain.User, error) {
	var (
		id         uuid.UUID
		roleID     *uuid.UUID
		username   *string
		email      *string
		phone      string
		pwHash     string
		salt       *string
		attrsJSON  []byte
		active     bool
		isApproved bool
		createdAt  time.Time
		updatedAt  time.Time
		deletedAt  int64
		lastSeen   *time.Time
	)

	err := row.Scan(
		&id, &roleID, &username, &email, &phone,
		&pwHash, &salt, &attrsJSON,
		&active, &isApproved,
		&createdAt, &updatedAt, &deletedAt, &lastSeen,
	)
	if err != nil {
		return nil, err
	}

	return reconstructUserFromRow(
		id, roleID, username, email, phone, pwHash, attrsJSON,
		active, isApproved, createdAt, updatedAt, deletedAt, lastSeen,
	), nil
}

// scanUserFromRows scans a user from pgx.Rows.
func scanUserFromRows(rows pgx.Rows) (*domain.User, error) {
	var (
		id         uuid.UUID
		roleID     *uuid.UUID
		username   *string
		email      *string
		phone      string
		pwHash     string
		salt       *string
		attrsJSON  []byte
		active     bool
		isApproved bool
		createdAt  time.Time
		updatedAt  time.Time
		deletedAt  int64
		lastSeen   *time.Time
	)

	err := rows.Scan(
		&id, &roleID, &username, &email, &phone,
		&pwHash, &salt, &attrsJSON,
		&active, &isApproved,
		&createdAt, &updatedAt, &deletedAt, &lastSeen,
	)
	if err != nil {
		return nil, err
	}

	return reconstructUserFromRow(
		id, roleID, username, email, phone, pwHash, attrsJSON,
		active, isApproved, createdAt, updatedAt, deletedAt, lastSeen,
	), nil
}

// reconstructUserFromRow builds a domain.User from raw scanned values.
func reconstructUserFromRow(
	id uuid.UUID,
	roleID *uuid.UUID,
	username *string,
	emailStr *string,
	phone, pwHash string,
	attrsJSON []byte,
	active, isApproved bool,
	createdAt, updatedAt time.Time,
	deletedAtUnix int64,
	lastSeen *time.Time,
) *domain.User {
	phonVO, _ := domain.NewPhone(phone)
	password := domain.NewPasswordFromHash(pwHash)

	var emailVO *domain.Email
	if emailStr != nil {
		e, err := domain.NewEmail(*emailStr)
		if err == nil {
			emailVO = &e
		}
	}

	var attrs map[string]any
	if len(attrsJSON) > 0 {
		_ = json.Unmarshal(attrsJSON, &attrs)
	}

	var deletedAt *time.Time
	if deletedAtUnix != 0 {
		t := time.Unix(deletedAtUnix, 0)
		deletedAt = &t
	}

	return domain.ReconstructUser(
		id,
		createdAt, updatedAt, deletedAt,
		phonVO,
		emailVO,
		username,
		password,
		roleID,
		attrs,
		active, isApproved,
		lastSeen,
		nil, // sessions loaded separately
	)
}

// scanSessionFromRows scans a session from pgx.Rows.
func scanSessionFromRows(rows pgx.Rows) (*domain.Session, error) {
	var (
		id               uuid.UUID
		userID           uuid.UUID
		deviceID         *string
		deviceName       *string
		deviceType       *string
		ipAddress        *string
		userAgent        *string
		refreshTokenHash *string
		expiresAt        time.Time
		lastActivity     time.Time
		revoked          bool
		createdAt        time.Time
		updatedAt        time.Time
	)

	err := rows.Scan(
		&id, &userID, &deviceID, &deviceName, &deviceType,
		&ipAddress, &userAgent, &refreshTokenHash,
		&expiresAt, &lastActivity, &revoked,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	deref := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}

	s := domain.ReconstructSession(
		id,
		createdAt, updatedAt, nil,
		userID,
		deref(deviceID), deref(deviceName),
		domain.SessionDeviceType(deref(deviceType)),
		deref(ipAddress), deref(userAgent), deref(refreshTokenHash),
		expiresAt, lastActivity,
		revoked,
	)
	return s, nil
}
