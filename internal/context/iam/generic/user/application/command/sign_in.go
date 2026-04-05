package command

import (
	"context"
	"crypto/rsa"
	"strings"
	"time"

	"gct/internal/context/iam/generic/user/domain"
	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	jwtpkg "gct/internal/kernel/infrastructure/security/jwt"

	"github.com/google/uuid"
)

// SignInCommand represents an authentication attempt using login credentials and device metadata.
// Login accepts either a phone number or email — the handler auto-detects the format via "@" presence.
// DeviceType is uppercased internally to match the PostgreSQL ENUM constraint (e.g., "WEB", "MOBILE").
// APIKey is the plain X-API-Key header value used to resolve which integration the session belongs to.
type SignInCommand struct {
	Login      string
	Password   string
	DeviceType string
	IP         string
	UserAgent  string
	APIKey     string
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
	repo        domain.UserRepository
	eventBus    application.EventBus
	logger      commandLogger
	signIn      domain.SignInService
	jwtConfig   JWTConfig
	maxSessions func(ctx context.Context) int
}

// defaultMaxSessions is the fallback cap when no dynamic resolver is wired.
// Matches the seeded "user.max_sessions" default.
const defaultMaxSessions = 3

// NewSignInHandler creates a new SignInHandler.
//
// An optional maxSessionsFn may be supplied to read the per-user active
// session cap dynamically from SiteSetting (or any other source). When nil
// or omitted, a constant returning defaultMaxSessions (3) is used. This
// keeps the User BC free of a cross-BC import on SiteSetting — the caller
// wires the closure in bootstrap.
func NewSignInHandler(
	repo domain.UserRepository,
	eventBus application.EventBus,
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
		eventBus:    eventBus,
		logger:      logger,
		signIn:      domain.SignInService{},
		jwtConfig:   jwtCfg,
		maxSessions: fn,
	}
}

// Handle executes the SignInCommand and returns SignInResult.
func (h *SignInHandler) Handle(ctx context.Context, cmd SignInCommand) (result *SignInResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "SignInHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "SignIn", "user")()

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
		return nil, apperrors.MapToServiceError(err)
	}

	deviceType := domain.SessionDeviceType(strings.ToUpper(cmd.DeviceType))

	session, err := h.signIn.SignIn(user, cmd.Password, deviceType, cmd.IP, cmd.UserAgent, resolved.Name)
	if err != nil {
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
		count, cerr := h.repo.ActiveSessionCount(ctx, domain.UserID(user.ID()))
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
		if _, rerr := h.repo.RevokeOldestActiveSession(ctx, domain.UserID(user.ID())); rerr != nil {
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

	// Persist user (with the updated session containing the refresh token hash).
	if err := h.repo.Update(ctx, user); err != nil {
		h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "SignIn", Entity: "user", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, user.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "SignIn", Entity: "user", Err: err}.KV()...)
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
	)
	if err != nil {
		h.logger.Errorc(ctx, "access token generation failed", logger.F{Op: "SignIn", Entity: "user", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	result = &SignInResult{
		UserID:       user.ID(),
		SessionID:    session.ID(),
		AccessToken:  accessToken,
		RefreshToken: refToken.String(),
	}

	return result, nil
}

// findUser looks up a user by phone or email depending on the login format.
// Returns ErrServiceUnauthorized when user is not found (sign-in should not reveal user existence).
func (h *SignInHandler) findUser(ctx context.Context, login string) (*domain.User, error) {
	if strings.Contains(login, "@") {
		email, err := domain.NewEmail(login)
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

	phone, err := domain.NewPhone(login)
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
