package routes

import (
	"mcloud/controllers"

	"github.com/gin-gonic/gin"
)

// registerDataExchange 数据备份域：8 表 xlsx 整体导出/导入
func registerDataExchange(g *gin.RouterGroup) {
	dataExchange := controllers.NewDataExchangeController()
	g.GET("/export/all", dataExchange.Export)
	g.POST("/import/all", dataExchange.Import)
}
