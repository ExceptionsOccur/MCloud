package routes

import (
	"mcloud/controllers"

	"github.com/gin-gonic/gin"
)

// registerStats 统计域：IP 使用统计、连通性探测、业务统计
func registerStats(g *gin.RouterGroup) {
	stats := controllers.NewStatsController()
	statsGroup := g.Group("/stats")
	statsGroup.GET("/ip-usage", stats.IPUsage)
	statsGroup.POST("/probe", stats.Probe)
	statsGroup.GET("/business", stats.BusinessStats)
}
