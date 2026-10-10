package routes

import (
	"mcloud/controllers"

	"github.com/gin-gonic/gin"
)

// registerPublicIPs 公网 IP 资源池域
func registerPublicIPs(g *gin.RouterGroup) {
	publicIP := controllers.NewPublicIPController()
	publicIPGroup := g.Group("/public-ips")
	publicIPGroup.GET("", publicIP.List)
	publicIPGroup.POST("", publicIP.Create)
	publicIPGroup.PUT("/:id", publicIP.Update)
	publicIPGroup.DELETE("/:id", publicIP.Delete)
}
