package user

import (
	"gct/config"
	"gct/internal/repo"
	"gct/internal/usecase/user/client"
	"gct/internal/usecase/user/session"
	"gct/pkg/logger"
)

type UseCase struct {
	Client  client.UseCaseI
	Session session.UseCaseI
}

func New(r *repo.Repo, logger logger.Log, cfg *config.Config) *UseCase {
	return &UseCase{
		Client:  client.New(r.Persistent, logger, cfg),
		Session: session.New(r.Persistent, logger),
	}
}
