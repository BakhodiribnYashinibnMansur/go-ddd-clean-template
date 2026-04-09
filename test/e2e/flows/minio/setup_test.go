package minio

import (
	"context"
	"net/http/httptest"
	"testing"

	"encoding/base64"

	"gct/internal/app"
	authzmw "gct/internal/context/iam/generic/authz/interfaces/http/middleware"
	"gct/internal/context/iam/generic/user/application/command"
	"gct/internal/kernel/consts"
	"gct/internal/kernel/infrastructure/eventbus"
	"gct/internal/kernel/infrastructure/logger"
	sharedmw "gct/internal/kernel/infrastructure/middleware"
	jwtpkg "gct/internal/kernel/infrastructure/security/jwt"
	"gct/internal/kernel/infrastructure/security/keyring"
	usermw "gct/internal/context/iam/generic/user/interfaces/http/middleware"
	userport "gct/internal/context/iam/generic/user/interfaces/port"
	"gct/test/e2e/common/setup"

	"github.com/gin-gonic/gin"
	miniogo "github.com/minio/minio-go/v7"
)

func TestMain(m *testing.M) {
	setup.SetupTestEnvironment(m)
}

func cleanDB(t *testing.T) {
	t.Helper()
	setup.CleanDB(t)
}

func startTestServer() *httptest.Server {
	l := logger.New("debug")

	eventBus := eventbus.NewInMemoryEventBus()

	pepperBytes, err := base64.RawStdEncoding.DecodeString(setup.TestCfg.JWT.RefreshPepper)
	if err != nil {
		panic("failed to decode refresh pepper: " + err.Error())
	}
	hasher, err := jwtpkg.NewRefreshHasher(pepperBytes)
	if err != nil {
		panic("failed to create refresh hasher: " + err.Error())
	}
	apiKeyPepper, err := base64.RawStdEncoding.DecodeString(setup.TestCfg.JWT.APIKeyPepper)
	if err != nil {
		panic("failed to decode api key pepper: " + err.Error())
	}
	kr, err := keyring.New(setup.TestCfg.JWT.KeysDir, setup.TestCfg.JWT.KeyBits)
	if err != nil {
		panic("failed to create keyring: " + err.Error())
	}

	jwtCfg := command.JWTConfig{
		Issuer:        setup.TestCfg.JWT.Issuer,
		RefreshHasher: hasher,
	}

	bcs, err := app.NewDDDBoundedContexts(
		context.Background(), setup.TestPG.Pool, eventBus, l, nil,
		jwtCfg, setup.TestCfg, apiKeyPepper, kr, app.SecurityDeps{},
	)
	if err != nil {
		panic("failed to initialize DDD bounded contexts: " + err.Error())
	}

	handler := gin.New()

	sharedmw.Setup(handler, setup.TestCfg, setup.TestRedis, nil, nil, nil, l)

	authMW := usermw.NewAuthMiddleware(bcs.User.FindSession, bcs.User.FindUserForAuth, setup.TestCfg, l)
	authzMiddleware := authzmw.NewAuthzMiddleware(bcs.Authz.CheckAccess, userport.NewAuthLookupAdapter(bcs.User.FindUserForAuth), l)
	csrfMW := sharedmw.HybridMiddleware(l, consts.CookieCsrfToken)

	bucket := "test-bucket"
	setup.TestMinio.MakeBucket(context.Background(), bucket, miniogo.MakeBucketOptions{})

	app.RegisterDDDRoutes(handler, bcs, authMW.AuthClientAccess, authzMiddleware.Authz, csrfMW, l, app.RouteOptions{
		Minio:       setup.TestMinio,
		MinioBucket: bucket,
	})

	return httptest.NewServer(handler)
}
