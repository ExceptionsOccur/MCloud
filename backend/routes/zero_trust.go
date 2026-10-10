package routes

import (
	"mcloud/controllers"

	"github.com/gin-gonic/gin"
)

// registerZeroTrusts 零信任台账域
func registerZeroTrusts(g *gin.RouterGroup) {
	zeroTrust := controllers.NewZeroTrustController()
	zeroTrustGroup := g.Group("/zero-trusts")
	zeroTrustGroup.GET("", zeroTrust.List)
	zeroTrustGroup.POST("", zeroTrust.Create)
	zeroTrustGroup.POST("/batch", zeroTrust.BatchCreateText)
	zeroTrustGroup.PUT("/:id", zeroTrust.Update)
	zeroTrustGroup.DELETE("/:id", zeroTrust.Delete)
}
