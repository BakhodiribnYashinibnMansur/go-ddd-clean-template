package user

import (
	"github.com/evrone/go-clean-template/internal/repo/persistent/postgres/user/client"
	"github.com/evrone/go-clean-template/pkg/db/postgres"
	"go.uber.org/zap"
)

// User aggregates user-related repositories.
type User struct {
	client.UserRepoI
	// Add other interfaces here as needed (Staff, Notification, Session)
}

// NewUserRepo creates a new User repository aggregating sub-repositories.
func NewUserRepo(pg *postgres.Postgres, logger *zap.Logger) *User {
	return &User{
		UserRepoI: client.NewUserRepo(pg, logger),
	}
}
