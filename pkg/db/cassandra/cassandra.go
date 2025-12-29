// Package cassandra implements Cassandra connection.
package cassandra

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"gct/config"
	"gct/pkg/logger"
)

const (
	// ENV PROD
	numConnsProd   = 4
	timeoutProd    = 10 * time.Second
	connectTimeout = 10 * time.Second

	// ENV DEV
	numConnsDev = 2
	timeoutDev  = 5 * time.Second

	defaultConsistency = gocql.Quorum
)

// Cassandra struct wraps gocql.Session.
type Cassandra struct {
	Session *gocql.Session
}

// New creates a new Cassandra session with optimized settings.
func New(ctx context.Context, env string, cfg config.Cassandra, l logger.Log, opts ...Option) (*Cassandra, error) {
	cluster := gocql.NewCluster(cfg.Host)
	cluster.Keyspace = cfg.Name
	cluster.Port = cfg.Port

	if cfg.User != "" && cfg.Password != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: cfg.User,
			Password: cfg.Password,
		}
	}

	// Apply pool configuration
	applyClusterConfig(env, cluster)

	// Apply custom options
	for _, opt := range opts {
		opt(cluster)
	}

	session, err := cluster.CreateSession()
	if err != nil {
		l.Errorw("failed to create Cassandra session", zap.Error(err))
		return nil, fmt.Errorf("create cassandra session: %w", err)
	}

	c := &Cassandra{
		Session: session,
	}

	l.Infow("Cassandra connected successfully")

	return c, nil
}

// applyClusterConfig applies cluster configuration with defaults.
func applyClusterConfig(env string, cluster *gocql.ClusterConfig) {
	if env == "production" || env == "PROD" {
		cluster.NumConns = numConnsProd
		cluster.Timeout = timeoutProd
	} else {
		cluster.NumConns = numConnsDev
		cluster.Timeout = timeoutDev
	}

	cluster.Consistency = defaultConsistency
	cluster.ConnectTimeout = connectTimeout
}

// Close gracefully closes the Cassandra session.
func (c *Cassandra) Close() {
	if c != nil && c.Session != nil {
		c.Session.Close()
	}
}
