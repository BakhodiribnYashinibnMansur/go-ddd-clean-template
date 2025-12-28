package config

// MinioStore configuration for MinIO object storage
type MinioStore struct {
	Endpoint  string `env:"MINIO_ENDPOINT" json:"endpoint"`
	AccessKey string `env:"MINIO_ACCESS_KEY" json:"access_key"`
	SecretKey string `env:"MINIO_SECRET_KEY" json:"secret_key"`
	UseSSL    bool   `env:"MINIO_USE_SSL" json:"use_ssl"`
	Region    string `env:"MINIO_REGION" json:"region"`
	Bucket    string `env:"MINIO_BUCKET" json:"bucket"`
}
