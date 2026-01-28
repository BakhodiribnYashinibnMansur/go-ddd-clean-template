package config

import (
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
				MySQL: MySQL{BaseDB: BaseDB{
					Host:     "localhost",
					Port:     3306,
					Name:     "test_db",
					User:     "test_user",
					Password: "test_password",
				}},
				MongoDB:       MongoDB{BaseDB: BaseDB{Host: "localhost", Port: 27017}},
				Redis:         Redis{BaseDB: BaseDB{Host: "localhost", Port: 6379, Name: "1"}},
				Cassandra:     Cassandra{BaseDB: BaseDB{Host: "localhost", Port: 9042}},
				Elasticsearch: Elasticsearch{BaseDB: BaseDB{Host: "localhost", Port: 9200}},
				ClickHouse:    ClickHouse{BaseDB: BaseDB{Host: "localhost", Port: 9000}},
				SqlLite:       SqlLite{File: "./test_data.db"},
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
				AccessTTL:  15 * time.Minute,
				RefreshTTL: 720 * time.Hour,
				Issuer:     "auth-service-test",
				PrivateKey: `-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDvFzRJ1otZ7BPH
eHpFBIdfWxgFR3W5b9YZeqEZWLKAkGgiBZIpo6+wrmL7wUw6YUoAAiRkUhuEhDao
shyzohbphughgDI2jenLETzuuTKuriRSx9fotYdPbO0lK5IDu+JFCnHGa8AYF0wY
49wMUbp8JHnaBDlxKBCxKBuKbgHwoF5IG+TrAt9zGsJRnctGp1kFBbgE4ozcljsV
QwyAd9GqlI/uS1z6WzoEyzBtcEH6+zOJF9OK//kOO7+Oc4qMxd1bzWkPCi/x+Pnd
Bc7XKr1DNtstn0fmrlMvQI3vZarQxYPiMRuRyb42BFGRmrEjVk0wY6rTsw14g1J+
vNmE59Z/AgMBAAECggEAJ1h5OWF+IzkvdBcGgA8ju/SAunWtEOwvnpfIpSQsk+2v
xVGHYSTXx8qa7XU89yqFhATWOlAsyRz85bwR7xnQjXOvBbxUBxhJjipzDZIanhZ4
UcsjY99juhVh3UkTSVwE+3mhiZa30P8cfcgZsUlN2Bokl1U0osOGI7FG/vvlg9R+
4LTADWVUpjbbV42IpGnh3DVocHHKFvT+BwmTSgwt6fMamN8dsuzJKsTcM1JFYN/x
rB0Cu0AsGgzMy17ZCj+i5kG65vRuNsUVTzb+Nneen/+qjwvO3/h/ARAY4ZHmca84
/bekD1KMv76r7FZ6HeOpu+yWAMNxO3seK7MsS8K8aQKBgQD6NGwiay3Enf+cxvC0
y4omddIFWAFt2KQYN352tLGvrsEoBcRCnfCbQacO2KpGJnb+O7bcpIW7gWdUaVz+
Vg1igmFOdpXo57B8Ct6ROH7SujOp9PliPIcKwPu0aErlQ/kAry+96s2yhjbna8+z
N5lBDPHD21RcJR+/LbjwhE19owKBgQD0oOGQphxQUmx0jSXO0jcWc/g44s29RhkW
pCULEGjG/3sXrc9eGr/VJ08qntmrJU70MPMjeLiu6UM/rdHeJoQjEtmfAZN2prNm
SnGM0XLBjRCE4EqblneJhvPioHXlfFVixsBJAJN1zm+7322NtXtOR1PCJXWA88Tz
mMNYtW2ZdQKBgQCnEbQW83xXKq1REWIPR04TSl8X5GDn6V4BMaUHPLbdOZKO1/Lq
DK5p7VfQyQpB11NjhZogENefkdPegJBw4CMF4Ut6aiLFp1eoLFXboF7G9UCkPwj6
+LGvk5c/Kti/6DhvpYr6hLwfdhFZTBsfb4Os9SjGgED/WmatcKlqKN3ZgwKBgQCi
HhxeaDdLY9RMSV5M+jNXxfMyf9wpG1N1FcMW2gEWICnLP3y1uLR45lwouq02Jrt0
SRxY3aBHCn9urBrxRkU7mTpvjfPUJhWuLJej4wpSCtJvvNS017rQgYcPIZgARa2w
kFbOCnuvDugtcZyA1UyqS8rOV1TP6L0VUp/jIhlIIQKBgEEmDV9ztio8fDwrsqev
GGk7K3250Mf35/7AVlfftLbhOzCIp2rgaroDrE7h+07x7L2drurnNMLN8nQoadS4
QnE2vfbCnBdOVabThW6XsSidJl/aUMeeMYUbEBYDK9h3rc6flvupbXnKvw7/x/AE
u5J3AecG71JSRqpMXiGl53kJ
-----END PRIVATE KEY-----`,
				PublicKey: `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA7xc0SdaLWewTx3h6RQSH
X1sYBUd1uW/WGXqhGViygJBoIgWSKaOvsK5i+8FMOmFKAAIkZFIbhIQ2qLIcs6IW
6YboIYAyNo3pyxE87rkyrq4kUsfX6LWHT2ztJSuSA7viRQpxxmvAGBdMGOPcDFG6
fCR52gQ5cSgQsSgbim4B8KBeSBvk6wLfcxrCUZ3LRqdZBQW4BOKM3JY7FUMMgHfR
qpSP7ktc+ls6BMswbXBB+vsziRfTiv/5Dju/jnOKjMXdW81pDwov8fj53QXO1yq9
QzbbLZ9H5q5TL0CN72Wq0MWD4jEbkcm+NgRRkZqxI1ZNMGOq07MNeINSfrzZhOfW
fwIDAQAB
-----END PUBLIC KEY-----`,
			},
			APIKeys: APIKeys{
				XApiKey: "test_api_key_12345",
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
			Connectivity: Connectivity{
				GRPC: GRPC{Port: "50051"},
				RMQ: RMQ{
					ServerExchange: "server",
					ClientExchange: "client",
					URL:            "amqp://guest:guest@localhost:5672/",
				},
				NATS: NATS{
					ServerExchange: "server",
					URL:            "nats://localhost:4222",
				},
				Kafka: Kafka{
					Brokers: []string{"localhost:9092"},
					Topic:   "topic",
					GroupId: "group",
				},
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
