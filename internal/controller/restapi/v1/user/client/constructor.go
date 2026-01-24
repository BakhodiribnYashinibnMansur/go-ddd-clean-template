// Package client manages user profile operations and authentication entry points (Login, Signup).
package client

import (
	"crypto/rsa"

	"gct/config"
	"gct/internal/usecase"
	"gct/pkg/csrf"
	"gct/pkg/jwt"
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
)

// Controller implements the high-level business logic for user management.
// It integrates JWT signing for tokens and CSRF token generation for secure forms.
type Controller struct {
	u             *usecase.UseCase
	l             logger.Log
	cfg           *config.Config
	privateKey    *rsa.PrivateKey // Key used for signing issued JWTs.
	csrfGenerator *csrf.Generator // Engine for generating secure CSRF tokens.
	csrfStore     csrf.Store      // Backend for persisting/verifying CSRF tokens.
}

// ControllerI defines the public contract for the client-side user operations.
type ControllerI interface {
	Create(c *gin.Context)       // Admin-only user creation.
	User(c *gin.Context)         // Fetch profile for the current user.
	Users(c *gin.Context)        // List and filter users (Admin).
	Update(c *gin.Context)       // Modify profile details.
	Delete(c *gin.Context)       // Account deactivation/deletion.
	SignIn(c *gin.Context)       // Authenticate and issue tokens.
	SignUp(c *gin.Context)       // Register a new account.
	SignOut(c *gin.Context)      // Invalidate session.
	CsrfToken(c *gin.Context)    // Issue a fresh CSRF token for the frontend.
	RefreshToken(c *gin.Context) // Use refresh token to rotate access tokens.
}

// New initializes a User Client controller.
// It pre-parses the JWT private key and configures CSRF protection based on application settings.
func New(u *usecase.UseCase, cfg *config.Config, l logger.Log) ControllerI {
	// Pre-parse the private key for signing tokens once during startup.
	pk, err := jwt.ParseRSAPrivateKey(cfg.JWT.PrivateKey)
	if err != nil {
		l.Fatal("ClientController - New - parsedPrivateKey error", err)
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
