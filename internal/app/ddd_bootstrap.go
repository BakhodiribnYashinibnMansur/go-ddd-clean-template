package app

import (
	"context"
	"fmt"

	"gct/config"
	"gct/internal/context/content/supporting/announcement"
	"gct/internal/context/iam/supporting/audit"
	"gct/internal/context/iam/generic/authz"
	"gct/internal/context/admin/supporting/statistics"
	"gct/internal/context/admin/supporting/dataexport"
	"gct/internal/context/admin/supporting/errorcode"
	"gct/internal/context/admin/generic/featureflag"
	"gct/internal/context/content/generic/file"
	"gct/internal/context/admin/supporting/integration"
	"gct/internal/context/ops/supporting/activitylog"
	"gct/internal/context/ops/supporting/iprule"

	"gct/internal/context/ops/generic/metric"
	"gct/internal/context/content/generic/notification"
	"gct/internal/context/ops/generic/ratelimit"
	"gct/internal/context/iam/generic/session"
	"gct/internal/kernel/application"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/metrics"
	"gct/internal/kernel/outbox"
	"gct/internal/context/admin/supporting/sitesetting"
	"gct/internal/context/ops/generic/systemerror"
	"gct/internal/context/content/generic/translation"
	"gct/internal/context/iam/generic/user"
	"gct/internal/context/iam/generic/user/application/command"
	usermw "gct/internal/context/iam/generic/user/interfaces/http/middleware"
	"gct/internal/context/iam/generic/usersetting"
	securityaudit "gct/internal/kernel/infrastructure/security/audit"
	"gct/internal/kernel/infrastructure/security/keyring"
	jwtpkg "gct/internal/kernel/infrastructure/security/jwt"
	securityratelimit "gct/internal/kernel/infrastructure/security/ratelimit"
	"gct/internal/kernel/infrastructure/security/revocation"
	integentity "gct/internal/context/admin/supporting/integration/domain/entity"
	integrationquery "gct/internal/context/admin/supporting/integration/application/query"
	"crypto/rsa"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DDDBoundedContexts holds all DDD bounded contexts.
type DDDBoundedContexts struct {
	User         *user.BoundedContext
	Authz        *authz.BoundedContext
	Session      *session.BoundedContext
	Audit        *audit.BoundedContext
	Statistics   *statistics.BoundedContext
	SystemError  *systemerror.BoundedContext
	Metric       *metric.BoundedContext
	FeatureFlag  *featureflag.BoundedContext
	Integration  *integration.BoundedContext

	Notification *notification.BoundedContext
	Announcement *announcement.BoundedContext
	Translation  *translation.BoundedContext
	SiteSetting  *sitesetting.BoundedContext
	RateLimit    *ratelimit.BoundedContext
	IPRule       *iprule.BoundedContext

	DataExport   *dataexport.BoundedContext
	File         *file.BoundedContext
	UserSetting  *usersetting.BoundedContext
	ErrorCode    *errorcode.BoundedContext
	ActivityLog  *activitylog.BoundedContext
}

// SecurityDeps groups Phase S1 security infrastructure that is wired into
// bounded contexts after construction. All fields are optional — nil values
// disable the corresponding security feature.
type SecurityDeps struct {
	AuditLogger securityaudit.Logger
	RevStore    *revocation.Store
	AuthLimiter *securityratelimit.AuthLimiter
	TBHPepper   []byte
}

// NewDDDBoundedContexts creates all bounded contexts with their dependencies.
// The Integration BC is constructed first so that callers can extract its
// ResolveJWTAPIKey handler to build the sign-in/middleware resolver adapters;
// the User BC then receives the wired JWTConfig with that resolver injected.
func NewDDDBoundedContexts(ctx context.Context, pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log, bm *metrics.BusinessMetrics, jwtCfg command.JWTConfig, cfg *config.Config, apiKeyPepper []byte, kr *keyring.Keyring, secDeps SecurityDeps) (*DDDBoundedContexts, error) {
	_ = bm // available for BC injection when needed

	// EventCommitter wraps repo writes + domain events in the transactional
	// outbox pattern. Pass nil writer for dev mode (direct event bus publish).
	committer := outbox.NewEventCommitter(pool, nil, eventBus, l)

	ffBC, err := featureflag.NewBoundedContext(ctx, pool, eventBus, committer, l)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize feature flag BC: %w", err)
	}

	integrationBC := integration.NewBoundedContext(pool, eventBus, committer, apiKeyPepper, cfg.JWT.CacheTTL, l)

	// Wire the sign-in resolver now that both the Integration BC and keyring
	// are available. User BC needs this to issue per-integration JWTs.
	jwtCfg.Resolver = &signInResolverAdapter{h: integrationBC.ResolveJWTAPIKey, kr: kr}

	// SiteSetting BC is constructed before User BC so sign-in can look up
	// the "user.max_sessions" cap through a runtime closure (no cross-BC
	// import — the User BC only sees the func).
	siteSettingBC := sitesetting.NewBoundedContext(pool, committer, l)
	maxSessionsFn := func(ctx context.Context) int {
		n, _ := siteSettingBC.UserMaxSessions.Handle(ctx)
		return n
	}

	userBC := user.NewBoundedContext(pool, eventBus, committer, l, jwtCfg, maxSessionsFn)

	// Phase S1: wire security deps into sign-in and sign-out handlers.
	userBC.WireSecurityDeps(secDeps.AuditLogger, secDeps.RevStore, command.SignInSecurityDeps{
		AuditLogger: secDeps.AuditLogger,
		Limiter:     secDeps.AuthLimiter,
		TBHPepper:   secDeps.TBHPepper,
	})

	return &DDDBoundedContexts{
		User:         userBC,
		Authz:        authz.NewBoundedContext(pool, eventBus, l),
		Audit:        audit.NewBoundedContext(pool, eventBus, l),
		SystemError:  systemerror.NewBoundedContext(pool, committer, l),
		Metric:       metric.NewBoundedContext(pool, committer, l),
		FeatureFlag:  ffBC,
		Integration:  integrationBC,

		Notification: notification.NewBoundedContext(pool, eventBus, committer, l),
		Announcement: announcement.NewBoundedContext(pool, committer, l),
		Translation:  translation.NewBoundedContext(pool, committer, l),
		SiteSetting:  siteSettingBC,
		RateLimit:    ratelimit.NewBoundedContext(pool, committer, l),
		IPRule:       iprule.NewBoundedContext(pool, committer, l),

		DataExport:   dataexport.NewBoundedContext(pool, committer, l),
		File:         file.NewBoundedContext(pool, committer, l),
		UserSetting:  usersetting.NewBoundedContext(pool, eventBus, l),
		ErrorCode:    errorcode.NewBoundedContext(pool, committer, l),

		Session:     session.NewBoundedContext(pool, eventBus, l),
		Statistics:  statistics.NewBoundedContext(pool, l),
		ActivityLog: activitylog.NewBoundedContext(pool, eventBus, l),
	}, nil
}

