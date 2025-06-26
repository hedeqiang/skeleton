package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hedeqiang/skeleton/internal/config"
	"github.com/hedeqiang/skeleton/internal/scheduler"
	"github.com/hedeqiang/skeleton/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志 - 转换为旧的Logger格式
	loggerConfig := &config.Logger{
		Level:      cfg.Logger.Level,
		Encoding:   "console",
		OutputPath: []string{"stdout"},
	}

	zapLogger, err := logger.New(loggerConfig)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	zapLogger.Info("Starting scheduler service...")

	// 创建调度器服务
	schedulerService, err := scheduler.NewSchedulerService(zapLogger)
	if err != nil {
		zapLogger.Fatal("Failed to create scheduler service", zap.Error(err))
	}

	// 创建任务管理器
	jobRegistry := scheduler.NewJobRegistry(schedulerService, zapLogger, cfg.Scheduler)

	// 启动任务管理器
	if err := jobRegistry.Start(); err != nil {
		zapLogger.Fatal("Failed to start job registry", zap.Error(err))
	}

	zapLogger.Info("Scheduler service started successfully")

	// 等待关闭信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	zapLogger.Info("Shutting down scheduler service...")

	// 停止任务管理器
	if err := jobRegistry.Stop(); err != nil {
		zapLogger.Error("Failed to stop job registry gracefully", zap.Error(err))
	}

	zapLogger.Info("Scheduler service stopped")
}
