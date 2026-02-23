package usecase

import (
	"gct/config"
	"gct/internal/repo"
	"gct/internal/usecase/audit"
	"gct/internal/usecase/authz"
	"gct/internal/usecase/database"
	errorcode "gct/internal/usecase/errorcode"
	"gct/internal/usecase/integration"
	"gct/internal/usecase/minio"
	"gct/internal/usecase/sitesetting"
	"gct/internal/usecase/translation"
	"gct/internal/usecase/user"
	"gct/pkg/asynq"
	"gct/pkg/logger"
)

// NewUseCase -.
func NewUseCase(repos *repo.Repo, logger logger.Log, cfg *config.Config, asynqClient *asynq.Client) *UseCase {
	return &UseCase{
		Repo:        repos,
		User:        user.New(repos, logger, cfg),
		Minio:       minio.New(repos, logger),
		Authz:       authz.New(repos, logger, cfg),
		Audit:       audit.New(repos.Persistent, logger),
		SiteSetting: sitesetting.New(repos.Persistent, logger),
		ErrorCode:   errorcode.New(repos, logger),
		Database:    database.New(repos.Persistent.Postgres, logger, cfg),
		Integration: integration.New(repos.Persistent.Postgres.Integration, logger, cfg),
		Translation: translation.New(repos.Persistent.Postgres.Translation, logger),
		AsynqClient: asynqClient,
	}
}
