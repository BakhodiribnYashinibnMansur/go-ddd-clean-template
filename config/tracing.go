package config

type Tracing struct {
	Enabled      bool    `env:"ENABLED" envDefault:"false"`
	ServiceName  string  `env:"SERVICE_NAME" envDefault:"go-clean-template"`
	Endpoint     string  `env:"ENDPOINT" envDefault:"http://localhost:14268/api/traces"`
	HttpEndpoint string  `env:"HTTP_ENDPOINT" envDefault:"http://localhost:16686"`
	Insecure     bool    `env:"INSECURE" envDefault:"true"`
	SamplerRatio float64 `env:"SAMPLER_RATIO" envDefault:"0.1" validate:"min=0,max=1"`
	Jaeger       Jaeger  `envPrefix:"JAEGER_"`
}

type Jaeger struct {
	URL string `env:"URL"`
}
