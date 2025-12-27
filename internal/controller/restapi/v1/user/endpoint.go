package user

import (
	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/user/client"
	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/user/session"
	"github.com/gin-gonic/gin"
)

func UserRoute(api *gin.RouterGroup, controller *UserController, middlewares ...gin.HandlerFunc) {
	group := api.Group("/user")
	if len(middlewares) > 0 {
		group.Use(middlewares...)
	}
	client.Route(group, controller.ClientI)
	session.Route(group, controller.SessionI)
}
