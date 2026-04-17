package postgres

import (
	"context"
	"time"

	userentity "gct/internal/context/iam/generic/user/domain/entity"
	"gct/internal/kernel/consts"
	shared "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/metadata"
	"gct/internal/kernel/infrastructure/pgxutil"

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
	"password_hash", "salt",
	"active", "is_approved",
	"created_at", "updated_at", "deleted_at", "last_seen",
}

// sessionSelectColumns are the columns for SELECT queries (ip_address cast to text for pgx).
var sessionSelectColumns = []string{
	"id", "user_id", "device_id", "device_name", "device_type",
	"ip_address::text", "user_agent", "refresh_token_hash",
	"expires_at", "last_activity", "revoked",
	"created_at", "updated_at",
	"integration_name",
	"previous_refresh_hash",
	"device_fingerprint",
}

// sessionInsertColumns are the columns for INSERT queries (no cast).
var sessionInsertColumns = []string{
	"id", "user_id", "device_id", "device_name", "device_type",
	"ip_address", "user_agent", "refresh_token_hash",
	"expires_at", "last_activity", "revoked",
	"created_at", "updated_at",
	"integration_name",
	"previous_refresh_hash",
	"device_fingerprint",
}

// UserWriteRepo implements userrepo.UserRepository using PostgreSQL.
type UserWriteRepo struct {
	pool     *pgxpool.Pool
	builder  squirrel.StatementBuilderType
	metadata *metadata.GenericMetadataRepo
}

// NewUserWriteRepo creates a new UserWriteRepo.
func NewUserWriteRepo(pool *pgxpool.Pool) *UserWriteRepo {
	return &UserWriteRepo{
		pool:     pool,
		builder:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		metadata: metadata.NewGenericMetadataRepo(pool),
	}
}

// Save inserts a new User aggregate (and its sessions) into the database.
// The caller provides a Querier (typically a transaction from EventCommitter).
func (r *UserWriteRepo) Save(ctx context.Context, q shared.Querier, user *userentity.User) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "UserWriteRepo.Save")
	defer func() { end(err) }()

	v := userToInsertValues(user)

	sql, args, err := r.builder.
		Insert(usersTable).
		Columns(userColumns...).
		Values(
			v.id, v.roleID, v.username, v.email, v.phone,
			v.passwordHash, "",
			v.active, v.isApproved,
			v.createdAt, v.updatedAt, v.deletedAt, v.lastSeen,
		).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = q.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, usersTable, map[string]any{"id": v.id})
	}

	// Persist user attributes via EAV table.
	if err := r.metadata.SetManyTx(ctx, q, metadata.EntityTypeUserAttributes, user.ID(), user.Attributes()); err != nil {
		return err
	}

	// Batch-insert all sessions in a single round-trip.
	return r.insertSessionsBatch(ctx, q, user.ID(), user.Sessions())
}

// insertSessionsBatch writes all sessions for a user in a single INSERT
// round-trip. Returns nil when the slice is empty.
func (r *UserWriteRepo) insertSessionsBatch(ctx context.Context, q shared.Querier, userID uuid.UUID, sessions []userentity.Session) error {
	if len(sessions) == 0 {
		return nil
	}

	qb := r.builder.
		Insert(sessionTable).
		Columns(sessionInsertColumns...)

	for i := range sessions {
		s := &sessions[i]
		qb = qb.Values(
			s.ID(), s.UserID(), s.DeviceID(), s.DeviceName(), string(s.DeviceType()),
			s.IPAddress().String(), s.UserAgent().String(), s.RefreshTokenHash(),
			s.ExpiresAt(), s.LastActivity(), s.IsRevoked(),
			s.CreatedAt(), s.UpdatedAt(),
			s.IntegrationName(),
			s.PreviousRefreshHash(),
			s.DeviceFingerprint(),
		)
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err := q.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, sessionTable, map[string]any{"user_id": userID, "count": len(sessions)})
	}
	return nil
}

