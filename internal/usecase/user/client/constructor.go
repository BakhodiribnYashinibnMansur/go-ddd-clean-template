package client

import (
	"crypto/rsa"

	"gct/config"
	"gct/internal/repo/persistent"
	"gct/pkg/jwt"
	"gct/pkg/logger"
)

// UseCase -.
type UseCase struct {
	repo       *persistent.Repo
	logger     logger.Log
	cfg        *config.Config
	privateKey *rsa.PrivateKey
}

// New -.
func New(r *persistent.Repo, logger logger.Log, cfg *config.Config) UseCaseI {
	pk, err := jwt.ParseRSAPrivateKey(cfg.JWT.PrivateKey)
	if err != nil {
		// Log error at constructor level but don't return it
		// The use case can still function without private key for some operations
	}
	return &UseCase{
		repo:       r,
		logger:     logger,
		cfg:        cfg,
		privateKey: pk,
	}
}
