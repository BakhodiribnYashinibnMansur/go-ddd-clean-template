package test

import (
	"gct/internal/controller/restapi/v1/test/state"

	"github.com/gin-gonic/gin"
)

// TestRoute defines the routing structure for test-centric endpoints.
func TestRoute(api *gin.RouterGroup, controller *Controller) {
	state.Route(api, controller.StateI)
}
