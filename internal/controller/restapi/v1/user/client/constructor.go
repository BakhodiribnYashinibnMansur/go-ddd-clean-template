package client

import (
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	u *usecase.UseCase
	l logger.Log
}

type ControllerI interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	SignIn(c *gin.Context)
	SignUp(c *gin.Context)
	SignOut(c *gin.Context)
}

func New(u *usecase.UseCase, l logger.Log) ControllerI {
	return &Controller{u: u, l: l}
}
