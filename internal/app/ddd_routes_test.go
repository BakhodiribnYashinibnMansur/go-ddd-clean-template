package app

import (
	"testing"

	miniogo "github.com/minio/minio-go/v7"
)

func TestRouteOptions_ZeroValue(t *testing.T) {
	opt := RouteOptions{}
	if opt.Minio != nil {
		t.Fatal("expected nil Minio client")
	}
	if opt.MinioBucket != "" {
		t.Fatal("expected empty MinioBucket")
	}
}

func TestRouteOptions_WithValues(t *testing.T) {
	opt := RouteOptions{
		Minio:       &miniogo.Client{},
		MinioBucket: "test-bucket",
	}
	if opt.Minio == nil {
		t.Fatal("expected non-nil Minio client")
	}
	if opt.MinioBucket != "test-bucket" {
		t.Fatalf("expected test-bucket, got %s", opt.MinioBucket)
	}
}
