package rest

import (
	"context"
	"time"

	"github.com/go-resty/resty/v2"
)

// Client represents a REST client wrapper
type Client struct {
	resty *resty.Client
}

// New creates a new Client instance with default settings
func New(timeout time.Duration) *Client {
	c := resty.New()
	c.SetTimeout(timeout)
	c.SetRetryCount(3)
	c.SetRetryWaitTime(100 * time.Millisecond)
	c.SetRetryMaxWaitTime(2 * time.Second)

	return &Client{resty: c}
}

// SendPostBasicRequest sends a POST request with Basic Auth
func (c *Client) SendPostBasicRequest(ctx context.Context, endpoint string, body interface{}, login, password string) (*resty.Response, error) {
	return c.resty.R().
		SetContext(ctx).
		SetBody(body).
		SetBasicAuth(login, password).
		Post(endpoint)
}

// SendPostBearerRequest sends a POST request with Bearer Auth
func (c *Client) SendPostBearerRequest(ctx context.Context, endpoint string, body interface{}, token string, query map[string]string) (*resty.Response, error) {
	return c.resty.R().
		SetContext(ctx).
		SetBody(body).
		SetAuthToken(token).
		SetQueryParams(query).
		Post(endpoint)
}

// SendGetBearerRequest sends a GET request with Bearer Auth
func (c *Client) SendGetBearerRequest(ctx context.Context, endpoint string, token string, query map[string]string) (*resty.Response, error) {
	return c.resty.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetQueryParams(query).
		Get(endpoint)
}
