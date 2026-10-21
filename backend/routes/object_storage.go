package routes

import (
	"mcloud/controllers"

	"github.com/gin-gonic/gin"
)

// registerObjectStorages 对象存储台账域（T-088）
func registerObjectStorages(g *gin.RouterGroup) {
	ctrl := controllers.NewObjectStorageController()
	group := g.Group("/object-storages")
	group.GET("", ctrl.List)
	group.POST("", ctrl.Create)
	group.PUT("/:id", ctrl.Update)
	group.DELETE("/:id", ctrl.Delete)
}
