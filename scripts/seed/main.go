package main

import (
	"github.com/hedeqiang/skeleton/internal/config"
	"github.com/hedeqiang/skeleton/internal/model"
	"github.com/hedeqiang/skeleton/pkg/database"
	"github.com/hedeqiang/skeleton/pkg/logger"
	"fmt"
	"log"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("Starting database seeding...")

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

	// 5. 创建种子数据
	zapLogger.Info("Creating seed data...")

	if err := seedUsers(mainDB, zapLogger); err != nil {
		zapLogger.Fatal("Failed to seed users", zap.Error(err))
	}

	zapLogger.Info("Database seeding completed successfully!")
	fmt.Println("Database seeding completed successfully!")
}

// seedUsers 创建示例用户数据
func seedUsers(db *gorm.DB, logger *zap.Logger) error {
	// 检查是否已经有用户数据
	var count int64
	if err := db.Model(&model.User{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		logger.Info("Users already exist, skipping user seeding", zap.Int64("count", count))
		return nil
	}

	// 创建示例用户
	users := []model.User{
		{
			Username: "admin",
			Email:    "admin@example.com",
			Password: hashPassword("admin123"),
			Status:   1,
		},
		{
			Username: "testuser",
			Email:    "test@example.com",
			Password: hashPassword("test123"),
			Status:   1,
		},
		{
			Username: "john_doe",
			Email:    "john@example.com",
			Password: hashPassword("john123"),
			Status:   1,
		},
	}

	// 批量创建用户
	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to create user %s: %w", user.Username, err)
		}
		logger.Info("Created user", zap.String("username", user.Username), zap.String("email", user.Email))
	}

	logger.Info("Successfully created sample users", zap.Int("count", len(users)))
	return nil
}

// hashPassword 加密密码
func hashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	return string(hashedPassword)
}
