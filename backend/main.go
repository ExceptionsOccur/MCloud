package main

import (
	"log"

	"mcloud/config"
	"mcloud/database"
	"mcloud/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	config.Load()

	database.Connect()
	database.Migrate()

	r := gin.Default()

	routes.SetupRoutes(r)

	addr := ":" + config.AppConfig.ServerPort
	log.Printf("服务器启动在 http://127.0.0.1%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}
