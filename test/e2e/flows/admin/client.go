package admin

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

// ── Auth helpers ────────────────────────────────────────────────────────────

// SignUp creates a new user account.
func (c *Client) SignUp(t *testing.T, username, phone, password string) *http.Response {
	t.Helper()
	body, _ := json.Marshal(map[string]any{
		"username": username,
		"phone":    phone,
		"password": password,
	})
	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/api/v1/auth/sign-up", bytes.NewBuffer(body))
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

// SignIn authenticates a user and returns tokens.
func (c *Client) SignIn(t *testing.T, phone, password string) *http.Response {
	t.Helper()
	body, _ := json.Marshal(map[string]any{
		"login":       phone,
		"password":    password,
		"device_type": "desktop",
	})
	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/api/v1/auth/sign-in", bytes.NewBuffer(body))
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

// ── Feature Flag endpoints ──────────────────────────────────────────────────

// CreateFeatureFlag creates a new feature flag.
func (c *Client) CreateFeatureFlag(t *testing.T, token string, payload map[string]any) *http.Response {
	t.Helper()
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/api/v1/feature-flags", bytes.NewBuffer(body))
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

// ListFeatureFlags returns a paginated list of feature flags.
func (c *Client) ListFeatureFlags(t *testing.T, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, c.endpoint+"/api/v1/feature-flags?limit=10&offset=0", nil)
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

// GetFeatureFlag returns a single feature flag by ID.
func (c *Client) GetFeatureFlag(t *testing.T, token, id string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, c.endpoint+"/api/v1/feature-flags/"+id, nil)
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

// UpdateFeatureFlag updates a feature flag by ID.
func (c *Client) UpdateFeatureFlag(t *testing.T, token, id string, payload map[string]any) *http.Response {
	t.Helper()
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPatch, c.endpoint+"/api/v1/feature-flags/"+id, bytes.NewBuffer(body))
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

// DeleteFeatureFlag deletes a feature flag by ID.
func (c *Client) DeleteFeatureFlag(t *testing.T, token, id string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, c.endpoint+"/api/v1/feature-flags/"+id, nil)
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

// ── Site Setting endpoints ──────────────────────────────────────────────────

// CreateSiteSetting creates a new site setting.
func (c *Client) CreateSiteSetting(t *testing.T, token string, payload map[string]any) *http.Response {
	t.Helper()
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/api/v1/site-settings", bytes.NewBuffer(body))
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

// ListSiteSettings returns a paginated list of site settings.
func (c *Client) ListSiteSettings(t *testing.T, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, c.endpoint+"/api/v1/site-settings?limit=10&offset=0", nil)
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

// GetSiteSetting returns a single site setting by ID.
func (c *Client) GetSiteSetting(t *testing.T, token, id string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, c.endpoint+"/api/v1/site-settings/"+id, nil)
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

// UpdateSiteSetting updates a site setting by ID.
func (c *Client) UpdateSiteSetting(t *testing.T, token, id string, payload map[string]any) *http.Response {
	t.Helper()
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPatch, c.endpoint+"/api/v1/site-settings/"+id, bytes.NewBuffer(body))
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

// DeleteSiteSetting deletes a site setting by ID.
func (c *Client) DeleteSiteSetting(t *testing.T, token, id string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, c.endpoint+"/api/v1/site-settings/"+id, nil)
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
