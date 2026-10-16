package routes

import (
	"mcloud/controllers"

	"github.com/gin-gonic/gin"
)

// registerAI AI 接入域（配置读写）
func registerAI(g *gin.RouterGroup) {
	ai := controllers.NewAIController()
	aiGroup := g.Group("/ai")
	aiGroup.GET("/settings", ai.GetSettings)
	aiGroup.PUT("/settings", ai.SaveSettings)
}
