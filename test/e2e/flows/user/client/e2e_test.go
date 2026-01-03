package client

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSignUp tests user signup with table-driven tests using client methods
func TestSignUp(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		phone          string
		password       string
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "valid signup",
			username:       "valid_user",
			phone:          "998901234500",
			password:       "password123",
			expectedStatus: http.StatusCreated,
			checkResponse:  true,
		},
		{
			name:           "duplicate phone",
			username:       "duplicate_user",
			phone:          "998901234500",
			password:       "password123",
			expectedStatus: http.StatusConflict,
			checkResponse:  false,
		},
	}

	cleanDB(t)
	server := startTestServer()
	defer server.Close()

	client := New(server.URL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ✅ Use client method
			resp := client.SignUp(t, tt.username, tt.phone, tt.password)
			defer resp.Body.Close()

			require.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.checkResponse {
				bodyBytes, _ := io.ReadAll(resp.Body)
				var result struct {
					Status string `json:"status"`
					Data   string `json:"data"`
				}
				json.Unmarshal(bodyBytes, &result)
				require.Equal(t, "SUCCESS", result.Status)
				require.NotEmpty(t, result.Data)
			}
		})
	}
}

// TestSignIn tests user signin with table-driven tests using client methods
func TestSignIn(t *testing.T) {
	tests := []struct {
		name           string
		phone          string
		password       string
		expectedStatus int
		checkDB        bool
	}{
		{
			name:           "valid signin",
			phone:          "998901234510",
			password:       "password123",
			expectedStatus: http.StatusOK,
			checkDB:        true,
		},
		{
			name:           "invalid password",
			phone:          "998901234510",
			password:       "wrongpassword",
			expectedStatus: http.StatusUnauthorized,
			checkDB:        false,
		},
		{
			name:           "user not found",
			phone:          "998909999999",
			password:       "password123",
			expectedStatus: http.StatusUnauthorized,
			checkDB:        false,
		},
	}

	cleanDB(t)
	server := startTestServer()
	defer server.Close()

	client := New(server.URL)

	// Setup user
	client.SignUp(t, "signin_test_user", "998901234510", "password123")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ✅ Use client method
			resp := client.SignIn(t, tt.phone, tt.password)
			defer resp.Body.Close()

			require.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.checkDB && resp.StatusCode == http.StatusOK {
				var result struct {
					Data struct {
						UserID    string `json:"user_id"`
						SessionID string `json:"session_id"`
					} `json:"data"`
				}
				json.NewDecoder(resp.Body).Decode(&result)

				// ✅ DB validation
				user := getUserFromDB(t, result.Data.UserID)
				require.NotNil(t, user)

				session := getSessionFromDB(t, result.Data.SessionID)
				require.NotNil(t, session)
				require.False(t, session.Revoked)
			}
		})
	}
}