// FindByID retrieves a User aggregate by ID, including its sessions.
func (r *UserWriteRepo) FindByID(ctx context.Context, id userentity.UserID) (result *userentity.User, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "UserWriteRepo.FindByID")
	defer func() { end(err) }()

	// Fetch user row.
	sql, args, err := r.builder.
		Select(userColumns...).
		From(usersTable).
		Where(squirrel.Eq{"id": id.UUID()}).
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

	// Load attributes from EAV table.
	attrs, err := r.metadata.GetAll(ctx, metadata.EntityTypeUserAttributes, user.ID())
	if err != nil {
		return nil, err
	}

	// Fetch sessions for this user.
	sessions, err := r.findSessionsByUserID(ctx, user.ID())
	if err != nil {
		return nil, err
	}

	// Reconstruct with sessions and attributes.
	return userentity.ReconstructUser(
		user.ID(),
		user.CreatedAt(),
		user.UpdatedAt(),
		user.DeletedAt(),
		user.Phone(),
		user.Email(),
		user.Username(),
		user.Password(),
		user.RoleID(),
		attrs,
		user.IsActive(),
		user.IsApproved(),
		user.LastSeen(),
		sessions,
	), nil
}

// Update updates the User aggregate in the database.
// The caller provides a Querier (typically a transaction from EventCommitter).
func (r *UserWriteRepo) Update(ctx context.Context, q shared.Querier, user *userentity.User) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "UserWriteRepo.Update")
	defer func() { end(err) }()

	v := userToInsertValues(user)

	sql, args, err := r.builder.
		Update(usersTable).
		Set("role_id", v.roleID).
		Set("username", v.username).
		Set("email", v.email).
		Set("phone", v.phone).
		Set("password_hash", v.passwordHash).
		Set("active", v.active).
		Set("is_approved", v.isApproved).
		Set("updated_at", v.updatedAt).
		Set("deleted_at", v.deletedAt).
		Set("last_seen", v.lastSeen).
		Where(squirrel.Eq{"id": v.id}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = q.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, usersTable, map[string]any{"id": v.id})
	}

	// Persist user attributes via EAV table.
	if err := r.metadata.SetManyTx(ctx, q, metadata.EntityTypeUserAttributes, user.ID(), user.Attributes()); err != nil {
		return err
	}

	return r.upsertSessions(ctx, q, user.Sessions())
}

// upsertSessions inserts all sessions (or updates conflicting rows) in a
// single INSERT ... ON CONFLICT round-trip. Safe for replay during token
// refresh.
func (r *UserWriteRepo) upsertSessions(ctx context.Context, q shared.Querier, sessions []userentity.Session) error {
	if len(sessions) == 0 {
		return nil
	}

	qb := r.builder.
		Insert(sessionTable).
		Columns(sessionInsertColumns...)

	for i := range sessions {
		s := &sessions[i]
		qb = qb.Values(
			s.ID(), s.UserID(), s.DeviceID(), s.DeviceName(), string(s.DeviceType()),
			s.IPAddress().String(), s.UserAgent().String(), s.RefreshTokenHash(),
			s.ExpiresAt(), s.LastActivity(), s.IsRevoked(),
			s.CreatedAt(), s.UpdatedAt(),
			s.IntegrationName(),
			s.PreviousRefreshHash(),
			s.DeviceFingerprint(),
		)
	}

	sql, args, err := qb.
		Suffix("ON CONFLICT (id) DO UPDATE SET refresh_token_hash = EXCLUDED.refresh_token_hash, previous_refresh_hash = EXCLUDED.previous_refresh_hash, last_activity = EXCLUDED.last_activity, revoked = EXCLUDED.revoked, updated_at = EXCLUDED.updated_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err := q.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, sessionTable, map[string]any{"count": len(sessions)})
	}
	return nil
}

// Delete performs a soft delete on the user by setting deleted_at to the current unix timestamp.
// The caller provides a Querier (typically a transaction from EventCommitter).
func (r *UserWriteRepo) Delete(ctx context.Context, q shared.Querier, id userentity.UserID) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "UserWriteRepo.Delete")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Update(usersTable).
		Set("deleted_at", squirrel.Expr("EXTRACT(EPOCH FROM NOW())::bigint")).
		Where(squirrel.Eq{"id": id.UUID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
	}

	if _, err = q.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, usersTable, nil)
	}
	return nil
}

// FindByPhone finds a user by phone number.
func (r *UserWriteRepo) FindByPhone(ctx context.Context, phone userentity.Phone) (result *userentity.User, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "UserWriteRepo.FindByPhone")
	defer func() { end(err) }()

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

	attrs, err := r.metadata.GetAll(ctx, metadata.EntityTypeUserAttributes, user.ID())
	if err != nil {
		return nil, err
	}

	sessions, err := r.findSessionsByUserID(ctx, user.ID())
	if err != nil {
		return nil, err
	}

	return userentity.ReconstructUser(
		user.ID(),
		user.CreatedAt(),
		user.UpdatedAt(),
		user.DeletedAt(),
		user.Phone(),
		user.Email(),
		user.Username(),
		user.Password(),
		user.RoleID(),
		attrs,
		user.IsActive(),
		user.IsApproved(),
		user.LastSeen(),
		sessions,
	), nil
}

