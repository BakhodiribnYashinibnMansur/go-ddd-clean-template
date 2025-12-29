package session

import (
	"github.com/gin-gonic/gin"

	"gct/internal/usecase"
	"gct/pkg/logger"
)

type ControllerI interface {
	Create(ctx *gin.Context)
	Session(ctx *gin.Context)
	Sessions(ctx *gin.Context)
	UpdateActivity(ctx *gin.Context)
	Delete(ctx *gin.Context)
	RevokeAll(ctx *gin.Context)
	RevokeByDevice(ctx *gin.Context)
}

type Controller struct {
	s *usecase.UseCase
	l logger.Log
}

func New(s *usecase.UseCase, l logger.Log) ControllerI {
	return &Controller{s: s, l: l}
}
