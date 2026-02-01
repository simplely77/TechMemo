package database

import (
	"fmt"
	"log"
	"techmemo/backend/config"
	"techmemo/backend/query"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Q *query.Query

func InitDB() error {
	cfg := config.AppConfig.Database

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	Q = query.Use(db)

	log.Println("Database connected successfully")
	return nil
}
