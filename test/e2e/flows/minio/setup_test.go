package minio

// TODO: This e2e test needs rewriting for the DDD architecture.
// The old imports (gct/internal/controller/restapi, gct/internal/repo,
// gct/internal/usecase) have been removed during the DDD migration.
//
// To rewrite startTestServer:
//   - Use gct/internal/app.NewDDDBoundedContexts to create bounded contexts
//   - Use gct/internal/app.RegisterDDDRoutes to wire HTTP routes
//   - See test/e2e/flows/user/client/helpers.go for a working DDD example

import (
	"net/http/httptest"
	"testing"

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

// startTestServer is a stub that panics until rewritten for DDD.
// TODO: Rewrite using DDD bootstrap (see test/e2e/flows/user/client/helpers.go).
func startTestServer() *httptest.Server {
	handler := gin.New()
	// TODO: Wire DDD bounded contexts and routes here.
	// Example from test/e2e/flows/user/client/helpers.go:
	//   eventBus := eventbus.NewInMemoryEventBus()
	//   jwtPrivateKey, _ := jwtpkg.ParseRSAPrivateKey(setup.TestCfg.JWT.PrivateKey)
	//   bcs := app.NewDDDBoundedContexts(setup.TestPG.Pool, eventBus, l, jwtPrivateKey, ...)
	//   app.RegisterDDDRoutes(handler, bcs, authMW, authzMW, csrfMW, l)
	// TODO: Wire DDD bounded contexts, then remove this panic.
	_ = handler
	panic("startTestServer: not yet rewritten for DDD architecture — see TODO in setup_test.go")
}
