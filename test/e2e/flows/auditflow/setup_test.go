package auditflow

import (
	"context"
	"net/http/httptest"
	"testing"

	"gct/internal/app"
	authzmw "gct/internal/context/iam/authz/interfaces/http/middleware"
	auditmw "gct/internal/context/iam/audit/interfaces/http/middleware"
	"gct/internal/context/iam/user/application/command"
	usermw "gct/internal/context/iam/user/interfaces/http/middleware"
	userport "gct/internal/context/iam/user/interfaces/port"
	"gct/internal/kernel/application"
	"gct/internal/kernel/consts"
	"gct/internal/kernel/infrastructure/eventbus"
	"gct/internal/kernel/infrastructure/logger"
	sharedmw "gct/internal/kernel/infrastructure/middleware"
	jwtpkg "gct/internal/kernel/infrastructure/security/jwt"
	"gct/test/e2e/common/setup"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	setup.SetupTestEnvironment(m)
}

func cleanDB(t *testing.T) {
	t.Helper()
	setup.CleanDB(t)
}

// testServer bundles the httptest server with the event bus used during
// construction so tests can attach cross-BC subscribers for event-flow
// assertions.
type testServer struct {
	*httptest.Server
	EventBus application.EventBus
}

// startAuditTestServer wires the full DDD stack with the Audit BC middleware
// attached (endpoint history + change audit). This mirrors the production
// router in app.go but keeps the surface minimal for E2E assertions on the
// Audit <-> User BC boundary.
func startAuditTestServer(t *testing.T) *testServer {
	t.Helper()

	l := logger.New("debug")
	eventBus := eventbus.NewInMemoryEventBus()

	jwtPrivateKey, err := jwtpkg.ParseRSAPrivateKey(setup.TestCfg.JWT.PrivateKey)
	if err != nil {
		t.Fatalf("failed to parse RSA private key: %s", err)
	}

	bcs, err := app.NewDDDBoundedContexts(
		context.Background(), setup.TestPG.Pool, eventBus, l, nil, command.JWTConfig{
			PrivateKey: jwtPrivateKey,
			Issuer:     setup.TestCfg.JWT.Issuer,
			AccessTTL:  setup.TestCfg.JWT.AccessTTL,
			RefreshTTL: setup.TestCfg.JWT.RefreshTTL,
		},
	)
	if err != nil {
		t.Fatalf("failed to initialize DDD bounded contexts: %s", err)
	}

	handler := gin.New()

	// Enable the Audit BC middleware for these tests. The shared test config
	// leaves the middleware flags disabled by default, so we flip them here.
	setup.TestCfg.Middleware.AuditHistory = true
	setup.TestCfg.Middleware.AuditChange = true

	auditMW := auditmw.NewAuditMiddleware(bcs.Audit.CreateEndpointHistory, bcs.Audit.CreateAuditLog, l)
	bcMW := &sharedmw.BCMiddleware{
		AuditHistory: auditMW.EndpointHistory(),
		AuditChange:  auditMW.ChangeAudit(),
	}

	sharedmw.Setup(handler, setup.TestCfg, setup.TestRedis, bcMW, nil, nil, l)

	authMW := usermw.NewAuthMiddleware(bcs.User.FindSession, bcs.User.FindUserForAuth, setup.TestCfg, l)
	authzMiddleware := authzmw.NewAuthzMiddleware(bcs.Authz.CheckAccess, userport.NewAuthLookupAdapter(bcs.User.FindUserForAuth), l)
	csrfMW := sharedmw.HybridMiddleware(l, consts.CookieCsrfToken)

	app.RegisterDDDRoutes(handler, bcs, authMW.AuthClientAccess, authzMiddleware.Authz, csrfMW, l)

	return &testServer{
		Server:   httptest.NewServer(handler),
		EventBus: eventBus,
	}
}
