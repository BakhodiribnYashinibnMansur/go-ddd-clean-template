package session

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	userClient "gct/test/e2e/flows/user/client"
	"github.com/stretchr/testify/require"
)

// TestListSessions tests listing sessions with table-driven tests
func TestListSessions(t *testing.T) {
	tests := []struct {
		name           string
		numSessions    int
		useValidToken  bool
		expectedStatus int
		verifyCount    bool
	}{
		{
			name:           "valid list 2 sessions",
			numSessions:    2,
			useValidToken:  true,
			expectedStatus: http.StatusOK,
			verifyCount:    true,
		},
		{
			name:           "unauthorized - no token",
			numSessions:    1,
			useValidToken:  false,
			expectedStatus: http.StatusUnauthorized,
			verifyCount:    false,
		},
	}

	server := startTestServer()
	defer server.Close()

	uClient := userClient.New(server.URL)
	sClient := New(server.URL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanDB(t) // Clean DB for each test case

			// Setup user and sessions
			// Use unique phone/username to be safe
			timestamp := time.Now().UnixNano()
			username := fmt.Sprintf("list_%d", timestamp)
			phone := fmt.Sprintf("90%d", timestamp%1000000000) // Ensure < 13 chars mostly, logic simplified
			if len(phone) > 13 {
				phone = phone[:13]
			}

			resp := uClient.SignUp(t, username, phone, "P@ssw0rd!")
			require.Equal(t, http.StatusCreated, resp.StatusCode, "SignUp failed")
			resp.Body.Close()

			var token string
			// Create sessions
			for i := range tt.numSessions {
				resp := uClient.SignIn(t, phone, "P@ssw0rd!")
				require.Equal(t, http.StatusOK, resp.StatusCode, "SignIn failed")

				if i == 0 { // Use first session token
					var result struct {
						Data struct {
							AccessToken string `json:"access_token"`
						} `json:"data"`
					}
					json.NewDecoder(resp.Body).Decode(&result)
					token = result.Data.AccessToken
				}
				resp.Body.Close()
			}

			if !tt.useValidToken {
				token = ""
			}

			// Test List
			resp = sClient.List(t, token)
			defer resp.Body.Close()

			require.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.verifyCount && resp.StatusCode == http.StatusOK {
				var result struct {
					Data []struct {
						ID      string `json:"id"`
						Revoked bool   `json:"revoked"`
					} `json:"data"`
				}
				bodyBytes, _ := io.ReadAll(resp.Body)
				json.Unmarshal(bodyBytes, &result)

				require.GreaterOrEqual(t, len(result.Data), tt.numSessions)

				// DB Validation
				for _, sess := range result.Data {
					dbSession := getSessionFromDB(t, sess.ID)
					require.NotNil(t, dbSession)
					require.Equal(t, sess.Revoked, dbSession.Revoked)
				}
			}
		})
	}
}

// TestDeleteSession tests deleting/revoking a session
func TestDeleteSession(t *testing.T) {
	tests := []struct {
		name           string
		useValidToken  bool
		expectedStatus int
		verifyRevoked  bool
	}{
		{
			name:           "valid delete session",
			useValidToken:  true,
			expectedStatus: http.StatusOK,
			verifyRevoked:  true,
		},
		{
			name:           "unauthorized - no token",
			useValidToken:  false,
			expectedStatus: http.StatusUnauthorized,
			verifyRevoked:  false,
		},
	}

	server := startTestServer()
	defer server.Close()

	uClient := userClient.New(server.URL)
	sClient := New(server.URL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanDB(t)

			timestamp := time.Now().UnixNano()
			phone := fmt.Sprintf("91%d", timestamp%1000000000)
			if len(phone) > 13 {
				phone = phone[:13]
			}

			resp := uClient.SignUp(t, "user_"+phone, phone, "P@ssw0rd!")
			require.Equal(t, http.StatusCreated, resp.StatusCode, "SignUp failed")
			resp.Body.Close()

			// Session 1 (to have a token)
			resp1 := uClient.SignIn(t, phone, "P@ssw0rd!")
			require.Equal(t, http.StatusOK, resp1.StatusCode, "SignIn 1 failed")
			var result1 struct {
				Data struct {
					AccessToken string `json:"access_token"`
				} `json:"data"`
			}
			json.NewDecoder(resp1.Body).Decode(&result1)
			resp1.Body.Close()
			token := result1.Data.AccessToken

			// Session 2 (to delete)
			resp2 := uClient.SignIn(t, phone, "P@ssw0rd!")
			require.Equal(t, http.StatusOK, resp2.StatusCode, "SignIn 2 failed")
			var result2 struct {
				Data struct {
					SessionID string `json:"session_id"`
				} `json:"data"`
			}
			json.NewDecoder(resp2.Body).Decode(&result2)
			resp2.Body.Close()
			targetSessionID := result2.Data.SessionID

			if !tt.useValidToken {
				token = ""
			}

			// Test Delete
			resp = sClient.Delete(t, token, targetSessionID)
			defer resp.Body.Close()

			require.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.verifyRevoked && resp.StatusCode == http.StatusOK {
				dbSession := getSessionFromDB(t, targetSessionID)
				if dbSession != nil {
					require.True(t, dbSession.Revoked, "session should be revoked")
				}
			}
		})
	}
}

// TestRevokeAllSessions tests revoking all sessions
func TestRevokeAllSessions(t *testing.T) {
	tests := []struct {
		name           string
		useValidToken  bool
		expectedStatus int
		verifyRevoked  bool
	}{
		{
			name:           "valid revoke all",
			useValidToken:  true,
			expectedStatus: http.StatusOK,
			verifyRevoked:  true,
		},
		{
			name:           "unauthorized - no token",
			useValidToken:  false,
			expectedStatus: http.StatusUnauthorized, // Returns 401 for RevokeAll
			verifyRevoked:  false,
		},
	}

	server := startTestServer()
	defer server.Close()

	uClient := userClient.New(server.URL)
	sClient := New(server.URL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanDB(t)

			timestamp := time.Now().UnixNano()
			phone := fmt.Sprintf("92%d", timestamp%1000000000)
			if len(phone) > 13 {
				phone = phone[:13]
			}

			resp := uClient.SignUp(t, "user_"+phone, phone, "P@ssw0rd!")
			require.Equal(t, http.StatusCreated, resp.StatusCode, "SignUp failed")
			resp.Body.Close()

			// Session 1
			resp1 := uClient.SignIn(t, phone, "P@ssw0rd!")
			require.Equal(t, http.StatusOK, resp1.StatusCode, "SignIn 1 failed")
			var result1 struct {
				Data struct {
					AccessToken string `json:"access_token"`
					SessionID   string `json:"session_id"`
				} `json:"data"`
			}
			json.NewDecoder(resp1.Body).Decode(&result1)
			resp1.Body.Close()
			token := result1.Data.AccessToken
			sessionID := result1.Data.SessionID

			if !tt.useValidToken {
				token = ""
			}

			// Test RevokeAll
			resp = sClient.RevokeAll(t, token)
			defer resp.Body.Close()

			require.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.verifyRevoked && resp.StatusCode == http.StatusOK {
				dbSession := getSessionFromDB(t, sessionID)
				if dbSession != nil {
					require.True(t, dbSession.Revoked, "current session should be revoked")
				}

				// Verify token is invalid
				checkResp := sClient.List(t, token)
				defer checkResp.Body.Close()
				require.Equal(t, http.StatusUnauthorized, checkResp.StatusCode)
			}
		})
	}
}
