package config

// RedisStore - Redis configuration structure.
type RedisStore struct {
	Host     string `env:"HOST" envDefault:"localhost"`
	Port     string `env:"PORT" envDefault:"6379"`
	Password string `env:"PASSWORD"`
	DB       int    `env:"DB" envDefault:"0"`
}

// Addr returns the Redis server address.
func (r *RedisStore) Addr() string {
	if r.Port == "" {
		r.Port = "6379"
	}
	if r.Host == "" {
		r.Host = "localhost"
	}
	return r.Host + ":" + r.Port
}
