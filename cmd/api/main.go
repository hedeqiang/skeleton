package main

import (
	"github.com/hedeqiang/skeleton/internal/wire"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	// 使用 Wire 创建应用实例
	application, err := wire.InitializeApplication()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// 确保在程序退出时清理资源
	defer func() {
		if err := application.Stop(); err != nil {
			application.Logger.Error("Error during application shutdown", zap.Error(err))
		}
	}()

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动应用
	go func() {
		if err := application.Run(); err != nil {
			application.Logger.Error("Application failed to run", zap.Error(err))
		}
	}()

	// 等待信号
	sig := <-sigChan
	application.Logger.Info("Received signal, shutting down...", zap.String("signal", sig.String()))
}
