package integration

import (
	"gct/config"
	"gct/internal/usecase/integration"
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
)

// ControllerI defines the interface for integration controller.
type ControllerI interface {
	IntegrationHandler
	APIKeyHandler
}

// Handler handles integration requests.
type IntegrationHandler interface {
	CreateIntegration(ctx *gin.Context)
	GetIntegration(ctx *gin.Context)
	ListIntegrations(ctx *gin.Context)
	UpdateIntegration(ctx *gin.Context)
	DeleteIntegration(ctx *gin.Context)
	ToggleIntegration(ctx *gin.Context)
}

// APIKeyHandler handles api key requests.
type APIKeyHandler interface {
	CreateAPIKey(ctx *gin.Context)
	GetAPIKey(ctx *gin.Context)
	ListAPIKeys(ctx *gin.Context)
	RevokeAPIKey(ctx *gin.Context)
	DeleteAPIKey(ctx *gin.Context)
}

type Controller struct {
	useCase integration.UseCaseI
	cfg     *config.Config
	logger  logger.Log
}

func New(useCase integration.UseCaseI, cfg *config.Config, logger logger.Log) ControllerI {
	return &Controller{
		useCase: useCase,
		cfg:     cfg,
		logger:  logger,
	}
}
