package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hedeqiang/skeleton/internal/wire"
	"go.uber.org/zap"
)

func main() {
	// 使用 Wire 创建应用实例
	application, err := wire.InitializeApplication()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// 创建一个 channel 来接收系统信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 在一个 goroutine 中启动应用
	go func() {
		if err := application.Run(); err != nil {
			application.Logger().Error("Application failed to run", zap.Error(err))
		}
	}()

	// 阻塞，直到接收到退出信号
	sig := <-quit
	application.Logger().Info("Received signal, shutting down...", zap.String("signal", sig.String()))

	// 创建一个带超时的 context 用于优雅关闭
	// 给予 10 秒钟处理现有请求
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 调用 Stop 方法来关闭应用
	if err := application.Stop(ctx); err != nil {
		application.Logger().Error("Error during application shutdown", zap.Error(err))
	}

	application.Logger().Info("Application shut down gracefully")
}
