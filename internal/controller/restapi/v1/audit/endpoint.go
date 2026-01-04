package audit

import (
	"gct/internal/controller/restapi/v1/audit/history"
	"gct/internal/controller/restapi/v1/audit/log"
	"github.com/gin-gonic/gin"
)

func AuditRoute(api *gin.RouterGroup, controller *Controller) {
	log.Route(api, controller.Log)
	history.Route(api, controller.History)
}
