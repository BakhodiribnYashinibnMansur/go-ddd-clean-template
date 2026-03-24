// Package elasticsearch implements Elasticsearch connection.
package elasticsearch

import (
	"context"
	"errors"
	"fmt"

	"gct/config"
	"gct/internal/shared/infrastructure/logger"
	"github.com/elastic/go-elasticsearch/v8"
	"go.uber.org/zap"
)

const (
	defaultMaxRetries = 3
)

// ErrElasticsearchStatus is returned when Elasticsearch returns an error status.
var ErrElasticsearchStatus = errors.New("elasticsearch returned error status")

// Elasticsearch struct wraps elasticsearch.Client.
type Elasticsearch struct {
	Client *elasticsearch.Client
}

// New creates a new Elasticsearch client with optimized settings.
func New(ctx context.Context, env string, cfg config.Elasticsearch, l logger.Log, opts ...Option) (*Elasticsearch, error) {
	esConfig := elasticsearch.Config{
		Addresses: []string{
			fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port),
		},
		MaxRetries: defaultMaxRetries,
	}

	if cfg.User != "" && cfg.Password != "" {
		esConfig.Username = cfg.User
		esConfig.Password = cfg.Password
	}

	// Apply custom options
	for _, opt := range opts {
		opt(&esConfig)
	}

	client, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		l.Errorw("failed to create Elasticsearch client", zap.Error(err))
		return nil, fmt.Errorf("create elasticsearch client: %w", err)
	}

	// Verify connection
	res, err := client.Info()
	if err != nil {
		l.Errorw("failed to get Elasticsearch info", zap.Error(err))
		return nil, fmt.Errorf("verify elasticsearch connection: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		l.Errorw("Elasticsearch returned error", zap.String("status", res.Status()))
		return nil, fmt.Errorf("%w: %s", ErrElasticsearchStatus, res.Status())
	}

	e := &Elasticsearch{
		Client: client,
	}

	l.Infow("Elasticsearch connected successfully")

	return e, nil
}
