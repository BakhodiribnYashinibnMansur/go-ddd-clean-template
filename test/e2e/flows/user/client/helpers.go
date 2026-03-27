package client

import (
	"net/http/httptest"

	"gct/internal/app"
	authzmw "gct/internal/authz/interfaces/http/middleware"
	"gct/internal/shared/domain/consts"
	sharedmw "gct/internal/shared/infrastructure/middleware"
	"gct/internal/shared/infrastructure/eventbus"
	"gct/internal/shared/infrastructure/logger"
	jwtpkg "gct/internal/shared/infrastructure/security/jwt"
	usermw "gct/internal/user/interfaces/http/middleware"
	"gct/test/e2e/common/setup"

	"github.com/gin-gonic/gin"
)

// startTestServer creates and starts a test HTTP server with full DDD application stack
func startTestServer() *httptest.Server {
	l := logger.New("debug")

	eventBus := eventbus.NewInMemoryEventBus()
	jwtPrivateKey, err := jwtpkg.ParseRSAPrivateKey(setup.TestCfg.JWT.PrivateKey)
	if err != nil {
		panic("failed to parse RSA private key: " + err.Error())
	}

	bcs := app.NewDDDBoundedContexts(
		setup.TestPG.Pool, eventBus, l, jwtPrivateKey,
		setup.TestCfg.JWT.Issuer, setup.TestCfg.JWT.AccessTTL, setup.TestCfg.JWT.RefreshTTL,
	)

	handler := gin.New()

	// Global middleware (no BC middleware for tests — keep it simple)
	sharedmw.Setup(handler, setup.TestCfg, setup.TestRedis, nil, l)

	// DDD auth/authz middleware
	authMW := usermw.NewAuthMiddleware(bcs.User.FindSession, bcs.User.FindUserForAuth, setup.TestCfg, l)
	authzMiddleware := authzmw.NewAuthzMiddleware(bcs.Authz.CheckAccess, bcs.User.FindUserForAuth, l)
	csrfMW := sharedmw.HybridMiddleware(l, consts.CookieCsrfToken)

	app.RegisterDDDRoutes(handler, bcs, authMW.AuthClientAccess, authzMiddleware.Authz, csrfMW, l)

	return httptest.NewServer(handler)
}
