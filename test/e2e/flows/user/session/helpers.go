package session

import (
	"context"
	"net/http/httptest"

	"gct/internal/app"
	authzmw "gct/internal/context/iam/authz/interfaces/http/middleware"
	"gct/internal/platform/domain/consts"
	sharedmw "gct/internal/platform/infrastructure/middleware"
	"gct/internal/platform/infrastructure/eventbus"
	"gct/internal/platform/infrastructure/logger"
	jwtpkg "gct/internal/platform/infrastructure/security/jwt"
	"gct/internal/context/iam/user/application/command"
	usermw "gct/internal/context/iam/user/interfaces/http/middleware"
	userport "gct/internal/context/iam/user/interfaces/port"
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

	bcs, err := app.NewDDDBoundedContexts(
		context.Background(), setup.TestPG.Pool, eventBus, l, nil, command.JWTConfig{
			PrivateKey: jwtPrivateKey,
			Issuer:     setup.TestCfg.JWT.Issuer,
			AccessTTL:  setup.TestCfg.JWT.AccessTTL,
			RefreshTTL: setup.TestCfg.JWT.RefreshTTL,
		},
	)
	if err != nil {
		panic("failed to initialize DDD bounded contexts: " + err.Error())
	}

	handler := gin.New()

	// Global middleware (no BC middleware for tests — keep it simple)
	sharedmw.Setup(handler, setup.TestCfg, setup.TestRedis, nil, nil, nil, l)

	// DDD auth/authz middleware
	authMW := usermw.NewAuthMiddleware(bcs.User.FindSession, bcs.User.FindUserForAuth, setup.TestCfg, l)
	authzMiddleware := authzmw.NewAuthzMiddleware(bcs.Authz.CheckAccess, userport.NewAuthLookupAdapter(bcs.User.FindUserForAuth), l)
	csrfMW := sharedmw.HybridMiddleware(l, consts.CookieCsrfToken)

	app.RegisterDDDRoutes(handler, bcs, authMW.AuthClientAccess, authzMiddleware.Authz, csrfMW, l)

	return httptest.NewServer(handler)
}
