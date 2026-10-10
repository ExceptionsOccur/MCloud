package routes

import (
	"mcloud/controllers"

	"github.com/gin-gonic/gin"
)

// registerCloudResources 云资源域：按区域 upsert
func registerCloudResources(g *gin.RouterGroup) {
	cloudResource := controllers.NewCloudResourceController()
	crGroup := g.Group("/cloud-resources")
	crGroup.GET("", cloudResource.List)
	crGroup.PUT("", cloudResource.Update)
}
