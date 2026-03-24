package user

import (
	"gct/internal/repo/persistent/postgres/user/client"
	"gct/internal/repo/persistent/postgres/user/session"
	"gct/internal/repo/persistent/postgres/user/setting"
	"gct/internal/shared/infrastructure/db/postgres"
	"gct/internal/shared/infrastructure/logger"
)

// User aggregates user-related repositories.
type User struct {
	Client      client.RepoI
	SessionRepo session.RepoI
	Setting     setting.RepoI
}

// NewUserRepo creates a new User repository aggregating sub-repositories.
func New(psql *postgres.Postgres, logger logger.Log) *User {
	return &User{
		Client:      client.New(psql, logger),
		SessionRepo: session.New(psql, logger),
		Setting:     setting.New(psql.Pool),
	}
}
