package persistent

import (
	"github.com/evrone/go-clean-template/internal/repo/persistent/postgres/user"
	"github.com/evrone/go-clean-template/pkg/db/postgres"
	"github.com/evrone/go-clean-template/pkg/logger"
)

type Repo struct {
	User *user.User
}

func New(pg *postgres.Postgres, logger logger.Log) *Repo {
	return &Repo{
		User: user.New(pg, logger),
	}
}
