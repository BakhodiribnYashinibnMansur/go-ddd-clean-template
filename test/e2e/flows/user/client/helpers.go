package client

import (
	"net/http/httptest"

	"gct/internal/controller/restapi"
	"gct/internal/repo"
	"gct/internal/usecase"
	"gct/pkg/logger"
	"gct/test/e2e/common/setup"
	"github.com/gin-gonic/gin"
)

// startTestServer creates and starts a test HTTP server with full application stack
func startTestServer() *httptest.Server {
	l := logger.New("debug")

	repositories := repo.New(setup.TestPG, nil, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)

	handler := gin.New()
	restapi.NewRouter(handler, setup.TestCfg, useCases, l)

	return httptest.NewServer(handler)
}
