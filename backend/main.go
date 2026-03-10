package main

import (
	"context"
	"fmt"
	"log"
	aiclient "techmemo/backend/ai/client"
	"techmemo/backend/ai/queue"
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

	var q queue.Queue
	redisCfg := config.AppConfig.Redis
	if redisCfg.Enabled {
		addr := fmt.Sprintf("%s:%s", redisCfg.Host, redisCfg.Port)
		q = queue.NewRedisQueue(addr, redisCfg.Password, redisCfg.DB)
		log.Println("使用 Redis 队列")
	} else {
		q = queue.NewMemoryQueue(100)
		log.Println("使用内存队列（仅开发环境）")
	}
	app.AIService.SetQueue(q)

	aiClient := aiclient.NewOpenAIClientFromConfig(config.AppConfig)
	app.AIService.SetAIClient(aiClient)

	handler := worker.NewHandler(app.AIService)

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
