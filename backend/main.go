package main

import (
	"log"
	"techmemo/backend/bootstrap"
	"techmemo/backend/config"
	"techmemo/backend/database"
	"techmemo/backend/router"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title TechMemo API
// @version 1.0
// @description TechMemo 后端服务
// @host localhost:8080
// @BasePath /api/v1
func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	app := bootstrap.InitApp()

	r := router.SetupRouter(app)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	port := config.AppConfig.Server.Port
	log.Printf("Server started at http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
