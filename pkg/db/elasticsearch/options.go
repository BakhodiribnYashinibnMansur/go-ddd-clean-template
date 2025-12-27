package elasticsearch

import (
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
)

// Option defines a function type for configuring Elasticsearch client.
type Option func(*elasticsearch.Config)

// WithMaxRetries sets the maximum number of retries.
func WithMaxRetries(n int) Option {
	return func(cfg *elasticsearch.Config) {
		cfg.MaxRetries = n
	}
}

// WithTransport sets a custom HTTP transport.
func WithTransport(t http.RoundTripper) Option {
	return func(cfg *elasticsearch.Config) {
		cfg.Transport = t
	}
}

// WithAddresses sets the Elasticsearch addresses.
func WithAddresses(addresses []string) Option {
	return func(cfg *elasticsearch.Config) {
		cfg.Addresses = addresses
	}
}

// WithAPIKey sets the API key for authentication.
func WithAPIKey(key string) Option {
	return func(cfg *elasticsearch.Config) {
		cfg.APIKey = key
	}
}
