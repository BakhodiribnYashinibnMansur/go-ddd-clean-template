package container

import (
	"context"
	"log"

	minio_client "github.com/minio/minio-go/v7"
	minio_credentials "github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/minio"

	"gct/config"
)

// RunMinioTestContainer is a function that runs a minio test container
func RunMinioTestContainer(cfg config.MinioStore) *minio_client.Client {
	ctx := context.Background()

	minioContainer, err := minio.RunContainer(ctx,
		testcontainers.WithImage(MinioImage),
	)
	if err != nil {
		log.Fatalf("failed to start minio container: %v", err)
	}

	endpoint, err := minioContainer.Endpoint(ctx, "")
	if err != nil {
		log.Fatalf("failed to get minio endpoint: %v", err)
	}

	// Create a new MinIO client
	client, err := minio_client.New(endpoint, &minio_client.Options{
		Creds:  minio_credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		log.Fatalf("failed to create minio client: %v", err)
	}

	// Create bucket if it doesn't exist
	err = client.MakeBucket(ctx, cfg.Bucket, minio_client.MakeBucketOptions{})
	if err != nil {
		// Check if bucket already exists
		exists, err := client.BucketExists(ctx, cfg.Bucket)
		if err != nil {
			log.Fatalf("failed to check bucket existence: %v", err)
		}
		if !exists {
			log.Fatalf("failed to create bucket: %v", err)
		}
	}

	return client
}
