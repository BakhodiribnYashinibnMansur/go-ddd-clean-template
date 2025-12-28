package client

import (
	"crypto/rsa"

	"github.com/evrone/go-clean-template/config"
	"github.com/evrone/go-clean-template/internal/repo/persistent"
	"github.com/evrone/go-clean-template/pkg/jwt"
	"github.com/evrone/go-clean-template/pkg/logger"
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
		logger.Error("ClientUseCase - New - parsedPrivateKey error", err)
	}
	return &UseCase{
		repo:       r,
		logger:     logger,
		cfg:        cfg,
		privateKey: pk,
	}
}
