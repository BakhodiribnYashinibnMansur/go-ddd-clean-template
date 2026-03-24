package persistent

import (
	"gct/config"
	"gct/internal/repo/persistent/minio"
	"gct/internal/repo/persistent/postgres"
	"gct/internal/repo/persistent/redis"
	dbPostgres "gct/internal/shared/infrastructure/db/postgres"
	"gct/internal/shared/infrastructure/logger"
	minioclient "github.com/minio/minio-go/v7"
	redisclient "github.com/redis/go-redis/v9"
)

type Repo struct {
	Postgres *postgres.Repo
	MinIO    *minio.Repo
	Redis    *redis.Repo
}

func New(pg *dbPostgres.Postgres, mClient *minioclient.Client, rClient *redisclient.Client, mConfig *config.MinioStore, logger logger.Log) *Repo {
	pgRepo, err := postgres.New(pg, logger)
	if err != nil {
		panic("failed to initialize postgres repo: " + err.Error())
	}
	return &Repo{
		Postgres: pgRepo,
		MinIO:    minio.New(mClient, mConfig),
		Redis:    redis.New(rClient, logger),
	}
}
