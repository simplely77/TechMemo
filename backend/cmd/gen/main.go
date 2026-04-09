package main

import (
	"fmt"
	"log"
	"techmemo/backend/config"

	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func main() {
	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	cfg := config.AppConfig.Database

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}

	g := gen.NewGenerator(gen.Config{
		OutPath: "query",
		Mode:    gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	g.UseDB(db)
	g.ApplyBasic(
		g.GenerateModel("search_history"),
		g.GenerateModel("note"),
		g.GenerateModel("knowledge_point"),
		g.GenerateModel("knowledge_relation"),
		g.GenerateModel("note_root_node"),
		g.GenerateModel("note_tag"),
		g.GenerateModel("note_version"),
		g.GenerateModel("tag"),
		g.GenerateModel("user"),
		g.GenerateModel("category"),
		g.GenerateModel("chat_message"),
		g.GenerateModel("chat_session"),
		g.GenerateModel("embedding"),
		g.GenerateModel("ai_process_log"),
	)
	g.Execute()
}
