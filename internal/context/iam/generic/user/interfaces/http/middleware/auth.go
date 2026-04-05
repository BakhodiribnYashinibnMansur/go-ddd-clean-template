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
	"gct/internal/kernel/infrastructure/security/jwt"
)

// IntegrationResolver resolves a plain X-API-Key to the verification material
// for the matching integration. Defined here (not imported) to avoid a
// cross-BC import cycle; the adapter is constructed in bootstrap.
type IntegrationResolver interface {
	Resolve(ctx context.Context, plainAPIKey string) (*ResolvedForVerify, error)
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
	issuer          string
	leeway          time.Duration
}

// NewAuthMiddleware initializes a new authentication middleware instance.
// Integration-specific RSA public keys and audiences are resolved per-request
// via the injected resolver.
func NewAuthMiddleware(
	findSession *query.FindSessionHandler,
	findUserForAuth *query.FindUserForAuthHandler,
	cfg *config.Config,
	l logger.Log,
	resolver IntegrationResolver,
	refreshHasher *jwt.RefreshHasher,
) *AuthMiddleware {
	return &AuthMiddleware{
		findSession:     findSession,
		findUserForAuth: findUserForAuth,
		cfg:             cfg,
		l:               l,
		resolver:        resolver,
		refreshHasher:   refreshHasher,
		issuer:          cfg.JWT.Issuer,
		leeway:          cfg.JWT.Leeway,
	}
}