// signInResolverAdapter implements command.IntegrationResolver for the User
// BC's sign-in handler: it resolves a plain X-API-Key through the Integration
// BC, then loads the current RSA private key from the on-disk keyring.
type signInResolverAdapter struct {
	h  *integrationquery.ResolveJWTAPIKeyHandler
	kr *keyring.Keyring
}

// Resolve maps a plaintext API key to the signing material for the matching
// integration. Any failure propagates to the caller, which must surface it
// as a generic 401 to avoid leaking key existence.
func (a *signInResolverAdapter) Resolve(ctx context.Context, plainAPIKey string) (*command.JWTResolved, error) {
	view, err := a.h.Handle(ctx, integrationquery.ResolveJWTAPIKeyQuery{PlainAPIKey: plainAPIKey})
	if err != nil {
		return nil, err
	}
	kp, err := a.kr.EnsureAndLoad(view.Name, view.KeyID)
	if err != nil {
		return nil, err
	}
	return &command.JWTResolved{
		Name:        view.Name,
		PrivateKey:  kp.PrivateKey,
		KeyID:       kp.KeyID,
		AccessTTL:   view.AccessTTL,
		RefreshTTL:  view.RefreshTTL,
		MaxSessions: view.MaxSessions,
	}, nil
}

// middlewareResolverAdapter implements middleware.IntegrationResolver for the
// auth middleware: it resolves the plaintext X-API-Key and exposes the
// current + previous public keys for verification (supporting rotation).
type middlewareResolverAdapter struct {
	h *integrationquery.ResolveJWTAPIKeyHandler
}

// Resolve maps a plaintext API key to the verification material for the
// matching integration. The previous public key is best-effort: if it fails
// to parse we silently drop it — the current key still works.
func (a *middlewareResolverAdapter) Resolve(ctx context.Context, plainAPIKey string) (*usermw.ResolvedForVerify, error) {
	view, err := a.h.Handle(ctx, integrationquery.ResolveJWTAPIKeyQuery{PlainAPIKey: plainAPIKey})
	if err != nil {
		return nil, err
	}
	pub, err := jwtpkg.ParseRSAPublicKey([]byte(view.PublicKeyPEM))
	if err != nil {
		return nil, err
	}
	var prev *rsa.PublicKey
	if view.PreviousPublicKeyPEM != "" {
		if p, perr := jwtpkg.ParseRSAPublicKey([]byte(view.PreviousPublicKeyPEM)); perr == nil {
			prev = p
		}
	}
	return &usermw.ResolvedForVerify{
		Name:              view.Name,
		PublicKey:         pub,
		PreviousPublicKey: prev,
		KeyID:             view.KeyID,
		PreviousKeyID:     view.PreviousKeyID,
		BindingMode:       view.BindingMode,
		MaxSessions:       view.MaxSessions,
	}, nil
}

// NewMiddlewareResolver builds the middleware-side resolver adapter. Exposed
// so app.go can construct the AuthMiddleware after NewDDDBoundedContexts has
// returned.
func NewMiddlewareResolver(bcs *DDDBoundedContexts) usermw.IntegrationResolver {
	return &middlewareResolverAdapter{h: bcs.Integration.ResolveJWTAPIKey}
}

// integrationListerAdapter bridges the Integration BC's read repository to the
// keyring.IntegrationLister interface so the rotation handler does not import
// the Integration BC's domain package directly.
type integrationListerAdapter struct {
	readRepo integrationReadRepoForKeyring
}

// integrationReadRepoForKeyring is the subset of the Integration read repo
// that the keyring rotation handler needs.
type integrationReadRepoForKeyring interface {
	ListActiveJWT(ctx context.Context) ([]integentity.JWTIntegrationView, error)
}

func (a *integrationListerAdapter) ListActiveJWT(ctx context.Context) ([]keyring.JWTIntegrationView, error) {
	views, err := a.readRepo.ListActiveJWT(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]keyring.JWTIntegrationView, len(views))
	for i, v := range views {
		out[i] = keyring.JWTIntegrationView{
			Name:            v.Name,
			KeyID:           v.KeyID,
			RotatedAt:       v.RotatedAt,
			RotateEveryDays: v.RotateEveryDays,
		}
	}
	return out, nil
}
