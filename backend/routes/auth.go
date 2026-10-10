package routes

import (
	"mcloud/controllers"

	"github.com/gin-gonic/gin"
)

// registerAuth 认证域：登录/登出公开；me/改密走受保护组
func registerAuth(api, protected *gin.RouterGroup) {
	auth := controllers.NewAuthController()

	api.POST("/auth/login", auth.Login)
	api.POST("/auth/logout", auth.Logout)

	authProtected := protected.Group("/auth")
	authProtected.GET("/me", auth.Me)
	authProtected.POST("/password", auth.ChangePassword)
}
