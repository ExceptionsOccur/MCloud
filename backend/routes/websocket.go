package routes

import (
	"mcloud/controllers"

	"github.com/gin-gonic/gin"
)

// registerWS WS 探测通道：token 走 query，当前仅校验非空（JWT 有效性校验任务 T-022）
func registerWS(api *gin.RouterGroup) {
	api.GET("/ws/probe", func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			controllers.Error(c, 40101, "未提供认证信息")
			return
		}
		controllers.HandleProbeWS(c)
	})
}
