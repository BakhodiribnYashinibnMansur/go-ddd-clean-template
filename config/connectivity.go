package config

import "strings"

// Connectivity groups communication protocols and brokers -.
type Connectivity struct {
	GRPC  GRPC  `envPrefix:"GRPC_"`
	RMQ   RMQ   `envPrefix:"RMQ_"`
	NATS  NATS  `envPrefix:"NATS_"`
	Kafka Kafka `envPrefix:"KAFKA_"`
}

type (
	// GRPC -.
	GRPC struct {
		Port string `env:"PORT,required"`
	}

	// RMQ -.
	RMQ struct {
		ServerExchange string `env:"RPC_SERVER,required"`
		ClientExchange string `env:"RPC_CLIENT,required"`
		URL            string `env:"URL,required"`
	}

	// NATS -.
	NATS struct {
		ServerExchange string `env:"RPC_SERVER,required"`
		URL            string `env:"URL,required"`
	}

	// Kafka -.
	Kafka struct {
		Brokers []string `env:"BROKERS,required" envSeparator:","`
		Topic   string   `env:"TOPIC,required"`
		GroupId string   `env:"GROUP_ID,required"`
	}
)

// Addr returns the gRPC server address.
func (g *GRPC) Addr() string {
	if g.Port == "" {
		return ":50051"
	}
	if strings.HasPrefix(g.Port, ":") {
		return g.Port
	}
	return ":" + g.Port
}
