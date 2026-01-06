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

type Controller struct {
	u             *usecase.UseCase
	l             logger.Log
	cfg           *config.Config
	privateKey    *rsa.PrivateKey
	csrfGenerator *csrf.Generator
	csrfStore     csrf.Store
}

type ControllerI interface {
	Create(c *gin.Context)
	User(c *gin.Context)
	Users(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	SignIn(c *gin.Context)
	SignUp(c *gin.Context)
	SignOut(c *gin.Context)
	CsrfToken(c *gin.Context)
	RefreshToken(c *gin.Context)
}

func New(u *usecase.UseCase, cfg *config.Config, l logger.Log) ControllerI {
	pk, err := jwt.ParseRSAPrivateKey(cfg.JWT.PrivateKey)
	if err != nil {
		l.Fatal("ClientController - New - parsedPrivateKey error", err)
	}

	// Initialize CSRF protection with dedicated secret
	csrfGen := csrf.NewGenerator(csrf.Config{
		Secret:     []byte(cfg.App.CSRFSecret),
		Expiration: csrf.DefaultExpiration,
	})
	csrfStore := csrf.NewMemoryStore() // Use memory store (can be replaced with Redis)

	return &Controller{
		u:             u,
		l:             l,
		cfg:           cfg,
		privateKey:    pk,
		csrfGenerator: csrfGen,
		csrfStore:     csrfStore,
	}
}
