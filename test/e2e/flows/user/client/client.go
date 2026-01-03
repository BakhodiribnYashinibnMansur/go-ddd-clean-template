package client

import (
	"bytes"
	"encoding/json"
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

// SignUp creates a new user account
func (c *Client) SignUp(t *testing.T, username, phone, password string) *http.Response {
	body, _ := json.Marshal(map[string]string{
		"username": username,
		"phone":    phone,
		"password": password,
	})
	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/api/v1/users/sign-up", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

// SignIn authenticates a user and returns tokens
func (c *Client) SignIn(t *testing.T, phone, password string) *http.Response {
	body, _ := json.Marshal(map[string]string{
		"phone":    phone,
		"password": password,
	})
	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/api/v1/users/sign-in", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

// SignOut revokes the current session
func (c *Client) SignOut(t *testing.T, token string) *http.Response {
	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/api/v1/users/sign-out", nil)
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

// Get retrieves user details by ID
func (c *Client) Get(t *testing.T, token, userID string) *http.Response {
	req, err := http.NewRequest(http.MethodGet, c.endpoint+"/api/v1/users/"+userID, nil)
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

// Update modifies user information
func (c *Client) Update(t *testing.T, token, userID, username string) *http.Response {
	body, _ := json.Marshal(map[string]string{
		"username": username,
	})
	req, err := http.NewRequest(http.MethodPatch, c.endpoint+"/api/v1/users/"+userID, bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

// Delete removes a user account
func (c *Client) Delete(t *testing.T, token, userID string) *http.Response {
	req, err := http.NewRequest(http.MethodDelete, c.endpoint+"/api/v1/users/"+userID, nil)
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
