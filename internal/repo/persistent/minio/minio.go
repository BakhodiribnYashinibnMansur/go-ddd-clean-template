package minio

import (
	"context"
	"io"

	"gct/config"
	"gct/internal/domain"
	"github.com/minio/minio-go/v7"
)

// RepoI defines the interface for MinIO storage operations
type RepoI interface {
	// Upload operations
	UploadImage(ctx context.Context, file io.Reader, fileSize int64, contentType string) (string, error)
	UploadDocument(ctx context.Context, file io.Reader, fileSize int64, contentType string) (string, error)
	UploadVideo(ctx context.Context, file io.Reader, fileSize int64, contentType string) (string, error)
	UploadFile(ctx context.Context, filePath, contentType string) (string, error)

	// Download operations
	GetFileURL(ctx context.Context, fileName string) (string, error)
	GetFileURLs(ctx context.Context, files []domain.File) ([]domain.File, error)

	// Object operations
	ObjectExists(ctx context.Context, fileName string) error
	DeleteFile(ctx context.Context, fileName string) error

	// Health check
	HealthCheck(ctx context.Context) error
}

// Repo implements MinIO storage operations
type Repo struct {
	client *minio.Client
	config *config.MinioStore
}

// New creates a new MinIO repository
func New(client *minio.Client, cfg *config.MinioStore) *Repo {
	return &Repo{
		client: client,
		config: cfg,
	}
}

// HealthCheck verifies MinIO connection
func (r *Repo) HealthCheck(ctx context.Context) error {
	exists, err := r.client.BucketExists(ctx, r.config.Bucket)
	if err != nil {
		return err
	}

	if !exists {
		return domain.ErrBucketNotFound
	}

	return nil
}
