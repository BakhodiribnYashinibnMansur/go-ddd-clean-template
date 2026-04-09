package command

import (
	"context"
	"crypto/rsa"
	"strings"
	"time"

	userentity "gct/internal/context/iam/generic/user/domain/entity"
	userrepo "gct/internal/context/iam/generic/user/domain/repository"
	usersvc "gct/internal/context/iam/generic/user/domain/service"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/metrics"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/infrastructure/security/audit"
	jwtpkg "gct/internal/kernel/infrastructure/security/jwt"
	"gct/internal/kernel/infrastructure/security/ratelimit"
	"gct/internal/kernel/infrastructure/security/tbh"
	"gct/internal/kernel/outbox"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

// SignInCommand represents an authentication attempt using login credentials and device metadata.
// Login accepts either a phone number or email — the handler auto-detects the format via "@" presence.
// DeviceType is uppercased internally to match the PostgreSQL ENUM constraint (e.g., "WEB", "MOBILE").
// APIKey is the plain X-API-Key header value used to resolve which integration the session belongs to.
type SignInCommand struct {
	Login             string
	Password          string
	DeviceType        string
	IP                string
	UserAgent         string
	APIKey            string
	DeviceFingerprint string
}

// SignInResult holds the output of a successful sign-in.
type SignInResult struct {
	UserID       uuid.UUID
	SessionID    uuid.UUID
	AccessToken  string
	RefreshToken string
}

// IntegrationResolver resolves a plain X-API-Key to the signing material for
// a specific integration. Defined here (not imported) to avoid a cross-BC
// import cycle; the adapter is constructed in bootstrap.
type IntegrationResolver interface {
	Resolve(ctx context.Context, plainAPIKey string) (*JWTResolved, error)
}

// JWTResolved is the output of IntegrationResolver.Resolve containing
// everything SignInHandler needs to issue access + refresh tokens.
type JWTResolved struct {
	Name        string // = audience
	PrivateKey  *rsa.PrivateKey
	KeyID       string
	AccessTTL   time.Duration
	RefreshTTL  time.Duration
	MaxSessions int
}

// JWTConfig holds the parameters needed for JWT token generation.
type JWTConfig struct {
	Issuer        string
	Resolver      IntegrationResolver
	RefreshHasher *jwtpkg.RefreshHasher
}

// SignInHandler handles the SignInCommand.
type SignInHandler struct {
	repo        userrepo.UserRepository
	committer   *outbox.EventCommitter
	logger      commandLogger
	signIn      usersvc.SignInService
	jwtConfig   JWTConfig
	maxSessions func(ctx context.Context) int

	// Phase S1 security — all optional, nil-safe.
	auditLogger audit.Logger
	limiter     *ratelimit.AuthLimiter
	tbhPepper   []byte

	// Business metrics — optional, nil-safe.
	bm *metrics.BusinessMetrics
}

// defaultMaxSessions is the fallback cap when no dynamic resolver is wired.
// Matches the seeded "user.max_sessions" default.
const defaultMaxSessions = 3

// SignInSecurityDeps groups optional Phase S1 security dependencies for sign-in.
// All fields are optional — nil values disable the corresponding check.
type SignInSecurityDeps struct {
	AuditLogger audit.Logger
	Limiter     *ratelimit.AuthLimiter
	TBHPepper   []byte
}

// NewSignInHandler creates a new SignInHandler.
//
// An optional maxSessionsFn may be supplied to read the per-user active
// session cap dynamically from SiteSetting (or any other source). When nil
// or omitted, a constant returning defaultMaxSessions (3) is used. This
// keeps the User BC free of a cross-BC import on SiteSetting — the caller
// wires the closure in bootstrap.
func NewSignInHandler(
	repo userrepo.UserRepository,
	committer *outbox.EventCommitter,
	logger commandLogger,
	jwtCfg JWTConfig,
	maxSessionsFn ...func(ctx context.Context) int,
) *SignInHandler {
	var fn func(context.Context) int
	if len(maxSessionsFn) > 0 && maxSessionsFn[0] != nil {
		fn = maxSessionsFn[0]
	} else {
		fn = func(context.Context) int { return defaultMaxSessions }
	}
	return &SignInHandler{
		repo:        repo,
		committer:   committer,
		logger:      logger,
		signIn:      usersvc.SignInService{},
		jwtConfig:   jwtCfg,
		maxSessions: fn,
		auditLogger: audit.NoopLogger{},
	}
}

// WithSecurityDeps injects Phase S1 security dependencies into the handler.
// Call after NewSignInHandler; safe to omit entirely (all checks degrade
// gracefully when deps are nil).
func (h *SignInHandler) WithSecurityDeps(deps SignInSecurityDeps) *SignInHandler {
	if deps.AuditLogger != nil {
		h.auditLogger = deps.AuditLogger
	}
	h.limiter = deps.Limiter
	h.tbhPepper = deps.TBHPepper
	return h
}

// WithBusinessMetrics injects business metrics into the handler.
func (h *SignInHandler) WithBusinessMetrics(bm *metrics.BusinessMetrics) {
	h.bm = bm
}

// Handle executes the SignInCommand and returns SignInResult.
func (h *SignInHandler) Handle(ctx context.Context, cmd SignInCommand) (result *SignInResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "SignInHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "SignIn", "user")()

	// Phase S1: rate-limit checks (IP + user).
	if h.limiter != nil {
		if limitErr := h.limiter.CheckIP(ctx, cmd.IP); limitErr != nil {
			h.auditLogger.Log(ctx, audit.Entry{
				Event: audit.EventSignInFailed, IPAddress: cmd.IP, UserAgent: cmd.UserAgent,
				Metadata: map[string]any{"reason": "ip_rate_limited"},
			})
			return nil, apperrors.NewHandlerError(apperrors.ErrHandlerTooManyRequests, "")
		}
		if limitErr := h.limiter.CheckUser(ctx, cmd.Login); limitErr != nil {
			h.auditLogger.Log(ctx, audit.Entry{
				Event: audit.EventAccountLocked, IPAddress: cmd.IP, UserAgent: cmd.UserAgent,
				Metadata: map[string]any{"login": cmd.Login},
			})
			return nil, apperrors.NewServiceError(apperrors.ErrServiceUnauthorized, "")
		}
	}

	// Resolve the integration from the supplied X-API-Key. Any failure here
	// surfaces as a generic 401 — we must never leak whether the key exists.
	if h.jwtConfig.Resolver == nil {
		h.logger.Errorc(ctx, "integration resolver not configured", logger.F{Op: "SignIn", Entity: "user"}.KV()...)
		return nil, apperrors.NewServiceError(apperrors.ErrServiceUnauthorized, "")
	}
	resolved, err := h.jwtConfig.Resolver.Resolve(ctx, cmd.APIKey)
	if err != nil || resolved == nil {
		return nil, apperrors.NewServiceError(apperrors.ErrServiceUnauthorized, "")
	}

	// Find user by phone or email based on login format.
	user, err := h.findUser(ctx, cmd.Login)
	if err != nil {
		h.recordFailedAttempt(ctx, cmd)
		return nil, apperrors.MapToServiceError(err)
	}

	deviceType := userentity.SessionDeviceType(strings.ToUpper(cmd.DeviceType))

	session, err := h.signIn.SignIn(user, cmd.Password, deviceType, cmd.IP, cmd.UserAgent, resolved.Name, cmd.DeviceFingerprint)
	if err != nil {
		h.recordFailedAttempt(ctx, cmd)
		return nil, apperrors.MapToServiceError(err)
	}

	// Enforce the "max concurrent sessions per user" cap by evicting the
	// oldest active sessions until the user is under the cap. We never
	// refuse sign-in — any repo failure here degrades into an admit-
	// everyone posture and is logged upstream. The safety cap of 5 loop
	// iterations prevents runaway eviction should the DB return stale
	// counts.
	maxN := h.maxSessions(ctx)
	if maxN <= 0 {
		maxN = defaultMaxSessions
	}
	for i := 0; i < 5; i++ {
		count, cerr := h.repo.ActiveSessionCount(ctx, userentity.UserID(user.ID()))
		if cerr != nil {
			h.logger.Warnc(ctx, "active session count failed",
				logger.F{Op: "SignIn", Entity: "user", Err: cerr}.KV()...)
			break
		}
		// The current sign-in's session has not been persisted yet, so it
		// is NOT counted. Evict once we're already AT the cap, so after
		// the new session is inserted we land exactly on maxN.
		if count < maxN {
			break
		}
		if _, rerr := h.repo.RevokeOldestActiveSession(ctx, userentity.UserID(user.ID())); rerr != nil {
			h.logger.Warnc(ctx, "revoke oldest session failed",
				logger.F{Op: "SignIn", Entity: "user", Err: rerr}.KV()...)
			break
		}
	}

	// Generate refresh token and store its hash on the session before persisting.
	refToken, err := jwtpkg.GenerateRefreshToken(
		h.jwtConfig.RefreshHasher,
		user.ID().String(),
		session.ID().String(),
		session.DeviceID(),
		resolved.RefreshTTL,
	)
	if err != nil {
		h.logger.Errorc(ctx, "refresh token generation failed", logger.F{Op: "SignIn", Entity: "user", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}
	session.SetRefreshTokenHash(refToken.Hashed)

	// Persist user (with the updated session containing the refresh token hash)
	// and publish domain events via the transactional outbox.
	if err := h.committer.Commit(ctx, func(ctx context.Context) error {
		if err := h.repo.Update(ctx, user); err != nil {
			h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "SignIn", Entity: "user", Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, user.Events); err != nil {
		return nil, err
	}

	// Compute TBH claim for device-binding when pepper is configured.
	var tbhClaim string
	if len(h.tbhPepper) > 0 {
		tbhClaim = tbh.Compute(h.tbhPepper, cmd.IP, cmd.UserAgent)
	}

	// Generate access token (signed JWT) using the per-integration key material.
	accessToken, err := jwtpkg.GenerateAccessToken(
		user.ID().String(),
		session.ID().String(),
		h.jwtConfig.Issuer,
		resolved.Name,
		resolved.KeyID,
		resolved.PrivateKey,
		resolved.AccessTTL,
		tbhClaim,
	)
	if err != nil {
		h.logger.Errorc(ctx, "access token generation failed", logger.F{Op: "SignIn", Entity: "user", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	// Phase S1: reset rate-limit counters + audit success on successful sign-in.
	if h.limiter != nil {
		_ = h.limiter.ResetUser(ctx, cmd.Login)
	}
	uid := user.ID()
	sid := session.ID()
	h.auditLogger.Log(ctx, audit.Entry{
		Event: audit.EventSignInSuccess, IPAddress: cmd.IP, UserAgent: cmd.UserAgent,
		UserID: &uid, SessionID: &sid, IntegrationName: resolved.Name,
	})

	// Business metrics: successful sign-in.
	h.bm.Inc(ctx, "user_signins", attribute.String("integration", resolved.Name))

	result = &SignInResult{
		UserID:       user.ID(),
		SessionID:    session.ID(),
		AccessToken:  accessToken,
		RefreshToken: refToken.String(),
	}

	return result, nil
}

// recordFailedAttempt increments rate-limit counters and emits an audit event
// for a failed authentication attempt.
func (h *SignInHandler) recordFailedAttempt(ctx context.Context, cmd SignInCommand) {
	h.bm.Inc(ctx, "user_signin_failures", attribute.String("ip", cmd.IP))
	if h.limiter != nil {
		_ = h.limiter.RecordFailedIP(ctx, cmd.IP)
		_ = h.limiter.RecordFailedUser(ctx, cmd.Login)
	}
	h.auditLogger.Log(ctx, audit.Entry{
		Event: audit.EventSignInFailed, IPAddress: cmd.IP, UserAgent: cmd.UserAgent,
		Metadata: map[string]any{"login": cmd.Login},
	})
}

// findUser looks up a user by phone or email depending on the login format.
// Returns ErrServiceUnauthorized when user is not found (sign-in should not reveal user existence).
func (h *SignInHandler) findUser(ctx context.Context, login string) (*userentity.User, error) {
	if strings.Contains(login, "@") {
		email, err := userentity.NewEmail(login)
		if err != nil {
			return nil, apperrors.MapToServiceError(err)
		}
		user, err := h.repo.FindByEmail(ctx, email)
		if err != nil {
			if apperrors.Is(err, apperrors.ErrRepoNotFound) {
				return nil, apperrors.NewServiceError(apperrors.ErrServiceUnauthorized, "")
			}
			return nil, apperrors.MapToServiceError(err)
		}
		return user, nil
	}

	phone, err := userentity.NewPhone(login)
	if err != nil {
		return nil, apperrors.MapToServiceError(err)
	}
	user, err := h.repo.FindByPhone(ctx, phone)
	if err != nil {
		if apperrors.Is(err, apperrors.ErrRepoNotFound) {
			return nil, apperrors.NewServiceError(apperrors.ErrServiceUnauthorized, "")
		}
		return nil, apperrors.MapToServiceError(err)
	}
	return user, nil
}
