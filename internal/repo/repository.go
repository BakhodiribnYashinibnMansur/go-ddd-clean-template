package repo

import (
	"gct/config"
	"gct/internal/repo/integration/rest"
	"gct/internal/repo/persistent"
	"gct/pkg/db/postgres"
	"gct/pkg/logger"

	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
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
func New(pg *postgres.Postgres, mClient *minio.Client, rClient *redis.Client, mConfig *config.MinioStore, logger logger.Log) *Repo {
	return &Repo{
		Persistent: persistent.New(pg, mClient, rClient, mConfig, logger),
		Client:     rest.New(30), // 30 second timeout for REST client
	}
}

// Getter methods for backward compatibility and clean access
func (r *Repo) User() *persistent.Repo {
	return r.Persistent
}
