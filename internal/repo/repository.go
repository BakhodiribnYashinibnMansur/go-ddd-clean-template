package repo

import (
	"github.com/evrone/go-clean-template/internal/repo/integration/rest"
	"github.com/evrone/go-clean-template/internal/repo/persistent"
	"github.com/evrone/go-clean-template/pkg/db/postgres"
	"github.com/evrone/go-clean-template/pkg/logger"
)

// Repo represents the main repository structure that aggregates
// both persistent storage and external integration repositories
type Repo struct {
	// Persistent layer repositories
	Persistent *persistent.Repo

	// Integration layer repositories
	Client *rest.Client
	// Add other integration repos here (Kafka, Redis, etc.)
}

// New creates a new repository instance with both persistent and integration layers
func New(pg *postgres.Postgres, logger logger.Log) *Repo {
	return &Repo{
		Persistent: persistent.New(pg, logger),
		Client:     rest.New(30), // 30 second timeout for REST client
	}
}

// Getter methods for backward compatibility and clean access
func (r *Repo) User() *persistent.Repo {
	return r.Persistent
}
