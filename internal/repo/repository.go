package repo

import (
	minioClient "github.com/minio/minio-go/v7"
	redisClient "github.com/redis/go-redis/v9"

	"gct/config"
	"gct/internal/repo/integration/rest"
	"gct/internal/repo/persistent"
	"gct/pkg/db/postgres"
	"gct/pkg/logger"
)

// Repo represents the main repository structure
type Repo struct {
	Persistent *persistent.Repo
	Client     *rest.Client
}

// New creates a new repository instance
func New(pg *postgres.Postgres, mClient *minioClient.Client, rClient *redisClient.Client, mConfig *config.MinioStore, logger logger.Log) *Repo {
	return &Repo{
		Persistent: persistent.New(pg, mClient, rClient, mConfig, logger),
		Client:     rest.New(30),
	}
}
