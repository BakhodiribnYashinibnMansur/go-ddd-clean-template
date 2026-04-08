package postgres

import (
	"context"
	"time"

	integentity "gct/internal/context/admin/supporting/integration/domain/entity"
	"gct/internal/kernel/consts"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/metadata"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = consts.TableIntegrations

var writeColumns = []string{
	"id", "name", "description", "base_url", "is_active", "created_at", "updated_at",
	"jwt_api_key_hash", "jwt_access_ttl_seconds", "jwt_refresh_ttl_seconds",
	"jwt_public_key_pem", "jwt_previous_public_key_pem",
	"jwt_key_id", "jwt_previous_key_id",
	"jwt_rotated_at", "jwt_rotate_every_days", "jwt_binding_mode", "jwt_max_sessions",
}

// IntegrationWriteRepo implements integentity.IntegrationRepository using PostgreSQL.
type IntegrationWriteRepo struct {
	pool     *pgxpool.Pool
	builder  squirrel.StatementBuilderType
	metadata *metadata.GenericMetadataRepo
}

// NewIntegrationWriteRepo creates a new IntegrationWriteRepo.
func NewIntegrationWriteRepo(pool *pgxpool.Pool) *IntegrationWriteRepo {
	return &IntegrationWriteRepo{
		pool:     pool,
		builder:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		metadata: metadata.NewGenericMetadataRepo(pool),
	}
}

// nullableBytes returns nil for an empty byte slice, otherwise the slice.
// This lets nullable BYTEA columns be written as SQL NULL.
func nullableBytes(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return b
}

// nullableInt returns nil for a zero int, otherwise the value.
func nullableInt(n int) any {
	if n == 0 {
		return nil
	}
	return n
}

// nullableString returns nil for an empty string, otherwise the value.
func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// nullableTime returns nil for a nil pointer, otherwise the dereferenced value.
func nullableTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return *t
}

// Save inserts a new Integration aggregate into the database.
func (r *IntegrationWriteRepo) Save(ctx context.Context, i *integentity.Integration) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "IntegrationWriteRepo.Save")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Insert(tableName).
		Columns(writeColumns...).
		Values(
			i.ID(),
			i.Name(),
			i.Type(),
			i.WebhookURL(),
			i.Enabled(),
			i.CreatedAt(),
			i.UpdatedAt(),
			nullableBytes(i.JWTAPIKeyHash()),
			nullableInt(int(i.JWTAccessTTL()/time.Second)),
			nullableInt(int(i.JWTRefreshTTL()/time.Second)),
			nullableString(i.JWTPublicKeyPEM()),
			nullableString(i.JWTPreviousPublicKeyPEM()),
			nullableString(i.JWTKeyID()),
			nullableString(i.JWTPreviousKeyID()),
			nullableTime(i.JWTRotatedAt()),
			i.JWTRotateEveryDays(),
			i.JWTBindingMode(),
			i.JWTMaxSessions(),
		).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = pgxutil.QuerierFromContext(ctx, r.pool).Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	if err := r.metadata.SetMany(ctx, metadata.EntityTypeIntegrationConfig, i.ID(), i.Config()); err != nil {
		return err
	}

	return nil
}

// FindByID retrieves an Integration aggregate by ID.
func (r *IntegrationWriteRepo) FindByID(ctx context.Context, id integentity.IntegrationID) (result *integentity.Integration, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "IntegrationWriteRepo.FindByID")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Select(writeColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id.UUID()}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	entity, err := scanIntegration(row)
	if err != nil {
		return nil, err
	}

	config, err := r.metadata.GetAll(ctx, metadata.EntityTypeIntegrationConfig, entity.ID())
	if err != nil {
		return nil, err
	}

	return integentity.ReconstructIntegration(
		entity.ID(), entity.CreatedAt(), entity.UpdatedAt(), nil,
		entity.Name(), entity.Type(), entity.APIKey(), entity.WebhookURL(),
		entity.Enabled(), config,
		entity.JWTAPIKeyHash(), entity.JWTAccessTTL(), entity.JWTRefreshTTL(),
		entity.JWTPublicKeyPEM(), entity.JWTPreviousPublicKeyPEM(),
		entity.JWTKeyID(), entity.JWTPreviousKeyID(),
		entity.JWTRotatedAt(), entity.JWTRotateEveryDays(),
		entity.JWTBindingMode(), entity.JWTMaxSessions(),
	), nil
}

