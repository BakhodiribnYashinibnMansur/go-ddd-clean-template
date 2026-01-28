package user

import (
	"gct/config"
	"gct/internal/repo"
	"gct/internal/usecase/user/client"
	"gct/internal/usecase/user/session"
	"gct/pkg/logger"
)

type UseCaseI interface {
	Client() client.UseCaseI
	Session() session.UseCaseI
}

type UseCase struct {
	client  client.UseCaseI
	session session.UseCaseI
}

func New(r *repo.Repo, logger logger.Log, cfg *config.Config) UseCaseI {
	return &UseCase{
		client:  client.New(r.Persistent, logger, cfg),
		session: session.New(r.Persistent, logger),
	}
}

func (uc *UseCase) Client() client.UseCaseI   { return uc.client }
func (uc *UseCase) Session() session.UseCaseI { return uc.session }
