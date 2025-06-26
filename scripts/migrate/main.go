package main

import (
	"github.com/hedeqiang/skeleton/internal/config"
	"github.com/hedeqiang/skeleton/internal/model"
	"github.com/hedeqiang/skeleton/pkg/database"
	"github.com/hedeqiang/skeleton/pkg/logger"
	"fmt"
	"log"

	"go.uber.org/zap"
)

func main() {
	fmt.Println("Starting database migration...")

	// 1. 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. 初始化日志
	zapLogger, err := logger.New(&cfg.Logger)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	// 3. 初始化数据库连接
	dataSources, err := database.NewDatabases(cfg.Databases)
	if err != nil {
		zapLogger.Fatal("Failed to initialize databases", zap.Error(err))
	}

	// 4. 获取主数据库连接
	mainDB, exists := dataSources["default"]
	if !exists {
		zapLogger.Fatal("Main database connection not found")
	}

	// 5. 执行自动迁移
	zapLogger.Info("Running auto migration...")

	err = mainDB.AutoMigrate(
		&model.User{},
		// 在这里添加其他模型
	)

	if err != nil {
		zapLogger.Fatal("Failed to run auto migration", zap.Error(err))
	}

	zapLogger.Info("Database migration completed successfully!")
	fmt.Println("Database migration completed successfully!")
}