// Update updates an Integration aggregate in the database.
func (r *IntegrationWriteRepo) Update(ctx context.Context, i *integentity.Integration) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "IntegrationWriteRepo.Update")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Update(tableName).
		Set("name", i.Name()).
		Set("description", i.Type()).
		Set("base_url", i.WebhookURL()).
		Set("is_active", i.Enabled()).
		Set("updated_at", i.UpdatedAt()).
		Set("jwt_api_key_hash", nullableBytes(i.JWTAPIKeyHash())).
		Set("jwt_access_ttl_seconds", nullableInt(int(i.JWTAccessTTL()/time.Second))).
		Set("jwt_refresh_ttl_seconds", nullableInt(int(i.JWTRefreshTTL()/time.Second))).
		Set("jwt_public_key_pem", nullableString(i.JWTPublicKeyPEM())).
		Set("jwt_previous_public_key_pem", nullableString(i.JWTPreviousPublicKeyPEM())).
		Set("jwt_key_id", nullableString(i.JWTKeyID())).
		Set("jwt_previous_key_id", nullableString(i.JWTPreviousKeyID())).
		Set("jwt_rotated_at", nullableTime(i.JWTRotatedAt())).
		Set("jwt_rotate_every_days", i.JWTRotateEveryDays()).
		Set("jwt_binding_mode", i.JWTBindingMode()).
		Set("jwt_max_sessions", i.JWTMaxSessions()).
		Where(squirrel.Eq{"id": i.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = pgxutil.QuerierFromContext(ctx, r.pool).Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	if err := r.metadata.SetMany(ctx, metadata.EntityTypeIntegrationConfig, i.ID(), i.Config()); err != nil {
		return err
	}

	return nil
}

// Delete removes an Integration by ID.
func (r *IntegrationWriteRepo) Delete(ctx context.Context, id integentity.IntegrationID) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "IntegrationWriteRepo.Delete")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Delete(tableName).
		Where(squirrel.Eq{"id": id.UUID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
	}

	if _, err = pgxutil.QuerierFromContext(ctx, r.pool).Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// RotateJWTKey atomically installs new JWT key material for an integration,
// moving current values to the "previous" slots.
func (r *IntegrationWriteRepo) RotateJWTKey(ctx context.Context, id integentity.IntegrationID, newPublicPEM, newKeyID string) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "IntegrationWriteRepo.RotateJWTKey")
	defer func() { end(err) }()

	sql := `
UPDATE ` + tableName + `
   SET jwt_previous_public_key_pem = jwt_public_key_pem,
       jwt_previous_key_id         = jwt_key_id,
       jwt_public_key_pem          = $1,
       jwt_key_id                  = $2,
       jwt_rotated_at              = NOW(),
       updated_at                  = NOW()
 WHERE id = $3`

	if _, err = pgxutil.QuerierFromContext(ctx, r.pool).Exec(ctx, sql, newPublicPEM, newKeyID, id.UUID()); err != nil {
		return apperrors.HandlePgError(err, tableName, map[string]any{"id": id.UUID()})
	}
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func scanIntegration(row pgx.Row) (*integentity.Integration, error) {
	var (
		id          uuid.UUID
		name        string
		description *string
		baseURL     string
		isActive    bool
		createdAt   time.Time
		updatedAt   time.Time

		jwtAPIKeyHash           []byte
		jwtAccessTTLSecs        *int
		jwtRefreshTTLSecs       *int
		jwtPublicKeyPEM         *string
		jwtPreviousPublicKeyPEM *string
		jwtKeyID                *string
		jwtPreviousKeyID        *string
		jwtRotatedAt            *time.Time
		jwtRotateEveryDays      int
		jwtBindingMode          string
		jwtMaxSessions          int
	)

	err := row.Scan(
		&id, &name, &description, &baseURL, &isActive, &createdAt, &updatedAt,
		&jwtAPIKeyHash, &jwtAccessTTLSecs, &jwtRefreshTTLSecs,
		&jwtPublicKeyPEM, &jwtPreviousPublicKeyPEM,
		&jwtKeyID, &jwtPreviousKeyID,
		&jwtRotatedAt, &jwtRotateEveryDays, &jwtBindingMode, &jwtMaxSessions,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, map[string]any{"id": id})
	}

	desc := ""
	if description != nil {
		desc = *description
	}
	var accessTTL, refreshTTL time.Duration
	if jwtAccessTTLSecs != nil {
		accessTTL = time.Duration(*jwtAccessTTLSecs) * time.Second
	}
	if jwtRefreshTTLSecs != nil {
		refreshTTL = time.Duration(*jwtRefreshTTLSecs) * time.Second
	}
	pub := ""
	if jwtPublicKeyPEM != nil {
		pub = *jwtPublicKeyPEM
	}
	prevPub := ""
	if jwtPreviousPublicKeyPEM != nil {
		prevPub = *jwtPreviousPublicKeyPEM
	}
	kid := ""
	if jwtKeyID != nil {
		kid = *jwtKeyID
	}
	prevKid := ""
	if jwtPreviousKeyID != nil {
		prevKid = *jwtPreviousKeyID
	}

	return integentity.ReconstructIntegration(
		id, createdAt, updatedAt, nil, name, desc, "", baseURL, isActive, nil,
		jwtAPIKeyHash, accessTTL, refreshTTL,
		pub, prevPub, kid, prevKid,
		jwtRotatedAt, jwtRotateEveryDays, jwtBindingMode, jwtMaxSessions,
	), nil
}
