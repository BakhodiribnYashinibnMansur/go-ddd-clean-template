package config

import (
	"os"
	"reflect"
	"sync"
	"time"
)

var (
	testInstance *Config
	testOnce     sync.Once
)

// NewTestConfig returns app config for testing (Singleton).
// It fills the config with dummy data suitable for testing.
func NewTestConfig() (*Config, error) {
	testOnce.Do(func() {
		cfg := &Config{
			App: App{
				Name:        "go-clean-template-test",
				Version:     "1.0.0",
				Environment: "test",
				CSRFSecret:  "test-csrf-secret-key-for-testing-min-32-chars",
			},
			HTTP: HTTP{
				Port:           "8080",
				UsePreforkMode: false,
			},
			Log: Log{
				Level: "debug",
			},
			Database: Database{
				Postgres: Postgres{
					BaseDB: BaseDB{
						Host:     "localhost",
						Port:     5432,
						Name:     "test_db",
						User:     "test_user",
						Password: "test_password",
						SSLMode:  "disable",
						PoolMax:  10,
					},
				},
				Redis: Redis{BaseDB: BaseDB{Host: "localhost", Port: 6379, Name: "1"}},
			},
			Redis: RedisStore{
				Host:     "localhost",
				Port:     "6379",
				Password: "",
				DB:       0,
			},
			Minio: MinioStore{
				Endpoint:  "localhost:9000",
				AccessKey: "test_access_key",
				SecretKey: "test_secret_key",
				UseSSL:    false,
				Region:    "us-east-1",
				Bucket:    "test-bucket",
			},
			JWT: JWT{
				Issuer:        "auth-service-test",
				Leeway:        30 * time.Second,
				CacheTTL:      30 * time.Second,
				KeyBits:       2048, // smaller for test speed
				KeysDir:       os.TempDir(),
				RefreshPepper: "dGVzdC1yZWZyZXNoLXBlcHBlci1taW4tMzItYnl0ZXMtbG9uZw",
				APIKeyPepper:  "dGVzdC1yZWZyZXNoLXBlcHBlci1taW4tMzItYnl0ZXMtbG9uZw",
			},
			APIKeys: APIKeys{
				SignExpireTime: 10,
			},
			Metrics: Metrics{Enabled: true},
			Swagger: Swagger{Enabled: true},
			Proto:   Proto{Enabled: true},
			Admin:   Admin{Enabled: true},
			Cookie: Cookie{
				Domain:   "localhost",
				Path:     "/",
				HttpOnly: true,
				MaxAge:   3600,
				Secure:   false,
			},
			Telegram: Telegram{
				BotToken: "test_token",
				ChatID:   "test_chat_id",
			},
			Firebase: Firebase{
				Mobile: FirebaseConf{Type: "service_account"},
				Web:    FirebaseConf{Type: "service_account"},
			},
		}

		// Clean up string fields from quotes
		cleanConfigStrings(reflect.ValueOf(cfg).Elem())

		testInstance = cfg
	})

	return testInstance, nil
}

// ResetTestConfig resets the test config singleton (useful for tests)
func ResetTestConfig() {
	testOnce = sync.Once{}
	testInstance = nil
}
