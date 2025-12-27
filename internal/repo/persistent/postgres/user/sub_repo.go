package user

import (
	"github.com/evrone/go-clean-template/internal/repo/persistent/postgres/user/client"
	"github.com/evrone/go-clean-template/internal/repo/persistent/postgres/user/session"
	"github.com/evrone/go-clean-template/pkg/db/postgres"
	"github.com/evrone/go-clean-template/pkg/logger"
)

// User aggregates user-related repositories.
type User struct {
	Client      client.RepoI
	SessionRepo session.RepoI
	// Add other interfaces here as needed (Staff, Notification)
}

// NewUserRepo creates a new User repository aggregating sub-repositories.
func New(psql *postgres.Postgres, logger logger.Log) *User {
	return &User{
		Client:      client.New(psql, logger),
		SessionRepo: session.New(psql, logger),
	}
}
