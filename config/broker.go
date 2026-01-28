package config

// Broker holds configuration for message broker integrations.
type Broker struct {
	Kafka    BrokerEnabled `yaml:"kafka"`    // Kafka message broker.
	NATS     BrokerEnabled `yaml:"nats"`     // NATS message broker.
	RabbitMQ BrokerEnabled `yaml:"rabbitmq"` // RabbitMQ message broker.
}

// BrokerEnabled holds simple enabled flag for brokers.
type BrokerEnabled struct {
	Enabled bool `yaml:"enabled" env:"ENABLED" envDefault:"false"` // Enable/disable broker.
}
