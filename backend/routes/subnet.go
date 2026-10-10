package routes

import (
	"mcloud/controllers"

	"github.com/gin-gonic/gin"
)

// registerSubnets IP 网段域
func registerSubnets(g *gin.RouterGroup) {
	subnet := controllers.NewSubnetController()
	subnetGroup := g.Group("/ip-subnets")
	subnetGroup.GET("", subnet.List)
	subnetGroup.POST("", subnet.Create)
	subnetGroup.PUT("/:id", subnet.Update)
	subnetGroup.DELETE("/:id", subnet.Delete)
}
