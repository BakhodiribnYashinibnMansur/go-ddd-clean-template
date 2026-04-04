package telegram

import (
	"time"

	"gct/internal/shared/infrastructure/circuitbreaker"
	"gct/internal/shared/infrastructure/httpclient"
	"gct/internal/shared/infrastructure/logger"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	token      string
	chatID     string
	topics     map[MessageType]string
	timeout    time.Duration
	http       *httpclient.Client
	telegramCB *circuitbreaker.Breaker
	rdb        redis.Cmdable
	log        logger.Log
	apiSink    httpclient.Sink
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
		APIName: "telegram",
		Timeout: c.timeout,
	}, c.apiSink, c.log)

	c.telegramCB = circuitbreaker.New("telegram", circuitbreaker.Config{
		FailureThreshold: 3,
		Timeout:          120 * time.Second,
	})

	return c
}
