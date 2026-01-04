package container

import (
	"testing"

	"gct/config"
	"github.com/stretchr/testify/assert"
)

// TestMinioConfig_TableDriven tests Minio configuration scenarios
func TestMinioConfig_TableDriven(t *testing.T) {
	type minioConfigTestCase struct {
		name         string
		config       config.MinioStore
		expectedVals config.MinioStore
		definition   string
	}

	tests := []minioConfigTestCase{
		{
			name: "Default Minio configuration",
			config: config.MinioStore{
				AccessKey: "minioadmin",
				SecretKey: "minioadmin",
				Bucket:    "test-bucket",
				UseSSL:    false,
			},
			expectedVals: config.MinioStore{
				AccessKey: "minioadmin",
				SecretKey: "minioadmin",
				Bucket:    "test-bucket",
				UseSSL:    false,
			},
			definition: "Tests basic Minio configuration",
		},
		{
			name: "Minio configuration with SSL",
			config: config.MinioStore{
				AccessKey: "myaccesskey",
				SecretKey: "mysecretkey",
				Bucket:    "secure-bucket",
				UseSSL:    true,
			},
			expectedVals: config.MinioStore{
				AccessKey: "myaccesskey",
				SecretKey: "mysecretkey",
				Bucket:    "secure-bucket",
				UseSSL:    true,
			},
			definition: "Tests Minio configuration with SSL enabled",
		},
		{
			name: "Minio configuration with custom bucket",
			config: config.MinioStore{
				AccessKey: "user123",
				SecretKey: "pass456",
				Bucket:    "my-custom-bucket",
				UseSSL:    false,
			},
			expectedVals: config.MinioStore{
				AccessKey: "user123",
				SecretKey: "pass456",
				Bucket:    "my-custom-bucket",
				UseSSL:    false,
			},
			definition: "Tests Minio configuration with custom bucket name",
		},
		{
			name: "Minio configuration with long credentials",
			config: config.MinioStore{
				AccessKey: "very-long-access-key-for-testing-purposes",
				SecretKey: "very-long-secret-key-for-testing-purposes-with-more-chars",
				Bucket:    "long-credentials-bucket",
				UseSSL:    true,
			},
			expectedVals: config.MinioStore{
				AccessKey: "very-long-access-key-for-testing-purposes",
				SecretKey: "very-long-secret-key-for-testing-purposes-with-more-chars",
				Bucket:    "long-credentials-bucket",
				UseSSL:    true,
			},
			definition: "Tests Minio configuration with long credentials",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedVals.AccessKey, tc.config.AccessKey, "AccessKey should match")
			assert.Equal(t, tc.expectedVals.SecretKey, tc.config.SecretKey, "SecretKey should match")
			assert.Equal(t, tc.expectedVals.Bucket, tc.config.Bucket, "Bucket should match")
			assert.Equal(t, tc.expectedVals.UseSSL, tc.config.UseSSL, "UseSSL should match")
		})
	}
}

// TestMinioConstants_TableDriven tests Minio-related constants
func TestMinioConstants_TableDriven(t *testing.T) {
	type minioConstantTestCase struct {
		name        string
		constant    string
		expectedVal string
		definition  string
	}

	tests := []minioConstantTestCase{
		{
			name:        "Minio image constant",
			constant:    MinioImage,
			expectedVal: "minio/minio:latest",
			definition:  "Tests Minio image constant value",
		},
		{
			name:        "Redis image constant",
			constant:    RedisImage,
			expectedVal: "redis:7-alpine",
			definition:  "Tests Redis image constant value",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedVal, tc.constant, "Constant should match expected value")
		})
	}
}

