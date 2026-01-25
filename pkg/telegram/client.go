package telegram

import (
	"net/http"
	"time"
)

type Client struct {
	token  string
	chatID string
	topics map[MessageType]string
	client *http.Client
}

type Option func(*Client)

func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.client.Timeout = timeout
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

	return c
}
