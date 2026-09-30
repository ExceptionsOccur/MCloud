package routes

import (
	"mcloud/controllers"
	"mcloud/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.Use(middleware.CORSMiddleware())

	api := r.Group("/api")

	// Auth - public
	auth := controllers.NewAuthController()
	api.POST("/auth/login", auth.Login)
	api.POST("/auth/logout", auth.Logout)

	// Auth - protected
	authProtected := api.Group("/auth")
	authProtected.Use(middleware.JWTAuth())
	{
		authProtected.GET("/me", auth.Me)
		authProtected.POST("/password", auth.ChangePassword)
	}

	// Hosts - protected
	host := controllers.NewHostController()
	hosts := api.Group("/hosts")
	hosts.Use(middleware.JWTAuth())
	{
		hosts.GET("", host.Filter)
		hosts.GET("/regions", host.ListRegions)
		hosts.GET("/:id", host.Get)
		hosts.POST("", host.Create)
		hosts.PUT("/:id", host.Update)
		hosts.DELETE("/:id", host.Delete)
	}

	// Batch - protected
	batch := controllers.NewBatchController()
	batchGroup := api.Group("/batch")
	batchGroup.Use(middleware.JWTAuth())
	{
		batchGroup.POST("/hosts", batch.BatchCreate)
		batchGroup.POST("/hosts/text", batch.BatchCreateText)
		batchGroup.PUT("/hosts", batch.BatchUpdate)
	}

	// CSV - protected
	csv := controllers.NewCSVController()
	csvGroup := api.Group("")
	csvGroup.Use(middleware.JWTAuth())
	{
		csvGroup.POST("/import", csv.Import)
		csvGroup.GET("/export", csv.Export)
		csvGroup.GET("/template", csv.Template)
	}

	// Cloud Resources - protected
	cloudResource := controllers.NewCloudResourceController()
	crGroup := api.Group("/cloud-resources")
	crGroup.Use(middleware.JWTAuth())
	{
		crGroup.GET("", cloudResource.List)
		crGroup.PUT("", cloudResource.Update)
	}

	// Stats - protected
	stats := controllers.NewStatsController()
	statsGroup := api.Group("/stats")
	statsGroup.Use(middleware.JWTAuth())
	{
		statsGroup.GET("/ip-usage", stats.IPUsage)
		statsGroup.POST("/probe", stats.Probe)
		statsGroup.GET("/business", stats.BusinessStats)
	}

	// IP Subnets - protected
	subnet := controllers.NewSubnetController()
	subnetGroup := api.Group("/ip-subnets")
	subnetGroup.Use(middleware.JWTAuth())
	{
		subnetGroup.GET("", subnet.List)
		subnetGroup.POST("", subnet.Create)
		subnetGroup.PUT("/:id", subnet.Update)
		subnetGroup.DELETE("/:id", subnet.Delete)
	}

	// Persons - protected
	person := controllers.NewPersonController()
	personGroup := api.Group("/persons")
	personGroup.Use(middleware.JWTAuth())
	{
		personGroup.GET("", person.List)
		personGroup.POST("", person.Create)
		personGroup.PUT("/:id", person.Update)
		personGroup.DELETE("/:id", person.Delete)
	}

	// WebSocket - token via query param
	api.GET("/ws/probe", func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			controllers.Error(c, 40101, "未提供认证信息")
			return
		}
		controllers.HandleProbeWS(c)
	})
}
