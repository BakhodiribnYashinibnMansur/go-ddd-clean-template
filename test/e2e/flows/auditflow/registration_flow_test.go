package auditflow

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"testing"
	"time"

	shareddomain "gct/internal/platform/domain"
	"gct/test/e2e/common/setup"
	userclient "gct/test/e2e/flows/user/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_AuditFlow_RegistrationProducesEndpointHistory exercises a cross-BC
// path: a request handled by the User BC must be observed by the Audit BC
// via the shared HTTP middleware pipeline and persisted to endpoint_history.
//
// BCs exercised: User (sign-up command) + Audit (EndpointHistory middleware
// + CreateEndpointHistory command handler + write repo).
//
// Cross-BC assertion: after a POST /auth/sign-up, a row exists in
// endpoint_history whose method/path/status match the User BC request
// outcome — proving the Audit BC received traffic written by the User BC
// via the generic middleware channel.
func TestE2E_AuditFlow_RegistrationProducesEndpointHistory(t *testing.T) {
	cleanDB(t)
	srv := startAuditTestServer(t)
	defer srv.Close()

	uc := userclient.New(srv.URL)

	resp := uc.SignUp(t, "audit_reg_user", "+998901230001", "P@ssw0rd!")
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// EndpointHistory writes happen in a background goroutine with a
	// background context. Poll briefly until the row materialises.
	row := waitForEndpointHistory(t, "POST", "/api/v1/auth/sign-up", 2*time.Second)
	require.NotNil(t, row, "expected endpoint_history row for sign-up, found none")
	assert.Equal(t, http.StatusCreated, row.StatusCode)
	assert.GreaterOrEqual(t, row.DurationMs, 0)
}

// TestE2E_AuditFlow_MutationProducesAuditLog verifies the ChangeAudit
// middleware (Audit BC) records state-changing requests performed by an
// authenticated User BC session.
//
// BCs exercised: User (sign-up, sign-in, update) + Audit (ChangeAudit
// middleware + CreateAuditLog command + write repo).
//
// Cross-BC assertion: after an authenticated PATCH /users/:id, a row exists
// in audit_log whose user_id and session_id match the session issued by the
// User BC sign-in — proving the Audit BC correlated the request to the
// identity attached to the Gin context by the User BC auth middleware.
func TestE2E_AuditFlow_MutationProducesAuditLog(t *testing.T) {
	cleanDB(t)
	srv := startAuditTestServer(t)
	defer srv.Close()

	uc := userclient.New(srv.URL)

	signUpResp := uc.SignUp(t, "audit_mut_user", "+998901230002", "P@ssw0rd!")
	signUpResp.Body.Close()
	require.Equal(t, http.StatusCreated, signUpResp.StatusCode)

	signInResp := uc.SignIn(t, "+998901230002", "P@ssw0rd!")
	defer signInResp.Body.Close()
	require.Equal(t, http.StatusOK, signInResp.StatusCode)

	var signIn struct {
		Data struct {
			AccessToken string `json:"access_token"`
			UserID      string `json:"user_id"`
			SessionID   string `json:"session_id"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(signInResp.Body).Decode(&signIn))
	require.NotEmpty(t, signIn.Data.AccessToken)

	updateResp := uc.Update(t, signIn.Data.AccessToken, signIn.Data.UserID, "audit_updated_name")
	defer updateResp.Body.Close()
	require.Equal(t, http.StatusOK, updateResp.StatusCode)

	// audit_log writes are async (goroutine with background context).
	row := waitForAuditLog(t, signIn.Data.UserID, 2*time.Second)
	require.NotNil(t, row, "expected audit_log row attributed to the session user, found none")
	assert.True(t, row.Success)
	require.NotNil(t, row.SessionID)
	assert.Equal(t, signIn.Data.SessionID, *row.SessionID)
}

// TestE2E_EventBus_UserCreatedPropagation validates that the in-memory event
// bus wired into the test server actually delivers the User BC's
// user.created domain event to a foreign subscriber — the core mechanism
// that every cross-BC reactor (audit, notifications, projections) depends on.
//
// BCs exercised: User (sign-up publishes user.created) + generic subscriber
// standing in for any consumer BC.
//
// Cross-BC assertion: the event arrives at the subscriber with a non-zero
// aggregate ID matching the user just created in PostgreSQL.
func TestE2E_EventBus_UserCreatedPropagation(t *testing.T) {
	cleanDB(t)
	srv := startAuditTestServer(t)
	defer srv.Close()

	var (
		mu         sync.Mutex
		aggregates []string
		received   = make(chan struct{}, 4)
	)

	err := srv.EventBus.Subscribe("user.created", func(_ context.Context, evt shareddomain.DomainEvent) error {
		mu.Lock()
		aggregates = append(aggregates, evt.AggregateID().String())
		mu.Unlock()
		select {
		case received <- struct{}{}:
		default:
		}
		return nil
	})
	require.NoError(t, err)

	uc := userclient.New(srv.URL)
	resp := uc.SignUp(t, "audit_evt_user", "+998901230003", "P@ssw0rd!")
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	select {
	case <-received:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for user.created event")
	}

	mu.Lock()
	defer mu.Unlock()
	require.NotEmpty(t, aggregates, "expected at least one user.created event")
	assert.NotEmpty(t, aggregates[0], "event aggregate ID must not be empty")

	// Confirm the aggregate ID corresponds to a real row — proving the event
	// describes the same entity the User BC persisted.
	var exists bool
	queryErr := setup.TestPG.Pool.QueryRow(
		t.Context(),
		`SELECT EXISTS (SELECT 1 FROM users WHERE id = $1::uuid AND deleted_at = 0)`,
		aggregates[0],
	).Scan(&exists)
	require.NoError(t, queryErr)
	assert.True(t, exists, "user.created aggregate ID should match a persisted user")
}

// --- DB polling helpers ---------------------------------------------------

type endpointHistoryRow struct {
	Method     string
	Path       string
	StatusCode int
	DurationMs int
}

func waitForEndpointHistory(t *testing.T, method, path string, timeout time.Duration) *endpointHistoryRow {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var row endpointHistoryRow
		err := setup.TestPG.Pool.QueryRow(
			t.Context(),
			`SELECT method, path, status_code, duration_ms
			   FROM endpoint_history
			  WHERE method = $1 AND path = $2
			  ORDER BY created_at DESC
			  LIMIT 1`,
			method, path,
		).Scan(&row.Method, &row.Path, &row.StatusCode, &row.DurationMs)
		if err == nil {
			return &row
		}
		time.Sleep(50 * time.Millisecond)
	}
	return nil
}

type auditLogRow struct {
	UserID    *string
	SessionID *string
	Success   bool
}

func waitForAuditLog(t *testing.T, userID string, timeout time.Duration) *auditLogRow {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var row auditLogRow
		err := setup.TestPG.Pool.QueryRow(
			t.Context(),
			`SELECT user_id::text, session_id::text, success
			   FROM audit_log
			  WHERE user_id = $1::uuid
			  ORDER BY created_at DESC
			  LIMIT 1`,
			userID,
		).Scan(&row.UserID, &row.SessionID, &row.Success)
		if err == nil {
			return &row
		}
		time.Sleep(50 * time.Millisecond)
	}
	return nil
}
