package routes

import (
	"mcloud/controllers"

	"github.com/gin-gonic/gin"
)

// registerPortMappings 端口映射台账域
func registerPortMappings(g *gin.RouterGroup) {
	portMapping := controllers.NewPortMappingController()
	mappingGroup := g.Group("/port-mappings")
	mappingGroup.GET("", portMapping.List)
	mappingGroup.POST("", portMapping.Create)
	mappingGroup.POST("/batch", portMapping.BatchCreateText)
	mappingGroup.PUT("/:id", portMapping.Update)
	mappingGroup.DELETE("/:id", portMapping.Delete)
}
