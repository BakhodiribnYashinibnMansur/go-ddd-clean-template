package client

import (
	"github.com/evrone/go-clean-template/pkg/db/postgres"
	"go.uber.org/zap"
)

// UserRepo handles user-related database operations.
type UserRepo struct {
	*postgres.Postgres
	logger *zap.Logger
}

// NewUserRepo creates a new UserRepo instance.
func NewUserRepo(pg *postgres.Postgres, logger *zap.Logger) *UserRepo {
	return &UserRepo{
		Postgres: pg,
		logger:   logger,
	}
}
