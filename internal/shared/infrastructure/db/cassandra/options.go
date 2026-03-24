package cassandra

import (
	"time"

	"github.com/gocql/gocql"
)

// Option defines a function type for configuring Cassandra cluster.
type Option func(*gocql.ClusterConfig)

// WithNumConns sets the number of connections per host.
func WithNumConns(n int) Option {
	return func(cluster *gocql.ClusterConfig) {
		cluster.NumConns = n
	}
}

// WithTimeout sets the connection timeout.
func WithTimeout(d time.Duration) Option {
	return func(cluster *gocql.ClusterConfig) {
		cluster.Timeout = d
	}
}

// WithConsistency sets the default consistency level.
func WithConsistency(c gocql.Consistency) Option {
	return func(cluster *gocql.ClusterConfig) {
		cluster.Consistency = c
	}
}

// WithConnectTimeout sets the initial connection timeout.
func WithConnectTimeout(d time.Duration) Option {
	return func(cluster *gocql.ClusterConfig) {
		cluster.ConnectTimeout = d
	}
}
