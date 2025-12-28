package user

import (
	"github.com/evrone/go-clean-template/config"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/internal/usecase/user/client"
	"github.com/evrone/go-clean-template/internal/usecase/user/session"
	"github.com/evrone/go-clean-template/pkg/logger"
)

type User struct {
	Client  client.UseCaseI
	Session session.UseCaseI
}

func New(r *repo.Repo, logger logger.Log, cfg *config.Config) *User {
	return &User{
		Client:  client.New(r.Persistent, logger, cfg),
		Session: session.New(r.Persistent, logger),
	}
}
