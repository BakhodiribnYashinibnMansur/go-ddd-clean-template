package postgres

import (
	"context"
	"time"

	"gct/internal/context/admin/supporting/integration/domain"
	"gct/internal/kernel/consts"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/metadata"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableAPIKeys = consts.TableAPIKeys

var readColumns = []string{
	"id", "name", "description", "base_url", "is_active", "created_at", "updated_at",
	"jwt_api_key_hash", "jwt_access_ttl_seconds", "jwt_refresh_ttl_seconds",
	"jwt_public_key_pem", "jwt_previous_public_key_pem",
	"jwt_key_id", "jwt_previous_key_id",
	"jwt_rotated_at", "jwt_rotate_every_days", "jwt_binding_mode", "jwt_max_sessions",
}

// IntegrationReadRepo implements domain.IntegrationReadRepository for the CQRS read side.
type IntegrationReadRepo struct {
	pool     *pgxpool.Pool
	builder  squirrel.StatementBuilderType
	metadata *metadata.GenericMetadataRepo
}

// NewIntegrationReadRepo creates a new IntegrationReadRepo.
func NewIntegrationReadRepo(pool *pgxpool.Pool) *IntegrationReadRepo {
	return &IntegrationReadRepo{
		pool:     pool,
		builder:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		metadata: metadata.NewGenericMetadataRepo(pool),
	}
}

// FindByID returns an IntegrationView for the given ID.
func (r *IntegrationReadRepo) FindByID(ctx context.Context, id domain.IntegrationID) (result *domain.IntegrationView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "IntegrationReadRepo.FindByID")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Select(readColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id.UUID()}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	view, err := scanIntegrationView(row)
	if err != nil {
		return nil, err
	}

	config, err := r.metadata.GetAll(ctx, metadata.EntityTypeIntegrationConfig, view.ID.UUID())
	if err != nil {
		return nil, err
	}
	view.Config = config

	return view, nil
}

// List returns a paginated list of IntegrationView with optional filters.
func (r *IntegrationReadRepo) List(ctx context.Context, filter domain.IntegrationFilter) (items []*domain.IntegrationView, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "IntegrationReadRepo.List")
	defer func() { end(err) }()

	conds := squirrel.And{}
	if filter.Search != nil {
		conds = append(conds, squirrel.ILike{"name": "%" + *filter.Search + "%"})
	}
	if filter.Enabled != nil {
		conds = append(conds, squirrel.Eq{"is_active": *filter.Enabled})
	}

	// Count total.
	countQB := r.builder.Select("COUNT(*)").From(tableName)
	if len(conds) > 0 {
		countQB = countQB.Where(conds)
	}
	countSQL, countArgs, err := countQB.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	if err = r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}

	// Fetch page.
	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	qb := r.builder.
		Select(readColumns...).
		From(tableName).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(filter.Offset))

	if len(conds) > 0 {
		qb = qb.Where(conds)
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}
	defer rows.Close()

	var views []*domain.IntegrationView
	for rows.Next() {
		v, err := scanIntegrationViewFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		views = append(views, v)
	}

	// Load metadata for each integration.
	for _, v := range views {
		config, err := r.metadata.GetAll(ctx, metadata.EntityTypeIntegrationConfig, v.ID.UUID())
		if err != nil {
			return nil, 0, err
		}
		v.Config = config
	}

	return views, total, nil
}

// integrationViewScanner is the shared scanner row/rows abstraction.
type integrationViewScanner interface {
	Scan(dest ...any) error
}

func scanIntegrationViewInto(s integrationViewScanner) (*domain.IntegrationView, error) {
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

	err := s.Scan(
		&id, &name, &description, &baseURL, &isActive, &createdAt, &updatedAt,
		&jwtAPIKeyHash, &jwtAccessTTLSecs, &jwtRefreshTTLSecs,
		&jwtPublicKeyPEM, &jwtPreviousPublicKeyPEM,
		&jwtKeyID, &jwtPreviousKeyID,
		&jwtRotatedAt, &jwtRotateEveryDays, &jwtBindingMode, &jwtMaxSessions,
	)
	if err != nil {
		return nil, err
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
	strOr := func(p *string) string {
		if p == nil {
			return ""
		}
		return *p
	}

	return &domain.IntegrationView{
		ID:                      domain.IntegrationID(id),
		Name:                    name,
		Type:                    desc,
		APIKey:                  "",
		WebhookURL:              baseURL,
		Enabled:                 isActive,
		CreatedAt:               createdAt,
		UpdatedAt:               updatedAt,
		HasJWT:                  len(jwtAPIKeyHash) > 0,
		JWTAccessTTL:            accessTTL,
		JWTRefreshTTL:           refreshTTL,
		JWTPublicKeyPEM:         strOr(jwtPublicKeyPEM),
		JWTPreviousPublicKeyPEM: strOr(jwtPreviousPublicKeyPEM),
		JWTKeyID:                strOr(jwtKeyID),
		JWTPreviousKeyID:        strOr(jwtPreviousKeyID),
		JWTRotatedAt:            jwtRotatedAt,
		JWTRotateEveryDays:      jwtRotateEveryDays,
		JWTBindingMode:          jwtBindingMode,
		JWTMaxSessions:          jwtMaxSessions,
	}, nil
}

func scanIntegrationView(row pgx.Row) (*domain.IntegrationView, error) {
	v, err := scanIntegrationViewInto(row)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}
	return v, nil
}

