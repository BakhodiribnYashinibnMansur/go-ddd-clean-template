// Package middleware provides Gin-compatible handlers for cross-cutting concerns
// like authentication, authorization, logging, auditing, and security.
package auth

import (
	"crypto/rsa"

	"gct/config"
	"gct/internal/usecase"
	"gct/internal/usecase/authz"
	"gct/internal/usecase/integration"
	"gct/internal/usecase/user/client"
	"gct/internal/usecase/user/session"
	"gct/pkg/jwt"
	"gct/pkg/logger"
)

// AuthMiddleware manages identity verification and permission enforcement.
// It integrates with user, session, and authorization use cases.
type AuthMiddleware struct {
	userUC        client.UseCaseI
	sessionuc     session.UseCaseI
	authzUC       authz.UseCaseI
	integrationUC integration.UseCaseI
	cfg           *config.Config
	l             logger.Log
	pubKey        *rsa.PublicKey
}

// NewAuthMiddleware initializes a new authentication middleware instance.
// It pre-parses the RSA public key once at startup for performance.
func NewAuthMiddleware(u *usecase.UseCase, cfg *config.Config, l logger.Log) *AuthMiddleware {
	pubKey, err := jwt.ParseRSAPublicKey(cfg.JWT.PublicKey)
	if err != nil {
		l.Fatalw("AuthMiddleware - NewAuthMiddleware - ParseRSAPublicKey", "error", err)
	}

	return &AuthMiddleware{
		userUC:        u.User.Client(),
		sessionuc:     u.User.Session(),
		authzUC:       u.Authz,
		integrationUC: u.Integration,
		cfg:           cfg,
		l:             l,
		pubKey:        pubKey,
	}
}
