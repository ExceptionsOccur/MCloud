package routes

import (
	"mcloud/controllers"

	"github.com/gin-gonic/gin"
)

// registerPersons 人员域
func registerPersons(g *gin.RouterGroup) {
	person := controllers.NewPersonController()
	personGroup := g.Group("/persons")
	personGroup.GET("", person.List)
	personGroup.POST("", person.Create)
	personGroup.PUT("/:id", person.Update)
	personGroup.DELETE("/:id", person.Delete)
}
