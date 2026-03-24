package client

import (
	"crypto/rsa"

	"gct/config"
	"gct/internal/repo/persistent"
	"gct/internal/shared/infrastructure/security/jwt"
	"gct/internal/shared/infrastructure/logger"
)

// UseCase -.
type UseCase struct {
	repo       *persistent.Repo
	logger     logger.Log
	cfg        *config.Config
	privateKey *rsa.PrivateKey
}

// New -.
func New(r *persistent.Repo, l logger.Log, cfg *config.Config) UseCaseI {
	// Parse private key for JWT signing
	privKey, err := jwt.ParseRSAPrivateKey(cfg.JWT.PrivateKey)
	if err != nil {
		l.Fatalw("failed to parse RSA private key", "error", err)
	}

	return &UseCase{
		repo:       r,
		logger:     l,
		cfg:        cfg,
		privateKey: privKey,
	}
}
