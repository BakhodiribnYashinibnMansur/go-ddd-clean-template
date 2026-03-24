// Package auth manages authentication entry points (Login, Signup, Refresh) and CSRF protection.
package auth

import (
	"crypto/rsa"

	"gct/config"
	"gct/internal/usecase"
	"gct/internal/shared/infrastructure/security/csrf"
	"gct/internal/shared/infrastructure/security/jwt"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// Controller implements the high-level business logic for authentication.
// It integrates JWT signing for tokens and CSRF token generation for secure forms.
type Controller struct {
	u             *usecase.UseCase
	l             logger.Log
	cfg           *config.Config
	privateKey    *rsa.PrivateKey // Key used for signing issued JWTs.
	csrfGenerator *csrf.Generator // Engine for generating secure CSRF tokens.
	csrfStore     csrf.Store      // Backend for persisting/verifying CSRF tokens.
}

// ControllerI defines the public contract for the authentication operations.
type ControllerI interface {
	SignIn(c *gin.Context)       // Authenticate and issue tokens.
	SignUp(c *gin.Context)       // Register a new account.
	SignOut(c *gin.Context)      // Invalidate session.
	CsrfToken(c *gin.Context)    // Issue a fresh CSRF token for the frontend.
	RefreshToken(c *gin.Context) // Use refresh token to rotate access tokens.
}

// New initializes an Auth controller.
// It pre-parses the JWT private key and configures CSRF protection based on application settings.
func New(u *usecase.UseCase, cfg *config.Config, l logger.Log) ControllerI {
	// Pre-parse the private key for signing tokens once during startup.
	pk, err := jwt.ParseRSAPrivateKey(cfg.JWT.PrivateKey)
	if err != nil {
		l.Warn("AuthController - New - parsedPrivateKey error or missing. JWT signing will be disabled.", err)
	}

	// Initialize the CSRF protection engine with the configured global secret.
	csrfGen := csrf.NewGenerator(csrf.Config{
		Secret:     []byte(cfg.App.CSRFSecret),
		Expiration: csrf.DefaultExpiration,
	})

	// Default to an in-memory store for CSRF; can be swapped for Redis in distributed environments.
	csrfStore := csrf.NewMemoryStore()

	return &Controller{
		u:             u,
		l:             l,
		cfg:           cfg,
		privateKey:    pk,
		csrfGenerator: csrfGen,
		csrfStore:     csrfStore,
	}
}
