package audit

import (
	"github.com/gin-gonic/gin"

	"gct/internal/controller/restapi/v1/audit/history"
	"gct/internal/controller/restapi/v1/audit/log"
)

func AuditRoute(api *gin.RouterGroup, controller *Controller) {
	log.Route(api, controller.Log)
	history.Route(api, controller.History)
}
