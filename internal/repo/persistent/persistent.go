package persistent

import (
	minioClient "github.com/minio/minio-go/v7"
	redisClient "github.com/redis/go-redis/v9"

	"gct/config"
	"gct/internal/repo/persistent/minio"
	"gct/internal/repo/persistent/postgres/user"
	"gct/internal/repo/persistent/redis"
	dbPostgres "gct/pkg/db/postgres"
	"gct/pkg/logger"
)

type Repo struct {
	Postgres *user.User
	MinIO    *minio.Repo
	Redis    *redis.Repo
}

func New(pg *dbPostgres.Postgres, mClient *minioClient.Client, rClient *redisClient.Client, mConfig *config.MinioStore, logger logger.Log) *Repo {
	return &Repo{
		Postgres: user.New(pg, logger),
		MinIO:    minio.New(mClient, mConfig),
		Redis:    redis.New(rClient, logger),
	}
}