// TestGetUser tests getting user details with table-driven tests using client methods
func TestGetUser(t *testing.T) {
	tests := []struct {
		name           string
		useValidToken  bool
		useValidUserID bool
		expectedStatus int
	}{
		{
			name:           "valid get user",
			useValidToken:  true,
			useValidUserID: true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized - no token",
			useValidToken:  false,
			useValidUserID: true,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "user not found",
			useValidToken:  true,
			useValidUserID: false,
			expectedStatus: http.StatusNotFound,
		},
	}

	cleanDB(t)
	server := startTestServer()
	defer server.Close()

	client := New(server.URL)

	// Setup user and get token
	client.SignUp(t, "get_test_user", "998901234520", "password123")
	signInResp := client.SignIn(t, "998901234520", "password123")
	var signInResult struct {
		Data struct {
			AccessToken string `json:"access_token"`
			UserID      string `json:"user_id"`
		} `json:"data"`
	}
	json.NewDecoder(signInResp.Body).Decode(&signInResult)
	signInResp.Body.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := signInResult.Data.UserID
			if !tt.useValidUserID {
				userID = "00000000-0000-0000-0000-000000000001"
			}

			token := signInResult.Data.AccessToken
			if !tt.useValidToken {
				token = ""
			}

			// ✅ Use client method
			resp := client.Get(t, token, userID)
			defer resp.Body.Close()

			require.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

// TestUpdateUser tests updating user with table-driven tests using client methods
func TestUpdateUser(t *testing.T) {
	tests := []struct {
		name           string
		newUsername    string
		useValidToken  bool
		expectedStatus int
		checkDB        bool
	}{
		{
			name:           "valid update",
			newUsername:    "updated_username",
			useValidToken:  true,
			expectedStatus: http.StatusOK,
			checkDB:        true,
		},
		{
			name:           "unauthorized - no token",
			newUsername:    "should_fail",
			useValidToken:  false,
			expectedStatus: http.StatusUnauthorized,
			checkDB:        false,
		},
	}

	cleanDB(t)
	server := startTestServer()
	defer server.Close()

	client := New(server.URL)

	// Setup user
	client.SignUp(t, "update_test_user", "998901234530", "password123")
	signInResp := client.SignIn(t, "998901234530", "password123")
	var signInResult struct {
		Data struct {
			AccessToken string `json:"access_token"`
			UserID      string `json:"user_id"`
		} `json:"data"`
	}
	json.NewDecoder(signInResp.Body).Decode(&signInResult)
	signInResp.Body.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := signInResult.Data.AccessToken
			if !tt.useValidToken {
				token = ""
			}

			// ✅ Use client method
			resp := client.Update(t, token, signInResult.Data.UserID, tt.newUsername)
			defer resp.Body.Close()

			require.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.checkDB && resp.StatusCode == http.StatusOK {
				user := getUserFromDB(t, signInResult.Data.UserID)
				require.NotNil(t, user)
				if user.Username != nil {
					require.Equal(t, tt.newUsername, *user.Username)
				}
			}
		})
	}
}

// TestDeleteUser tests deleting user with table-driven tests using client methods
func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name           string
		useValidToken  bool
		expectedStatus int
		verifyDeleted  bool
	}{
		{
			name:           "valid delete",
			useValidToken:  true,
			expectedStatus: http.StatusOK,
			verifyDeleted:  true,
		},
		{
			name:           "unauthorized - no token",
			useValidToken:  false,
			expectedStatus: http.StatusUnauthorized,
			verifyDeleted:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanDB(t)
			server := startTestServer()
			defer server.Close()

			client := New(server.URL)

			// Setup user
			client.SignUp(t, "delete_test_user", "998901234540", "password123")
			signInResp := client.SignIn(t, "998901234540", "password123")
			var signInResult struct {
				Data struct {
					AccessToken string `json:"access_token"`
					UserID      string `json:"user_id"`
				} `json:"data"`
			}
			json.NewDecoder(signInResp.Body).Decode(&signInResult)
			signInResp.Body.Close()

			token := signInResult.Data.AccessToken
			if !tt.useValidToken {
				token = ""
			}

			// ✅ Use client method
			resp := client.Delete(t, token, signInResult.Data.UserID)
			defer resp.Body.Close()

			require.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.verifyDeleted && resp.StatusCode == http.StatusOK {
				// Verify user cannot access after deletion
				resp := client.Get(t, signInResult.Data.AccessToken, signInResult.Data.UserID)
				defer resp.Body.Close()

				require.True(t, resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusNotFound)
			}
		})
	}
}

// TestSignOut tests user signout with table-driven tests using client methods
func TestSignOut(t *testing.T) {
	tests := []struct {
		name           string
		useValidToken  bool
		expectedStatus int
		verifyRevoked  bool
	}{
		{
			name:           "valid signout",
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

	cleanDB(t)
	server := startTestServer()
	defer server.Close()

	client := New(server.URL)

	// Setup user
	client.SignUp(t, "signout_test_user", "998901234550", "password123")
	signInResp := client.SignIn(t, "998901234550", "password123")
	var signInResult struct {
		Data struct {
			AccessToken string `json:"access_token"`
			UserID      string `json:"user_id"`
		} `json:"data"`
	}
	json.NewDecoder(signInResp.Body).Decode(&signInResult)
	signInResp.Body.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := signInResult.Data.AccessToken
			if !tt.useValidToken {
				token = ""
			}

			// ✅ Use client method
			resp := client.SignOut(t, token)
			defer resp.Body.Close()

			require.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.verifyRevoked && resp.StatusCode == http.StatusOK {
				// Verify token is now invalid
				resp := client.Get(t, signInResult.Data.AccessToken, signInResult.Data.UserID)
				defer resp.Body.Close()

				require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			}
		})
	}
}