// FindByEmail finds a user by email address.
func (r *UserWriteRepo) FindByEmail(ctx context.Context, email userentity.Email) (result *userentity.User, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "UserWriteRepo.FindByEmail")
	defer func() { end(err) }()

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

	attrs, err := r.metadata.GetAll(ctx, metadata.EntityTypeUserAttributes, user.ID())
	if err != nil {
		return nil, err
	}

	sessions, err := r.findSessionsByUserID(ctx, user.ID())
	if err != nil {
		return nil, err
	}

	return userentity.ReconstructUser(
		user.ID(),
		user.CreatedAt(),
		user.UpdatedAt(),
		user.DeletedAt(),
		user.Phone(),
		user.Email(),
		user.Username(),
		user.Password(),
		user.RoleID(),
		attrs,
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
func (r *UserWriteRepo) findSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]userentity.Session, error) {
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

	var sessions []userentity.Session
	for rows.Next() {
		s, err := scanSessionFromRows(rows)
		if err != nil {
			return nil, apperrors.HandlePgError(err, sessionTable, nil)
		}
		sessions = append(sessions, *s)
	}

	return sessions, nil
}

// ActiveSessionCount returns the number of non-revoked, non-expired sessions
// for the given user. Consulted during sign-in to decide whether to evict an
// old session before admitting a new one.
func (r *UserWriteRepo) ActiveSessionCount(ctx context.Context, userID userentity.UserID) (count int, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "UserWriteRepo.ActiveSessionCount")
	defer func() { end(err) }()

	const sql = `SELECT COUNT(*) FROM ` + sessionTable +
		` WHERE user_id = $1 AND revoked = false AND expires_at > NOW()`

	if err = r.pool.QueryRow(ctx, sql, userID.UUID()).Scan(&count); err != nil {
		return 0, apperrors.HandlePgError(err, sessionTable, nil)
	}
	return count, nil
}

// RevokeOldestActiveSession revokes the single oldest active session for the
// user, ordered by last_activity ASC NULLS FIRST, created_at ASC. Returns the
// revoked session ID, or NilSessionID when no active session was available to
// revoke (idempotent — safe to call in a loop).
func (r *UserWriteRepo) RevokeOldestActiveSession(ctx context.Context, userID userentity.UserID) (result userentity.SessionID, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "UserWriteRepo.RevokeOldestActiveSession")
	defer func() { end(err) }()

	const sql = `UPDATE ` + sessionTable + ` SET revoked = true, updated_at = NOW()
 WHERE id = (
   SELECT id FROM ` + sessionTable + `
    WHERE user_id = $1 AND revoked = false AND expires_at > NOW()
    ORDER BY last_activity ASC NULLS FIRST, created_at ASC
    LIMIT 1
 )
 RETURNING id`

	var id uuid.UUID
	if err = r.pool.QueryRow(ctx, sql, userID.UUID()).Scan(&id); err != nil {
		if err == pgx.ErrNoRows {
			return userentity.NilSessionID, nil
		}
		return userentity.NilSessionID, apperrors.HandlePgError(err, sessionTable, nil)
	}
	return userentity.SessionID(id), nil
}

// RevokeSessionsByIntegration revokes all active sessions for a user within a
// specific integration. Returns the count of revoked sessions.
func (r *UserWriteRepo) RevokeSessionsByIntegration(ctx context.Context, userID userentity.UserID, integrationName string) (count int, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "UserWriteRepo.RevokeSessionsByIntegration")
	defer func() { end(err) }()

	const sql = `UPDATE ` + sessionTable + ` SET revoked = true, updated_at = NOW()
 WHERE user_id = $1 AND integration_name = $2 AND revoked = false AND expires_at > NOW()`

	tag, err := r.pool.Exec(ctx, sql, userID.UUID(), integrationName)
	if err != nil {
		return 0, apperrors.HandlePgError(err, sessionTable, nil)
	}
	return int(tag.RowsAffected()), nil
}

