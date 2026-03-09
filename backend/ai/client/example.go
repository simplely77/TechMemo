package aiclient

// 示例：如何使用AI客户端（基于现有配置系统）

/*
import (
	"context"
	"log"
	"techmemo/backend/ai_client"
	"techmemo/backend/config"
	"techmemo/backend/dao"
	"techmemo/backend/worker"
)

func main() {
	// 1. 加载应用配置
	err := config.LoadConfig()
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	// 2. 从应用配置创建AI客户端
	aiClient := aiclient.NewOpenAIClientFromConfig(config.AppConfig.AI)

	// 3. 初始化DAO层
	q := query.Q // 假设已经初始化了query
	aiDao := dao.NewAIDao(q)
	noteDao := dao.NewNoteDao(q)

	// 4. 创建worker handler
	handler := worker.NewHandler(aiDao, noteDao, aiClient)

	// 5. 创建队列和worker
	queue := worker.NewMemoryQueue(100) // 缓冲区大小100
	worker := worker.NewWorker(queue, handler, 5) // 5个worker并发处理

	// 6. 启动worker
	ctx := context.Background()
	err = worker.Start(ctx)
	if err != nil {
		log.Fatal("启动worker失败:", err)
	}

	log.Println("AI处理系统启动成功")
}

// 或者直接使用配置创建AI客户端
func createAIClientDirectly() {
	// 方式1：使用环境变量（如果配置文件中没有设置）
	apiKey := "your-api-key-here"
	model := "gpt-4-turbo-preview"
	aiClient := aiclient.NewOpenAIClient(apiKey, model)

	// 方式2：使用配置系统（推荐）
	aiClient = aiclient.NewOpenAIClientFromConfig(config.AppConfig.AI)
}
*/
