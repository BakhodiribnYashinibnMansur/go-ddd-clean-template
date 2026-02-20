package usecase

import (
	"gct/internal/repo"
	"gct/internal/usecase/audit"
	"gct/internal/usecase/authz"
	"gct/internal/usecase/database"
	errorcode "gct/internal/usecase/errorcode"
	"gct/internal/usecase/integration"
	"gct/internal/usecase/minio"
	"gct/internal/usecase/sitesetting"
	"gct/internal/usecase/user"
	"gct/pkg/asynq"
)

// UseCase -.
type UseCase struct {
	Repo        *repo.Repo
	User        user.UseCaseI
	Minio       minio.Interface
	Authz       authz.UseCaseI
	Audit       audit.UseCaseI
	SiteSetting sitesetting.UseCaseI
	ErrorCode   errorcode.UseCaseI
	Database    database.UseCaseI
	Integration integration.UseCaseI
	AsynqClient *asynq.Client
}