// scanUser scans a single user row (pgx.Row) and returns a User aggregate without sessions.
func scanUser(row pgx.Row) (*userentity.User, error) {
	var (
		id         uuid.UUID
		roleID     *uuid.UUID
		username   *string
		email      *string
		phone      string
		pwHash     string
		salt       *string
		active     bool
		isApproved bool
		createdAt  time.Time
		updatedAt  time.Time
		deletedAt  int64
		lastSeen   *time.Time
	)

	err := row.Scan(
		&id, &roleID, &username, &email, &phone,
		&pwHash, &salt,
		&active, &isApproved,
		&createdAt, &updatedAt, &deletedAt, &lastSeen,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, usersTable, nil)
	}

	return reconstructUserFromRow(
		id, roleID, username, email, phone, pwHash,
		active, isApproved, createdAt, updatedAt, deletedAt, lastSeen,
	), nil
}

// scanUserFromRows scans a user from pgx.Rows.
func scanUserFromRows(rows pgx.Rows) (*userentity.User, error) {
	var (
		id         uuid.UUID
		roleID     *uuid.UUID
		username   *string
		email      *string
		phone      string
		pwHash     string
		salt       *string
		active     bool
		isApproved bool
		createdAt  time.Time
		updatedAt  time.Time
		deletedAt  int64
		lastSeen   *time.Time
	)

	err := rows.Scan(
		&id, &roleID, &username, &email, &phone,
		&pwHash, &salt,
		&active, &isApproved,
		&createdAt, &updatedAt, &deletedAt, &lastSeen,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, usersTable, nil)
	}

	return reconstructUserFromRow(
		id, roleID, username, email, phone, pwHash,
		active, isApproved, createdAt, updatedAt, deletedAt, lastSeen,
	), nil
}

// reconstructUserFromRow builds a userentity.User from raw scanned values.
// Attributes are loaded separately via the metadata repo; nil is passed here.
func reconstructUserFromRow(
	id uuid.UUID,
	roleID *uuid.UUID,
	username *string,
	emailStr *string,
	phone, pwHash string,
	active, isApproved bool,
	createdAt, updatedAt time.Time,
	deletedAtUnix int64,
	lastSeen *time.Time,
) *userentity.User {
	phonVO, _ := userentity.NewPhone(phone)
	password := userentity.NewPasswordFromHash(pwHash)

	var emailVO *userentity.Email
	if emailStr != nil {
		e, err := userentity.NewEmail(*emailStr)
		if err == nil {
			emailVO = &e
		}
	}

	var deletedAt *time.Time
	if deletedAtUnix != 0 {
		t := time.Unix(deletedAtUnix, 0)
		deletedAt = &t
	}

	return userentity.ReconstructUser(
		id,
		createdAt, updatedAt, deletedAt,
		phonVO,
		emailVO,
		username,
		password,
		roleID,
		nil, // attributes loaded separately via metadata repo
		active, isApproved,
		lastSeen,
		nil, // sessions loaded separately
	)
}

// scanSessionFromRows scans a session from pgx.Rows.
func scanSessionFromRows(rows pgx.Rows) (*userentity.Session, error) {
	var (
		id                  uuid.UUID
		userID              uuid.UUID
		deviceID            *string
		deviceName          *string
		deviceType          *string
		ipAddress           *string
		userAgent           *string
		refreshTokenHash    *string
		expiresAt           time.Time
		lastActivity        time.Time
		revoked             bool
		createdAt           time.Time
		updatedAt           time.Time
		integrationName     *string
		previousRefreshHash *string
		deviceFingerprint   *string
	)

	err := rows.Scan(
		&id, &userID, &deviceID, &deviceName, &deviceType,
		&ipAddress, &userAgent, &refreshTokenHash,
		&expiresAt, &lastActivity, &revoked,
		&createdAt, &updatedAt,
		&integrationName,
		&previousRefreshHash,
		&deviceFingerprint,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, sessionTable, nil)
	}

	deref := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}

	s := userentity.ReconstructSession(
		id,
		createdAt, updatedAt, nil,
		userID,
		deref(deviceID), deref(deviceName),
		userentity.SessionDeviceType(deref(deviceType)),
		deref(ipAddress), deref(userAgent), deref(refreshTokenHash),
		expiresAt, lastActivity,
		revoked,
		deref(integrationName),
		deref(previousRefreshHash),
		deref(deviceFingerprint),
	)
	return s, nil
}
