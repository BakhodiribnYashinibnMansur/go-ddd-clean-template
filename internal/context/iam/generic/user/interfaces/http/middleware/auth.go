// Package middleware provides Gin-compatible authentication handlers
// for the User bounded context, using DDD query handlers and shared domain types.
package middleware

import (
	"crypto/rsa"
	"time"

	"gct/config"
	"gct/internal/context/iam/generic/user/application/query"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/security/jwt"
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
	refreshHasher   *jwt.RefreshHasher
	audience        string
	leeway          time.Duration
}

// NewAuthMiddleware initializes a new authentication middleware instance.
// It pre-parses the RSA public key and constructs the refresh hasher once
// at startup, failing fast on misconfiguration.
func NewAuthMiddleware(
	findSession *query.FindSessionHandler,
	findUserForAuth *query.FindUserForAuthHandler,
	cfg *config.Config,
	l logger.Log,
) *AuthMiddleware {
	pubKey, err := jwt.ParseRSAPublicKey([]byte(cfg.JWT.PublicKey))
	if err != nil {
		l.Fatalw("AuthMiddleware - NewAuthMiddleware - ParseRSAPublicKey", "error", err)
	}

	pepper, err := cfg.JWT.DecodeRefreshPepper()
	if err != nil {
		l.Fatalw("AuthMiddleware - NewAuthMiddleware - DecodeRefreshPepper", "error", err)
	}
	hasher, err := jwt.NewRefreshHasher(pepper)
	if err != nil {
		l.Fatalw("AuthMiddleware - NewAuthMiddleware - NewRefreshHasher", "error", err)
	}

	return &AuthMiddleware{
		findSession:     findSession,
		findUserForAuth: findUserForAuth,
		cfg:             cfg,
		l:               l,
		pubKey:          pubKey,
		refreshHasher:   hasher,
		audience:        cfg.JWT.Audience,
		leeway:          cfg.JWT.Leeway,
	}
}
