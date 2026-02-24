package setting

import "github.com/gin-gonic/gin"

func Route(api *gin.RouterGroup, c ControllerI, authMiddleware gin.HandlerFunc) {
	g := api.Group("/users/me/settings", authMiddleware)
	{
		g.GET("", c.Gets)
		g.PUT("/:key", c.Set)
		g.POST("/passcode/verify", c.VerifyPasscode)
		g.DELETE("/passcode", c.RemovePasscode)
	}
}
