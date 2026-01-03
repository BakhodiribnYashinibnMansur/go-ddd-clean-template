package minio

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"gct/internal/controller/restapi"
	"gct/internal/repo"
	"gct/internal/usecase"
	"gct/pkg/logger"
	"gct/test/e2e/common/setup"
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

	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg)

	handler := gin.New()
	restapi.NewRouter(handler, setup.TestCfg, useCases, l)

	return httptest.NewServer(handler)
}
