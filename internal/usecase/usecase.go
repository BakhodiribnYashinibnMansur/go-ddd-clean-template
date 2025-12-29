package usecase

import (
	"gct/config"
	"gct/internal/repo"
	"gct/internal/usecase/minio"
	"gct/internal/usecase/user"
	"gct/pkg/logger"
)

// UseCase -.
type UseCase struct {
	User  *user.UseCase
	Minio *minio.UseCase
}

// NewUseCase -.
func NewUseCase(repos *repo.Repo, logger logger.Log, cfg *config.Config) *UseCase {
	return &UseCase{
		User:  user.New(repos, logger, cfg),
		Minio: minio.New(repos, logger),
	}
}
