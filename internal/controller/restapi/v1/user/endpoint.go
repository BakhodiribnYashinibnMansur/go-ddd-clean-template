package user

import (
	"gct/internal/controller/restapi/v1/user/client"
	"gct/internal/controller/restapi/v1/user/session"
	"github.com/gin-gonic/gin"
)

func UserRoute(api *gin.RouterGroup, controller *Controller, authMiddleware, refreshMiddleware, csrfMiddleware gin.HandlerFunc) {
	group := api.Group("/")

	client.Route(group, controller.ClientI, authMiddleware, refreshMiddleware, csrfMiddleware)
	session.Route(group, controller.SessionI, authMiddleware, csrfMiddleware)
}
