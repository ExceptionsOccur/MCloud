package routes

import (
	"mcloud/controllers"

	"github.com/gin-gonic/gin"
)

// registerAI AI 接入域（配置读写 + 申请表单提取与归档）
func registerAI(g *gin.RouterGroup) {
	ai := controllers.NewAIController()
	aiGroup := g.Group("/ai")
	aiGroup.GET("/settings", ai.GetSettings)
	aiGroup.PUT("/settings", ai.SaveSettings)
	aiGroup.POST("/extract", ai.Extract)
	aiGroup.POST("/archive", ai.Archive)
}
