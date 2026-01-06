package container

import (
	"context"
	"fmt"

	"gct/config"

	minio_client "github.com/minio/minio-go/v7"
	minio_credentials "github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/minio"
)

// RunMinioTestContainer is a function that runs a minio test container
// RunMinioTestContainer runs a minio test container
func RunMinioTestContainer(cfg config.MinioStore) (*minio_client.Client, testcontainers.Container, error) {
	ctx := context.Background()

	minioContainer, err := minio.RunContainer(ctx,
		testcontainers.WithImage(MinioImage),
		testcontainers.WithEnv(map[string]string{
			"MINIO_ROOT_USER":     cfg.AccessKey,
			"MINIO_ROOT_PASSWORD": cfg.SecretKey,
		}),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start minio container: %w", err)
	}

	endpoint, err := minioContainer.Endpoint(ctx, "")
	if err != nil {
		return nil, minioContainer, fmt.Errorf("failed to get minio endpoint: %w", err)
	}

	// Create a new MinIO client
	client, err := minio_client.New(endpoint, &minio_client.Options{
		Creds:  minio_credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, minioContainer, fmt.Errorf("failed to create minio client: %w", err)
	}

	// Create bucket if it doesn't exist
	err = client.MakeBucket(ctx, cfg.Bucket, minio_client.MakeBucketOptions{})
	if err != nil {
		// Check if bucket already exists
		exists, err := client.BucketExists(ctx, cfg.Bucket)
		if err != nil {
			return nil, minioContainer, fmt.Errorf("failed to check bucket existence: %w", err)
		}
		if !exists {
			return nil, minioContainer, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return client, minioContainer, nil
}
