package routes

import (
	"mcloud/controllers"

	"github.com/gin-gonic/gin"
)

// registerHosts 主机域：CRUD、批量、CSV 导入/导出/模板
func registerHosts(g *gin.RouterGroup) {
	host := controllers.NewHostController()
	hosts := g.Group("/hosts")
	hosts.GET("", host.Filter)
	hosts.GET("/regions", host.ListRegions)
	hosts.GET("/:id", host.Get)
	hosts.POST("", host.Create)
	hosts.PUT("/:id", host.Update)
	hosts.DELETE("/:id", host.Delete)

	batch := controllers.NewBatchController()
	batchGroup := g.Group("/batch")
	batchGroup.POST("/hosts", batch.BatchCreate)
	batchGroup.POST("/hosts/text", batch.BatchCreateText)
	batchGroup.PUT("/hosts", batch.BatchUpdate)

	csv := controllers.NewCSVController()
	g.POST("/import", csv.Import)
	g.GET("/export", csv.Export)
	g.GET("/template", csv.Template)
}
