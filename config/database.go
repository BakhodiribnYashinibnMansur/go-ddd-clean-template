package config

// Database groups all supported databases -.
type Database struct {
	Postgres      Postgres      `envPrefix:"PG_"`
	MySQL         MySQL         `envPrefix:"MYSQL_"`
	MongoDB       MongoDB       `envPrefix:"MONGO_"`
	Redis         Redis         `envPrefix:"REDIS_"`
	Cassandra     Cassandra     `envPrefix:"CASSANDRA_"`
	Elasticsearch Elasticsearch `envPrefix:"ELASTIC_"`
	ClickHouse    ClickHouse    `envPrefix:"CH_"`
	SqlLite       SqlLite       `envPrefix:"SQLITE_"`
}

// BaseDB contains common database connection fields -.
type BaseDB struct {
	Host     string `env:"HOST,required"`
	Port     int    `env:"PORT,required"`
	Name     string `env:"NAME,required"`
	User     string `env:"USER,required"`
	Password string `env:"PASSWORD,required"`
	SSLMode  string `env:"SSL_MODE" envDefault:"disable"`
	PoolMax  int    `env:"POOL_MAX" envDefault:"10"`
}

// Database specific configs -.
type (
	Postgres      struct{ BaseDB }
	MySQL         struct{ BaseDB }
	MongoDB       struct{ BaseDB }
	Redis         struct{ BaseDB }
	Cassandra     struct{ BaseDB }
	Elasticsearch struct{ BaseDB }
	ClickHouse    struct{ BaseDB }

	SqlLite struct {
		File string `env:"FILE,required"`
	}
)
