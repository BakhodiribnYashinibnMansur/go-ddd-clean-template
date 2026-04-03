package client

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
)

// TestAuthFlow_SignUpSignInSignOut exercises the full auth lifecycle:
// sign up -> sign in -> verify profile -> sign out -> verify token is revoked.
func TestAuthFlow_SignUpSignInSignOut(t *testing.T) {
	cleanDB(t)
	server := startTestServer()
	defer server.Close()

	c := New(server.URL)

	phone := "+998901234600"
	password := "P@ssw0rd!"
	username := "auth_flow_user"

	// Step 1: Sign Up
	resp := c.SignUp(t, username, phone, password)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("SignUp: expected status %d, got %d; body: %s", http.StatusCreated, resp.StatusCode, body)
	}

	// Step 2: Sign In
	resp = c.SignIn(t, phone, password)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("SignIn: expected status %d, got %d; body: %s", http.StatusOK, resp.StatusCode, body)
	}

	var signInBody struct {
		Data struct {
			AccessToken string `json:"access_token"`
			UserID      string `json:"user_id"`
			SessionID   string `json:"session_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&signInBody); err != nil {
		t.Fatalf("decode sign-in response: %v", err)
	}
	if signInBody.Data.AccessToken == "" {
		t.Fatal("access_token should not be empty")
	}
	if signInBody.Data.UserID == "" {
		t.Fatal("user_id should not be empty")
	}
	if signInBody.Data.SessionID == "" {
		t.Fatal("session_id should not be empty")
	}

	token := signInBody.Data.AccessToken
	userID := signInBody.Data.UserID
	sessionID := signInBody.Data.SessionID

	// Step 3: Verify profile is accessible
	resp = c.Get(t, token, userID)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Get profile: expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var profileBody map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&profileBody); err != nil {
		t.Fatalf("decode profile response: %v", err)
	}
	data, ok := profileBody["data"].(map[string]any)
	if !ok {
		t.Fatal("profile response missing 'data' object")
	}
	if data["phone"] != phone {
		t.Errorf("profile phone = %v, want %q", data["phone"], phone)
	}

	// Step 4: Sign Out
	resp = c.SignOut(t, token, userID, sessionID)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("SignOut: expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Step 5: Verify token is revoked
	time.Sleep(10 * time.Millisecond) // allow async revocation
	resp = c.Get(t, token, userID)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("after sign-out, Get profile: expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

// TestAuthFlow_InvalidCredentials verifies that sign-in with a wrong password
// returns an unauthorized error.
func TestAuthFlow_InvalidCredentials(t *testing.T) {
	cleanDB(t)
	server := startTestServer()
	defer server.Close()

	c := New(server.URL)

	phone := "+998901234601"
	password := "P@ssw0rd!"

	// Create user first
	resp := c.SignUp(t, "invalid_cred_user", phone, password)
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("setup SignUp: expected %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	// Sign in with wrong password
	resp = c.SignIn(t, phone, "WrongP@ss999")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("SignIn with wrong password: expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}

	// Verify the response is a proper error structure
	bodyBytes, _ := io.ReadAll(resp.Body)
	var errResp struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(bodyBytes, &errResp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if errResp.Status != "error" {
		t.Errorf("error response status = %q, want %q", errResp.Status, "error")
	}
}

// TestAuthFlow_DuplicateSignUp verifies that signing up with the same phone
// number twice returns a conflict error.
func TestAuthFlow_DuplicateSignUp(t *testing.T) {
	cleanDB(t)
	server := startTestServer()
	defer server.Close()

	c := New(server.URL)

	phone := "+998901234602"
	password := "P@ssw0rd!"

	// First sign-up should succeed
	resp := c.SignUp(t, "dup_user_1", phone, password)
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("first SignUp: expected %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	// Second sign-up with same phone should fail
	resp = c.SignUp(t, "dup_user_2", phone, password)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("duplicate SignUp: expected status %d, got %d", http.StatusConflict, resp.StatusCode)
	}

	// Verify error response structure
	bodyBytes, _ := io.ReadAll(resp.Body)
	var errResp struct {
		Status     string `json:"status"`
		StatusCode int    `json:"statusCode"`
	}
	if err := json.Unmarshal(bodyBytes, &errResp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if errResp.Status != "error" {
		t.Errorf("error response status = %q, want %q", errResp.Status, "error")
	}
	if errResp.StatusCode != http.StatusConflict {
		t.Errorf("error response statusCode = %d, want %d", errResp.StatusCode, http.StatusConflict)
	}
}