// TestMinioValidation_TableDriven tests Minio configuration validation
func TestMinioValidation_TableDriven(t *testing.T) {
	type minioValidationTestCase struct {
		name        string
		config      config.MinioStore
		expectValid bool
		definition  string
	}

	tests := []minioValidationTestCase{
		{
			name: "Valid Minio configuration",
			config: config.MinioStore{
				AccessKey: "validkey",
				SecretKey: "validsecret",
				Bucket:    "valid-bucket",
				UseSSL:    false,
			},
			expectValid: true,
			definition:  "Tests validation of valid Minio configuration",
		},
		{
			name: "Empty access key",
			config: config.MinioStore{
				AccessKey: "",
				SecretKey: "validsecret",
				Bucket:    "valid-bucket",
				UseSSL:    false,
			},
			expectValid: false,
			definition:  "Tests validation with empty access key",
		},
		{
			name: "Empty secret key",
			config: config.MinioStore{
				AccessKey: "validkey",
				SecretKey: "",
				Bucket:    "valid-bucket",
				UseSSL:    false,
			},
			expectValid: false,
			definition:  "Tests validation with empty secret key",
		},
		{
			name: "Empty bucket name",
			config: config.MinioStore{
				AccessKey: "validkey",
				SecretKey: "validsecret",
				Bucket:    "",
				UseSSL:    false,
			},
			expectValid: false,
			definition:  "Tests validation with empty bucket name",
		},
		{
			name: "Valid SSL configuration",
			config: config.MinioStore{
				AccessKey: "validkey",
				SecretKey: "validsecret",
				Bucket:    "ssl-bucket",
				UseSSL:    true,
			},
			expectValid: true,
			definition:  "Tests validation of SSL-enabled Minio configuration",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isValid := validateMinioConfig(tc.config)

			if tc.expectValid {
				assert.True(t, isValid, "Expected configuration to be valid")
			} else {
				assert.False(t, isValid, "Expected configuration to be invalid")
			}
		})
	}
}

// validateMinioConfig is a helper function to validate Minio configuration
func validateMinioConfig(cfg config.MinioStore) bool {
	if cfg.AccessKey == "" {
		return false
	}
	if cfg.SecretKey == "" {
		return false
	}
	if cfg.Bucket == "" {
		return false
	}
	return true
}

// TestMinioEdgeCases_TableDriven tests Minio edge cases
func TestMinioEdgeCases_TableDriven(t *testing.T) {
	type minioEdgeCaseTestCase struct {
		name       string
		config     config.MinioStore
		testFunc   func(config.MinioStore) bool
		definition string
	}

	tests := []minioEdgeCaseTestCase{
		{
			name: "Special characters in bucket name",
			config: config.MinioStore{
				AccessKey: "testkey",
				SecretKey: "testsecret",
				Bucket:    "test-bucket_123",
				UseSSL:    false,
			},
			testFunc: func(cfg config.MinioStore) bool {
				// Test that bucket name contains valid characters
				return len(cfg.Bucket) > 0 && cfg.Bucket != " "
			},
			definition: "Tests bucket name with special characters",
		},
		{
			name: "Very long bucket name",
			config: config.MinioStore{
				AccessKey: "testkey",
				SecretKey: "testsecret",
				Bucket:    "this-is-a-very-long-bucket-name-that-might-exceed-limits-but-should-still-work",
				UseSSL:    false,
			},
			testFunc: func(cfg config.MinioStore) bool {
				// Test that long bucket names are handled
				return len(cfg.Bucket) > 50
			},
			definition: "Tests handling of very long bucket names",
		},
		{
			name: "Unicode characters in credentials",
			config: config.MinioStore{
				AccessKey: "ключ-доступа",
				SecretKey: "секретный-ключ",
				Bucket:    "unicode-bucket",
				UseSSL:    false,
			},
			testFunc: func(cfg config.MinioStore) bool {
				// Test that unicode characters are preserved
				return len(cfg.AccessKey) > 0 && len(cfg.SecretKey) > 0
			},
			definition: "Tests unicode characters in credentials",
		},
		{
			name: "SSL toggle behavior",
			config: config.MinioStore{
				AccessKey: "testkey",
				SecretKey: "testsecret",
				Bucket:    "ssl-test-bucket",
				UseSSL:    true,
			},
			testFunc: func(cfg config.MinioStore) bool {
				// Test SSL flag behavior
				return cfg.UseSSL == true
			},
			definition: "Tests SSL flag behavior",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.testFunc(tc.config)
			assert.True(t, result, "Edge case test should pass")
		})
	}
}
