package telegram

import (
	"net/http"
	"time"

	"gct/internal/shared/infrastructure/circuitbreaker"
	"gct/internal/shared/infrastructure/logger"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	token      string
	chatID     string
	topics     map[MessageType]string
	client     *http.Client
	telegramCB *circuitbreaker.Breaker
	rdb        redis.Cmdable
	log        logger.Log
}

type Option func(*Client)

func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.client.Timeout = timeout
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

func New(token, chatID string, topics map[MessageType]string, opts ...Option) *Client {
	c := &Client{
		token:  token,
		chatID: chatID,
		topics: topics,
		client: &http.Client{Timeout: DefaultTimeout},
	}

	for _, opt := range opts {
		opt(c)
	}

	c.telegramCB = circuitbreaker.New("telegram", circuitbreaker.Config{
		FailureThreshold: 3,
		Timeout:          120 * time.Second,
	})

	return c
}
