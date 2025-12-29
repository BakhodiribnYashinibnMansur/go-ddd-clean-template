package minio

import (
	"context"
	"fmt"
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Option func(*minioOptions)

type minioOptions struct {
	minioOpts    *minio.Options
	bucket       string
	bucketRegion string // Region for bucket creation
	makeBucket   bool
}

func WithCredentials(accessKey, secretKey string) Option {
	return func(o *minioOptions) {
		o.minioOpts.Creds = credentials.NewStaticV4(accessKey, secretKey, "")
	}
}

// WithTokenCredentials allows providing a session token
func WithTokenCredentials(accessKey, secretKey, token string) Option {
	return func(o *minioOptions) {
		o.minioOpts.Creds = credentials.NewStaticV4(accessKey, secretKey, token)
	}
}

func WithSecure(secure bool) Option {
	return func(o *minioOptions) {
		o.minioOpts.Secure = secure
	}
}

// WithRegion sets the client region (used for signature v4)
func WithRegion(region string) Option {
	return func(o *minioOptions) {
		o.minioOpts.Region = region
	}
}

func WithTransport(transport http.RoundTripper) Option {
	return func(o *minioOptions) {
		o.minioOpts.Transport = transport
	}
}

func WithBucket(bucket, region string) Option {
	return func(o *minioOptions) {
		o.bucket = bucket
		o.bucketRegion = region
		o.makeBucket = true
	}
}

// New initializes a new MinIO client.
func New(endpoint string, opts ...Option) (*minio.Client, error) {
	options := &minioOptions{
		minioOpts: &minio.Options{
			Secure: true,
		},
	}

	for _, opt := range opts {
		opt(options)
	}

	minioClient, err := minio.New(endpoint, options.minioOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize minio client: %w", err)
	}

	if options.makeBucket && options.bucket != "" {
		ctx := context.Background()
		exists, err := minioClient.BucketExists(ctx, options.bucket)
		if err != nil {
			return nil, fmt.Errorf("failed to check bucket existence: %w", err)
		}

		if !exists {
			region := options.bucketRegion
			if region == "" {
				region = options.minioOpts.Region
			}
			err = minioClient.MakeBucket(ctx, options.bucket, minio.MakeBucketOptions{Region: region})
			if err != nil {
				return nil, fmt.Errorf("failed to create bucket: %w", err)
			}
		}
	}

	return minioClient, nil
}
