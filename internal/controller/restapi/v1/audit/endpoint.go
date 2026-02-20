package audit

import (
	"gct/internal/controller/restapi/v1/audit/history"
	"gct/internal/controller/restapi/v1/audit/log"
	"gct/internal/controller/restapi/v1/audit/metric"

	"github.com/gin-gonic/gin"
)

// AuditRoute delegates route registration to the corresponding sub-controllers (Log and History).
// It acts as a central hub for organizing all auditing-related API paths.
func AuditRoute(api *gin.RouterGroup, controller *Controller, authMiddleware, authzMiddleware gin.HandlerFunc) {
	log.Route(api, controller.Log, authMiddleware, authzMiddleware)
	history.Route(api, controller.History, authMiddleware, authzMiddleware)
	metric.Route(api, controller.Metric, authMiddleware, authzMiddleware)
}
