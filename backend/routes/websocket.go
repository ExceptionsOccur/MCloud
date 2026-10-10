package routes

import (
	"mcloud/controllers"
	"mcloud/middleware"

	"github.com/gin-gonic/gin"
)

// registerWS WS 探测通道：token 走 query（浏览器 WS 无法携带 Header），
// 升级前与 HTTP 同源校验 JWT 有效性，无效/过期/伪造拒绝升级（T-022）
func registerWS(api *gin.RouterGroup) {
	api.GET("/ws/probe", func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			controllers.Error(c, 40101, "未提供认证信息")
			return
		}
		if _, err := middleware.ValidateToken(token); err != nil {
			controllers.Error(c, 40101, err.Error())
			return
		}
		controllers.HandleProbeWS(c)
	})
}
