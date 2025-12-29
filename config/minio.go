package config

// MinioStore configuration for MinIO object storage
type MinioStore struct {
	Endpoint  string `env:"ENDPOINT"   json:"endpoint"`
	AccessKey string `env:"ACCESS_KEY" json:"access_key"`
	SecretKey string `env:"SECRET_KEY" json:"secret_key"`
	UseSSL    bool   `env:"USE_SSL"    json:"use_ssl"`
	Region    string `env:"REGION"     json:"region"`
	Bucket    string `env:"BUCKET"     json:"bucket"`
}
