package authz

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

// Client is an HTTP client for the authz E2E tests.
type Client struct {
	endpoint string
	client   *http.Client
}

// New creates a new authz test client.
func New(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
		client:   &http.Client{},
	}
}

// ---------------------------------------------------------------------------
// Auth helpers (needed to obtain a token for protected routes)
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Roles
// ---------------------------------------------------------------------------

// CreateRole creates a new role via POST /api/v1/roles.
func (c *Client) CreateRole(t *testing.T, token, name string, description *string) *http.Response {
	t.Helper()
	payload := map[string]any{"name": name}
	if description != nil {
		payload["description"] = *description
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/api/v1/roles", bytes.NewBuffer(body))
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

// ListRoles retrieves a paginated list of roles via GET /api/v1/roles.
func (c *Client) ListRoles(t *testing.T, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, c.endpoint+"/api/v1/roles", nil)
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

// GetRole retrieves a single role via GET /api/v1/roles/:id.
func (c *Client) GetRole(t *testing.T, token, roleID string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, c.endpoint+"/api/v1/roles/"+roleID, nil)
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

// UpdateRole updates a role via PATCH /api/v1/roles/:id.
func (c *Client) UpdateRole(t *testing.T, token, roleID string, name *string, description *string) *http.Response {
	t.Helper()
	payload := map[string]any{}
	if name != nil {
		payload["name"] = *name
	}
	if description != nil {
		payload["description"] = *description
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPatch, c.endpoint+"/api/v1/roles/"+roleID, bytes.NewBuffer(body))
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

// DeleteRole deletes a role via DELETE /api/v1/roles/:id.
func (c *Client) DeleteRole(t *testing.T, token, roleID string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, c.endpoint+"/api/v1/roles/"+roleID, nil)
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

// AssignPermissionToRole assigns a permission to a role via POST /api/v1/roles/:id/permissions.
func (c *Client) AssignPermissionToRole(t *testing.T, token, roleID, permissionID string) *http.Response {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"permission_id": permissionID,
	})
	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/api/v1/roles/"+roleID+"/permissions", bytes.NewBuffer(body))
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

// ---------------------------------------------------------------------------
// Permissions
// ---------------------------------------------------------------------------

// CreatePermission creates a new permission via POST /api/v1/permissions.
func (c *Client) CreatePermission(t *testing.T, token, name string, description *string) *http.Response {
	t.Helper()
	payload := map[string]any{"name": name}
	if description != nil {
		payload["description"] = *description
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/api/v1/permissions", bytes.NewBuffer(body))
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

// ListPermissions retrieves a paginated list of permissions via GET /api/v1/permissions.
func (c *Client) ListPermissions(t *testing.T, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, c.endpoint+"/api/v1/permissions", nil)
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

// DeletePermission deletes a permission via DELETE /api/v1/permissions/:id.
func (c *Client) DeletePermission(t *testing.T, token, permissionID string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, c.endpoint+"/api/v1/permissions/"+permissionID, nil)
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

// AssignScopeToPermission assigns a scope to a permission via POST /api/v1/permissions/:id/scopes.
func (c *Client) AssignScopeToPermission(t *testing.T, token, permissionID, path, method string) *http.Response {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"path":   path,
		"method": method,
	})
	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/api/v1/permissions/"+permissionID+"/scopes", bytes.NewBuffer(body))
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

// ---------------------------------------------------------------------------
// Policies
// ---------------------------------------------------------------------------

// CreatePolicy creates a new policy via POST /api/v1/policies.
func (c *Client) CreatePolicy(t *testing.T, token, permissionID, effect string, priority int) *http.Response {
	t.Helper()
	body, _ := json.Marshal(map[string]any{
		"permission_id": permissionID,
		"effect":        effect,
		"priority":      priority,
	})
	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/api/v1/policies", bytes.NewBuffer(body))
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

// ListPolicies retrieves a paginated list of policies via GET /api/v1/policies.
func (c *Client) ListPolicies(t *testing.T, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, c.endpoint+"/api/v1/policies", nil)
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

// UpdatePolicy updates a policy via PATCH /api/v1/policies/:id.
func (c *Client) UpdatePolicy(t *testing.T, token, policyID string, effect *string, priority *int) *http.Response {
	t.Helper()
	payload := map[string]any{}
	if effect != nil {
		payload["effect"] = *effect
	}
	if priority != nil {
		payload["priority"] = *priority
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPatch, c.endpoint+"/api/v1/policies/"+policyID, bytes.NewBuffer(body))
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

// DeletePolicy deletes a policy via DELETE /api/v1/policies/:id.
func (c *Client) DeletePolicy(t *testing.T, token, policyID string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, c.endpoint+"/api/v1/policies/"+policyID, nil)
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

// TogglePolicy toggles the enabled state of a policy via POST /api/v1/policies/:id/toggle.
func (c *Client) TogglePolicy(t *testing.T, token, policyID string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/api/v1/policies/"+policyID+"/toggle", nil)
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

// ---------------------------------------------------------------------------
// Scopes
// ---------------------------------------------------------------------------

// CreateScope creates a new scope via POST /api/v1/scopes.
func (c *Client) CreateScope(t *testing.T, token, path, method string) *http.Response {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"path":   path,
		"method": method,
	})
	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/api/v1/scopes", bytes.NewBuffer(body))
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

// ListScopes retrieves a paginated list of scopes via GET /api/v1/scopes.
func (c *Client) ListScopes(t *testing.T, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, c.endpoint+"/api/v1/scopes", nil)
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

// DeleteScope deletes a scope via DELETE /api/v1/scopes.
func (c *Client) DeleteScope(t *testing.T, token, path, method string) *http.Response {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"path":   path,
		"method": method,
	})
	req, err := http.NewRequest(http.MethodDelete, c.endpoint+"/api/v1/scopes", bytes.NewBuffer(body))
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
