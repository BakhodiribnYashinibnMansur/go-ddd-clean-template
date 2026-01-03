package client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func stringPtr(s string) *string {
	return &s
}

func createUserAndGetToken(t *testing.T, handler *gin.Engine, phone, password string) string {
	t.Helper()

	signupBody, _ := json.Marshal(map[string]string{
		"username": "test_user_" + phone,
		"phone":    phone,
		"password": password,
	})
	wSignup := httptest.NewRecorder()
	handler.ServeHTTP(wSignup,
		httptest.NewRequest(http.MethodPost, "/api/v1/users/sign-up", bytes.NewBuffer(signupBody)))
	if wSignup.Code != http.StatusCreated && wSignup.Code != http.StatusConflict {
		t.Fatalf("Sign-up failed with status %d: %s", wSignup.Code, wSignup.Body.String())
	}

	signinBody, _ := json.Marshal(map[string]string{
		"phone":    phone,
		"password": password,
	})
	wLogin := httptest.NewRecorder()
	handler.ServeHTTP(wLogin,
		httptest.NewRequest(http.MethodPost, "/api/v1/users/sign-in", bytes.NewBuffer(signinBody)))

	if wLogin.Code != http.StatusOK {
		t.Fatalf("Sign-in failed with status %d: %s", wLogin.Code, wLogin.Body.String())
	}

	var loginResp map[string]any
	json.Unmarshal(wLogin.Body.Bytes(), &loginResp)
	data, ok := loginResp["data"].(map[string]any)
	if !ok {
		t.Fatalf("Sign-in response data is not a map: %v", loginResp["data"])
	}
	return data["access_token"].(string)
}
