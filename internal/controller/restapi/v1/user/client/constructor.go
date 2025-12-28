package client

import (
	"crypto/rsa"

	"github.com/evrone/go-clean-template/config"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/jwt"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	u          *usecase.UseCase
	l          logger.Log
	cfg        *config.Config
	privateKey *rsa.PrivateKey
}

type ControllerI interface {
	Create(c *gin.Context)
	User(c *gin.Context)
	Users(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	SignIn(c *gin.Context)
	SignUp(c *gin.Context)
	SignOut(c *gin.Context)
}

func New(u *usecase.UseCase, cfg *config.Config, l logger.Log) ControllerI {
	pk, err := jwt.ParseRSAPrivateKey(cfg.JWT.PrivateKey)
	if err != nil {
		l.Error("ClientController - New - parsedPrivateKey error", err)
	}
	return &Controller{u: u, l: l, cfg: cfg, privateKey: pk}
}
