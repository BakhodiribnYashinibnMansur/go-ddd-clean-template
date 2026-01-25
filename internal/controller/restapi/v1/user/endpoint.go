package user

import (
	"gct/internal/controller/restapi/v1/user/client"
	"gct/internal/controller/restapi/v1/user/session"

	"github.com/gin-gonic/gin"
)

// UserRoute defines the routing structure for user-centric domains.
// It delegates the actual endpoint registration to the client and session sub-packages,
// passing required security middlewares for authentication and CSRF protection.
func UserRoute(api *gin.RouterGroup, controller *Controller, authMiddleware, csrfMiddleware gin.HandlerFunc) {
	group := api.Group("/")

	// Register account/profile related routes.
	client.Route(group, controller.ClientI, authMiddleware, csrfMiddleware)

	// Register session/login related routes.
	session.Route(group, controller.SessionI, authMiddleware, csrfMiddleware)
}
