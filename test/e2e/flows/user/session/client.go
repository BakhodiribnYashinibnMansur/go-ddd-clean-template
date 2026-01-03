package session

import (
	"net/http"
	"testing"
)

type Client struct {
	endpoint string
	client   *http.Client
}

func New(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
		client:   &http.Client{},
	}
}

// List retrieves all sessions for the authenticated user
func (c *Client) List(t *testing.T, token string) *http.Response {
	req, err := http.NewRequest(http.MethodGet, c.endpoint+"/api/v1/sessions/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

// Delete revokes/deletes a specific session
func (c *Client) Delete(t *testing.T, token, sessionID string) *http.Response {
	req, err := http.NewRequest(http.MethodDelete, c.endpoint+"/api/v1/sessions/"+sessionID, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

// RevokeAll revokes all sessions for the authenticated user
func (c *Client) RevokeAll(t *testing.T, token string) *http.Response {
	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/api/v1/sessions/revoke-all", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}
