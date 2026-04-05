package telegram

import (
	"time"

	"gct/internal/kernel/infrastructure/circuitbreaker"
	"gct/internal/kernel/infrastructure/httpclient"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	token             string
	chatID            string
	topics            map[MessageType]string
	timeout           time.Duration
	http              *httpclient.Client
	telegramCB        *circuitbreaker.Breaker
	rdb               redis.Cmdable
	log               logger.Log
	apiSink           httpclient.Sink
	apiSlowThreshold  time.Duration
	apiSuccessSampleR float64
}

type Option func(*Client)

func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

func WithRedis(rdb redis.Cmdable) Option {
	return func(c *Client) {
		c.rdb = rdb
	}
}

func WithLogger(l logger.Log) Option {
	return func(c *Client) {
		c.log = l
	}
}

// WithAPILogSink configures where failed Telegram API calls are logged for
// later inspection in external_api_logs.
func WithAPILogSink(sink httpclient.Sink) Option {
	return func(c *Client) {
		c.apiSink = sink
	}
}

// WithAPILogThresholds configures when successful outgoing calls are persisted:
// those slower than slowThreshold are always logged; the remaining fast ones
// are captured at the given sample rate (0 = never).
func WithAPILogThresholds(slowThreshold time.Duration, successSampleRate float64) Option {
	return func(c *Client) {
		c.apiSlowThreshold = slowThreshold
		c.apiSuccessSampleR = successSampleRate
	}
}

func New(token, chatID string, topics map[MessageType]string, opts ...Option) *Client {
	c := &Client{
		token:   token,
		chatID:  chatID,
		topics:  topics,
		timeout: DefaultTimeout,
	}

	for _, opt := range opts {
		opt(c)
	}

	c.http = httpclient.New(httpclient.Options{
		APIName:           "telegram",
		Timeout:           c.timeout,
		SlowThreshold:     c.apiSlowThreshold,
		SuccessSampleRate: c.apiSuccessSampleR,
	}, c.apiSink, c.log)

	c.telegramCB = circuitbreaker.New("telegram", circuitbreaker.Config{
		FailureThreshold: 3,
		Timeout:          120 * time.Second,
	})

	return c
}
