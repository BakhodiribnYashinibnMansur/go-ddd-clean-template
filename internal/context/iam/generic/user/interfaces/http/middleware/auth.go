// Package middleware provides Gin-compatible authentication handlers
// for the User bounded context, using DDD query handlers and shared domain types.
package middleware

import (
	"crypto/rsa"

	"gct/config"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/security/jwt"
	"gct/internal/context/iam/generic/user/application/query"
)

// AuthMiddleware manages identity verification for HTTP requests.
// It delegates session and user lookups to User BC query handlers
// and stores shared domain types (AuthSession, AuthUser) in the request context.
type AuthMiddleware struct {
	findSession     *query.FindSessionHandler
	findUserForAuth *query.FindUserForAuthHandler
	cfg             *config.Config
	l               logger.Log
	pubKey          *rsa.PublicKey
}

// NewAuthMiddleware initializes a new authentication middleware instance.
// It pre-parses the RSA public key once at startup for performance.
func NewAuthMiddleware(
	findSession *query.FindSessionHandler,
	findUserForAuth *query.FindUserForAuthHandler,
	cfg *config.Config,
	l logger.Log,
) *AuthMiddleware {
	pubKey, err := jwt.ParseRSAPublicKey(cfg.JWT.PublicKey)
	if err != nil {
		l.Fatalw("AuthMiddleware - NewAuthMiddleware - ParseRSAPublicKey", "error", err)
	}

	return &AuthMiddleware{
		findSession:     findSession,
		findUserForAuth: findUserForAuth,
		cfg:             cfg,
		l:               l,
		pubKey:          pubKey,
	}
}
