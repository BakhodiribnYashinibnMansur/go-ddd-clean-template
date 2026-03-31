package config

// SSE holds Server-Sent Events configuration.
type SSE struct {
	Enabled           bool  `yaml:"enabled" env:"ENABLED" envDefault:"false"`
	StreamMaxLen      int64 `yaml:"stream_max_len" env:"STREAM_MAX_LEN" envDefault:"1000"`
	HeartbeatInterval int   `yaml:"heartbeat_interval" env:"HEARTBEAT_INTERVAL" envDefault:"30"`
	ClientBufferSize  int   `yaml:"client_buffer_size" env:"CLIENT_BUFFER_SIZE" envDefault:"256"`
}
