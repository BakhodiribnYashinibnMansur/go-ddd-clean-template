package usecase

import (
	"gct/config"
	"gct/internal/repo"
	"gct/internal/usecase/audit"
	"gct/internal/usecase/authz"
	"gct/internal/usecase/minio"
	"gct/internal/usecase/user"
	"gct/pkg/logger"
)

// UseCase -.
type UseCase struct {
	User  *user.UseCase
	Minio *minio.UseCase
	Authz *authz.UseCase
	Audit *audit.UseCase
}

// NewUseCase -.
func NewUseCase(repos *repo.Repo, logger logger.Log, cfg *config.Config) *UseCase {
	return &UseCase{
		User:  user.New(repos, logger, cfg),
		Minio: minio.New(repos, logger),
		Authz: authz.New(repos, logger, cfg),
		Audit: audit.New(repos.Persistent, logger),
	}
}
