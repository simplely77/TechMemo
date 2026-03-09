package main

import (
	"context"
	"log"
	aiclient "techmemo/backend/ai/client"
	"techmemo/backend/ai/worker"
	"techmemo/backend/bootstrap"
	"techmemo/backend/config"
	"techmemo/backend/database"
	"techmemo/backend/router"
)

// @title TechMemo API
// @version 1.0
// @description TechMemo 后端服务
// @host localhost:8080
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	app := bootstrap.InitApp()

	app.AIService.SetQueue(worker.NewMemoryQueue(100))

	aiClient := aiclient.NewOpenAIClientFromConfig(config.AppConfig)

	handler := worker.NewHandler(app.AIService, app.AIDao, app.NoteDao, aiClient)

	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	// 启动worker
	go func() {
		if err := worker.NewWorker(app.AIService.GetQueue(), handler, 1).Start(workerCtx); err != nil {
			log.Printf("Worker启动失败: %v", err)
		}
	}()
	r := router.SetupRouter(app)

	port := config.AppConfig.Server.Port
	log.Printf("Server started at http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
