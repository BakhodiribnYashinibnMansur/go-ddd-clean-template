package admin

import (
	"context"
	"net/http/httptest"

	"gct/internal/app"
	authzmw "gct/internal/context/iam/generic/authz/interfaces/http/middleware"
	"gct/internal/context/iam/generic/user/application/command"
	usermw "gct/internal/context/iam/generic/user/interfaces/http/middleware"
	userport "gct/internal/context/iam/generic/user/interfaces/port"
	"gct/internal/kernel/consts"
	"gct/internal/kernel/infrastructure/eventbus"
	"gct/internal/kernel/infrastructure/logger"
	sharedmw "gct/internal/kernel/infrastructure/middleware"
	"gct/internal/kernel/infrastructure/security/keyring"
	jwtpkg "gct/internal/kernel/infrastructure/security/jwt"
	"gct/test/e2e/common/setup"

	"github.com/gin-gonic/gin"
)

// startTestServer creates and starts a test HTTP server with full DDD
// application stack, mirroring app.initRouter() at a smaller scale.
func startTestServer() *httptest.Server {
	l := logger.New("debug")

	eventBus := eventbus.NewInMemoryEventBus()

	apiKeyPepper, err := setup.TestCfg.JWT.DecodeAPIKeyPepper()
	if err != nil {
		panic("failed to decode API-key pepper: " + err.Error())
	}
	refreshPepper, err := setup.TestCfg.JWT.DecodeRefreshPepper()
	if err != nil {
		panic("failed to decode refresh pepper: " + err.Error())
	}
	kr, err := keyring.New(setup.TestCfg.JWT.KeysDir, setup.TestCfg.JWT.KeyBits)
	if err != nil {
		panic("failed to init keyring: " + err.Error())
	}

	bcs, err := app.NewDDDBoundedContexts(
		context.Background(), setup.TestPG.Pool, eventBus, l, nil,
		command.JWTConfig{Issuer: setup.TestCfg.JWT.Issuer},
		setup.TestCfg, apiKeyPepper, kr, app.SecurityDeps{},
	)
	if err != nil {
		panic("failed to initialize DDD bounded contexts: " + err.Error())
	}

	refreshHasher, err := jwtpkg.NewRefreshHasher(refreshPepper)
	if err != nil {
		panic("failed to init refresh hasher: " + err.Error())
	}

	handler := gin.New()
	sharedmw.Setup(handler, setup.TestCfg, setup.TestRedis, nil, nil, nil, l)

	authMW := usermw.NewAuthMiddleware(
		bcs.User.FindSession, bcs.User.FindUserForAuth, setup.TestCfg, l,
		app.NewMiddlewareResolver(bcs), refreshHasher,
	)
	authzMiddleware := authzmw.NewAuthzMiddleware(bcs.Authz.CheckAccess, userport.NewAuthLookupAdapter(bcs.User.FindUserForAuth), l)
	csrfMW := sharedmw.HybridMiddleware(l, consts.CookieCsrfToken)

	app.RegisterDDDRoutes(handler, bcs, authMW.AuthClientAccess, authzMiddleware.Authz, csrfMW, l)

	return httptest.NewServer(handler)
}
