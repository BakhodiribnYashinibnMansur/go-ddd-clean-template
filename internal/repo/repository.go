package repo

import (
	"time"

	"gct/config"
	"gct/internal/repo/integration/rest"
	"gct/internal/repo/persistent"
	"gct/internal/shared/infrastructure/db/postgres"
	"gct/internal/shared/infrastructure/logger"

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

const (
	// DefaultRESTTimeout is the default timeout for the REST client
	DefaultRESTTimeout = 30 * time.Second
)

// New creates a new repository instance with both persistent and integration layers
func New(pg *postgres.Postgres, mClient *minio.Client, rClient *redis.Client, mConfig *config.MinioStore, logger logger.Log) *Repo {
	return &Repo{
		Persistent: persistent.New(pg, mClient, rClient, mConfig, logger),
		Client:     rest.New(DefaultRESTTimeout),
	}
}

// Getter methods for backward compatibility and clean access
func (r *Repo) User() *persistent.Repo {
	return r.Persistent
}
