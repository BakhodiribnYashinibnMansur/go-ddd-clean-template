package user

import (
	"github.com/gin-gonic/gin"

	"gct/internal/controller/restapi/v1/user/client"
	"gct/internal/controller/restapi/v1/user/session"
)

func UserRoute(api *gin.RouterGroup, controller *Controller, middlewares ...gin.HandlerFunc) {
	group := api.Group("/user")
	if len(middlewares) > 0 {
		group.Use(middlewares...)
	}
	client.Route(group, controller.ClientI)
	session.Route(group, controller.SessionI)
}
