package routes

import (
	"mcloud/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 注册全部路由：CORS → 受保护域统一挂载 JWT（收口，各域子组继承）→ 各域按域文件注册
func SetupRoutes(r *gin.Engine) {
	r.Use(middleware.CORSMiddleware())

	api := r.Group("/api")

	protected := api.Group("")
	protected.Use(middleware.JWTAuth())

	registerAuth(api, protected)
	registerHosts(protected)
	registerDataExchange(protected)
	registerCloudResources(protected)
	registerStats(protected)
	registerSubnets(protected)
	registerPersons(protected)
	registerPublicIPs(protected)
	registerZeroTrusts(protected)
	registerPortMappings(protected)

	registerWS(api)
}