func scanIntegrationViewFromRows(rows pgx.Rows) (*domain.IntegrationView, error) {
	v, err := scanIntegrationViewInto(rows)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}
	return v, nil
}

// FindByAPIKey returns an IntegrationAPIKeyView for the given API key string.
func (r *IntegrationReadRepo) FindByAPIKey(ctx context.Context, apiKey string) (result *domain.IntegrationAPIKeyView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "IntegrationReadRepo.FindByAPIKey")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Select("id", "integration_id", "key", "is_active").
		From(tableAPIKeys).
		Where(squirrel.Eq{"key": apiKey}).
		Where(squirrel.Eq{"is_active": true}).
		Where(squirrel.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	var view domain.IntegrationAPIKeyView
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&view.ID,
		&view.IntegrationID,
		&view.Key,
		&view.Active,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableAPIKeys, nil)
	}

	return &view, nil
}

// jwtViewColumns is the column list for JWT-optimized hot-path lookups.
const jwtViewSelect = `SELECT id, name, jwt_access_ttl_seconds, jwt_refresh_ttl_seconds,
       jwt_public_key_pem, COALESCE(jwt_previous_public_key_pem, ''),
       jwt_key_id, COALESCE(jwt_previous_key_id, ''),
       jwt_binding_mode, jwt_max_sessions,
       jwt_rotated_at, jwt_rotate_every_days
  FROM ` + tableName

// ListActiveJWT returns all integrations that have jwt_api_key_hash set.
func (r *IntegrationReadRepo) ListActiveJWT(ctx context.Context) (out []domain.JWTIntegrationView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "IntegrationReadRepo.ListActiveJWT")
	defer func() { end(err) }()

	sql := jwtViewSelect + `
 WHERE is_active = true AND deleted_at IS NULL AND jwt_api_key_hash IS NOT NULL`

	rows, err := r.pool.Query(ctx, sql)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}
	defer rows.Close()

	for rows.Next() {
		v, err := scanJWTIntegrationView(rows)
		if err != nil {
			return nil, apperrors.HandlePgError(err, tableName, nil)
		}
		out = append(out, *v)
	}
	return out, nil
}

// FindJWTByHash returns the integration whose jwt_api_key_hash exactly matches
// the provided hash.
func (r *IntegrationReadRepo) FindJWTByHash(ctx context.Context, hash []byte) (result *domain.JWTIntegrationView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "IntegrationReadRepo.FindJWTByHash")
	defer func() { end(err) }()

	sql := jwtViewSelect + `
 WHERE is_active = true AND deleted_at IS NULL AND jwt_api_key_hash = $1`

	row := r.pool.QueryRow(ctx, sql, hash)
	v, err := scanJWTIntegrationView(row)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}
	return v, nil
}

func scanJWTIntegrationView(s integrationViewScanner) (*domain.JWTIntegrationView, error) {
	var (
		id              uuid.UUID
		name            string
		accessSecs      *int
		refreshSecs     *int
		publicPEM       string
		previousPEM     string
		keyID           string
		previousKeyID   string
		bindingMode     string
		maxSessions     int
		rotatedAt       *time.Time
		rotateEveryDays int
	)
	if err := s.Scan(
		&id, &name, &accessSecs, &refreshSecs,
		&publicPEM, &previousPEM,
		&keyID, &previousKeyID,
		&bindingMode, &maxSessions,
		&rotatedAt, &rotateEveryDays,
	); err != nil {
		return nil, err
	}
	var accessTTL, refreshTTL time.Duration
	if accessSecs != nil {
		accessTTL = time.Duration(*accessSecs) * time.Second
	}
	if refreshSecs != nil {
		refreshTTL = time.Duration(*refreshSecs) * time.Second
	}
	return &domain.JWTIntegrationView{
		ID:                   domain.IntegrationID(id),
		Name:                 name,
		AccessTTL:            accessTTL,
		RefreshTTL:           refreshTTL,
		PublicKeyPEM:         publicPEM,
		PreviousPublicKeyPEM: previousPEM,
		KeyID:                keyID,
		PreviousKeyID:        previousKeyID,
		BindingMode:          bindingMode,
		MaxSessions:          maxSessions,
		RotatedAt:            rotatedAt,
		RotateEveryDays:      rotateEveryDays,
	}, nil
}
