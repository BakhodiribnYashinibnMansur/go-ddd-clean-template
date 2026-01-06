package minio

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"

	"gct/config"

	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	minioClient "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	testRepo *Repo
	testCtx  context.Context
	ts       *httptest.Server
)

func TestMain(m *testing.M) {
	testCtx = context.Background()

	// Setup fake S3
	backend := s3mem.New()
	faker := gofakes3.New(backend)
	ts = httptest.NewServer(faker.Server())
	defer ts.Close()

	// MinIO configuration for testing
	cfg := config.MinioStore{
		Endpoint:  ts.URL, // httptest server URL (http://127.0.0.1:xxxxx)
		AccessKey: "YOUR_ACCESS_KEY",
		SecretKey: "YOUR_SECRET_KEY",
		UseSSL:    false,
		Bucket:    "test-bucket",
	}

	// Remove http:// prefix from endpoint because minio New() expects "host:port" or "host"
	// but httptest returns "http://host:port".
	// However, minio-go handles this? No, it expects endpoint without scheme if Secure is false?
	// Actually, if we pass scheme, it might confuse it or it might handle it.
	// Let's strip it to be safe, or check documentation.
	// minio.New(endpoint, ...) -> "Endpoint:  127.0.0.1:9000"

	endpoint := ts.URL[len("http://"):]

	// Create MinIO client connected to fake S3
	client, err := minioClient.New(endpoint, &minioClient.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		panic(err)
	}

	// Create bucket
	err = client.MakeBucket(testCtx, cfg.Bucket, minioClient.MakeBucketOptions{})
	if err != nil {
		panic(err)
	}

	// Use the fake s3 backed client
	testRepo = New(client, &cfg)

	code := m.Run()

	os.Exit(code)
}
