package minio_test

import (
	"net/http/httptest"
	"testing"

	"gct/config"
	"gct/internal/repo"
	"gct/internal/repo/persistent"
	minioRepo "gct/internal/repo/persistent/minio"
	"gct/internal/usecase/minio"
	"gct/pkg/logger"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	minioClient "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func setup(t *testing.T) *minio.UseCase {
	t.Helper()
	// Setup fake S3
	backend := s3mem.New()
	faker := gofakes3.New(backend)
	ts := httptest.NewServer(faker.Server())
	t.Cleanup(ts.Close)

	// MinIO configuration
	cfg := config.MinioStore{
		Endpoint:  ts.URL,
		AccessKey: "YOUR_ACCESS_KEY",
		SecretKey: "YOUR_SECRET_KEY",
		UseSSL:    false,
		Bucket:    "test-bucket",
	}

	endpoint := ts.URL[len("http://"):]

	// Create MinIO client
	client, err := minioClient.New(endpoint, &minioClient.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		t.Fatalf("failed to create minio client: %v", err)
	}

	// Create bucket
	err = client.MakeBucket(t.Context(), cfg.Bucket, minioClient.MakeBucketOptions{})
	if err != nil {
		t.Fatalf("failed to create bucket: %v", err)
	}

	// Construct Repos
	mRepo := minioRepo.New(client, &cfg)

	// Create persistent Repo manually to avoid needing Postgres/Redis connections
	pRepo := &persistent.Repo{
		MinIO: mRepo,
	}

	r := &repo.Repo{
		Persistent: pRepo,
	}

	log := logger.New("debug")

	return minio.New(r, log)
}
