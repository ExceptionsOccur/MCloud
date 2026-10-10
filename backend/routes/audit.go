package routes

import (
	"mcloud/controllers"

	"github.com/gin-gonic/gin"
)

// registerAudits 审计日志域：仅查询（只增不删由 service 层保证）
func registerAudits(g *gin.RouterGroup) {
	audit := controllers.NewAuditLogController()
	g.GET("/audit-logs", audit.List)
}
