package content

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

// ── Notifications ───────────────────────────────────────────────────────────

// CreateNotification creates a new notification.
func (c *Client) CreateNotification(t *testing.T, token string, userID, title, message, nType string) *http.Response {
	t.Helper()
	body, _ := json.Marshal(map[string]any{
		"user_id": userID,
		"title":   title,
		"message": message,
		"type":    nType,
	})
	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/api/v1/notifications", bytes.NewBuffer(body))
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

// ListNotifications retrieves all notifications.
func (c *Client) ListNotifications(t *testing.T, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, c.endpoint+"/api/v1/notifications", nil)
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

// GetNotification retrieves a single notification by ID.
func (c *Client) GetNotification(t *testing.T, token, id string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, c.endpoint+"/api/v1/notifications/"+id, nil)
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

// DeleteNotification removes a notification by ID.
func (c *Client) DeleteNotification(t *testing.T, token, id string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, c.endpoint+"/api/v1/notifications/"+id, nil)
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

// ── Announcements ───────────────────────────────────────────────────────────

// CreateAnnouncement creates a new announcement.
func (c *Client) CreateAnnouncement(t *testing.T, token string, title, content map[string]string, priority int) *http.Response {
	t.Helper()
	body, _ := json.Marshal(map[string]any{
		"title":    title,
		"content":  content,
		"priority": priority,
	})
	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/api/v1/announcements", bytes.NewBuffer(body))
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

// ListAnnouncements retrieves all announcements.
func (c *Client) ListAnnouncements(t *testing.T, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, c.endpoint+"/api/v1/announcements", nil)
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

// GetAnnouncement retrieves a single announcement by ID.
func (c *Client) GetAnnouncement(t *testing.T, token, id string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, c.endpoint+"/api/v1/announcements/"+id, nil)
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

// UpdateAnnouncement patches an existing announcement.
func (c *Client) UpdateAnnouncement(t *testing.T, token, id string, fields map[string]any) *http.Response {
	t.Helper()
	body, _ := json.Marshal(fields)
	req, err := http.NewRequest(http.MethodPatch, c.endpoint+"/api/v1/announcements/"+id, bytes.NewBuffer(body))
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

// DeleteAnnouncement removes an announcement by ID.
func (c *Client) DeleteAnnouncement(t *testing.T, token, id string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, c.endpoint+"/api/v1/announcements/"+id, nil)
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

// ── Translations ────────────────────────────────────────────────────────────

// CreateTranslation creates a new translation.
func (c *Client) CreateTranslation(t *testing.T, token, key, language, value, group string) *http.Response {
	t.Helper()
	body, _ := json.Marshal(map[string]any{
		"key":      key,
		"language": language,
		"value":    value,
		"group":    group,
	})
	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/api/v1/translations", bytes.NewBuffer(body))
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

// ListTranslations retrieves all translations.
func (c *Client) ListTranslations(t *testing.T, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, c.endpoint+"/api/v1/translations", nil)
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

// GetTranslation retrieves a single translation by ID.
func (c *Client) GetTranslation(t *testing.T, token, id string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, c.endpoint+"/api/v1/translations/"+id, nil)
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

// UpdateTranslation patches an existing translation.
func (c *Client) UpdateTranslation(t *testing.T, token, id string, fields map[string]any) *http.Response {
	t.Helper()
	body, _ := json.Marshal(fields)
	req, err := http.NewRequest(http.MethodPatch, c.endpoint+"/api/v1/translations/"+id, bytes.NewBuffer(body))
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

// DeleteTranslation removes a translation by ID.
func (c *Client) DeleteTranslation(t *testing.T, token, id string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, c.endpoint+"/api/v1/translations/"+id, nil)
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
