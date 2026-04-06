// Package middleware provides Gin-compatible authentication handlers
// for the User bounded context, using DDD query handlers and shared domain types.
package middleware

import (
	"context"
	"crypto/rsa"
	"time"

	"gct/config"
	"gct/internal/context/iam/generic/user/application/query"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/security/audit"
	"gct/internal/kernel/infrastructure/security/apikeythrottle"
	"gct/internal/kernel/infrastructure/security/jwt"
	"gct/internal/kernel/infrastructure/security/revocation"

	"github.com/google/uuid"
)

// IntegrationResolver resolves a plain X-API-Key to the verification material
// for the matching integration. Defined here (not imported) to avoid a
// cross-BC import cycle; the adapter is constructed in bootstrap.
type IntegrationResolver interface {
	Resolve(ctx context.Context, plainAPIKey string) (*ResolvedForVerify, error)
}

// SessionRevoker provides the ability to revoke sessions directly in the DB.
// Used by the refresh-token reuse detection to revoke all sessions for a
// (user, integration) pair without loading the full User aggregate.
type SessionRevoker interface {
	RevokeSessionsByIntegration(ctx context.Context, userID uuid.UUID, integrationName string) (int, error)
}

// ResolvedForVerify carries everything the middleware needs to cryptographically
// verify an access token: current + previous (rotating) public keys plus the
// audience/name and binding-mode policy.
type ResolvedForVerify struct {
	Name              string // = audience
	PublicKey         *rsa.PublicKey
	PreviousPublicKey *rsa.PublicKey // may be nil
	KeyID             string
	PreviousKeyID     string
	BindingMode       string
	MaxSessions       int
}

// AuthMiddleware manages identity verification for HTTP requests.
// It delegates session and user lookups to User BC query handlers
// and stores shared domain types (AuthSession, AuthUser) in the request context.
type AuthMiddleware struct {
	findSession     *query.FindSessionHandler
	findUserForAuth *query.FindUserForAuthHandler
	cfg             *config.Config
	l               logger.Log
	resolver        IntegrationResolver
	refreshHasher   *jwt.RefreshHasher
	sessionRevoker  SessionRevoker
	issuer          string
	leeway          time.Duration

	// Phase S1 security components — all optional, nil-safe.
	auditLogger     audit.Logger
	revStore        *revocation.Store
	apiKeyThrottle  *apikeythrottle.Throttle
	tbhPepper       []byte
}

// NewAuthMiddleware initializes a new authentication middleware instance.
// Integration-specific RSA public keys and audiences are resolved per-request
// via the injected resolver.
//
// An optional SessionRevoker may be supplied as the last argument. When
// provided, the refresh-token reuse detection logic can revoke all sessions
// for a (user, integration) pair. If omitted, reuse detection still rejects
// the request with 401 but cannot perform the bulk revocation.
func NewAuthMiddleware(
	findSession *query.FindSessionHandler,
	findUserForAuth *query.FindUserForAuthHandler,
	cfg *config.Config,
	l logger.Log,
	opts ...any,
) *AuthMiddleware {
	m := &AuthMiddleware{
		findSession:     findSession,
		findUserForAuth: findUserForAuth,
		cfg:             cfg,
		l:               l,
		issuer:          cfg.JWT.Issuer,
		leeway:          cfg.JWT.Leeway,
		auditLogger:     audit.NoopLogger{}, // default to noop
	}
	for _, o := range opts {
		switch v := o.(type) {
		case IntegrationResolver:
			m.resolver = v
		case *jwt.RefreshHasher:
			m.refreshHasher = v
		case SessionRevoker:
			m.sessionRevoker = v
		case audit.Logger:
			m.auditLogger = v
		case *revocation.Store:
			m.revStore = v
		case *apikeythrottle.Throttle:
			m.apiKeyThrottle = v
		case TBHPepper:
			m.tbhPepper = []byte(v)
		}
	}
	return m
}

// TBHPepper is a named type used to pass the TBH pepper through the
// variadic opts of NewAuthMiddleware without ambiguity.
type TBHPepper []byte
