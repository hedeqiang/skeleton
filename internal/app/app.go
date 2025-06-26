package app

import (
	"github.com/hedeqiang/skeleton/internal/router"
	"github.com/hedeqiang/skeleton/internal/scheduler"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hedeqiang/skeleton/internal/config"
	v1 "github.com/hedeqiang/skeleton/internal/handler/v1"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// App 应用程序结构体，直接包含所有依赖
type App struct {
	// HTTP 服务
	Engine *gin.Engine
	Server *http.Server

	// 基础设施依赖
	Logger      *zap.Logger
	Config      *config.Config
	DataSources map[string]*gorm.DB
	MainDB      *gorm.DB
	Redis       *redis.Client
	RabbitMQ    *amqp.Connection

	// 业务层依赖
	UserHandler      *v1.UserHandler
	HelloHandler     *v1.HelloHandler
	SchedulerHandler *v1.SchedulerHandler
	JobRegistry      *scheduler.JobRegistry
}

// NewApp 创建新的应用实例
func NewApp(
	logger *zap.Logger,
	config *config.Config,
	dataSources map[string]*gorm.DB,
	mainDB *gorm.DB,
	redis *redis.Client,
	rabbitMQ *amqp.Connection,
	userHandler *v1.UserHandler,
	helloHandler *v1.HelloHandler,
	schedulerHandler *v1.SchedulerHandler,
	jobRegistry *scheduler.JobRegistry,
) *App {
	// 创建处理器集合
	handlers := &router.Handlers{
		UserHandler:      userHandler,
		HelloHandler:     helloHandler,
		SchedulerHandler: schedulerHandler,
	}

	// 初始化路由
	engine := router.SetupRouter(logger, handlers)
	logger.Info("Router initialized successfully")

	// 初始化 HTTP Server
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.App.Host, config.App.Port),
		Handler: engine,
	}

	app := &App{
		Engine:           engine,
		Server:           server,
		Logger:           logger,
		Config:           config,
		DataSources:      dataSources,
		MainDB:           mainDB,
		Redis:            redis,
		RabbitMQ:         rabbitMQ,
		UserHandler:      userHandler,
		HelloHandler:     helloHandler,
		SchedulerHandler: schedulerHandler,
		JobRegistry:      jobRegistry,
	}

	logger.Info("Application initialized successfully",
		zap.String("host", config.App.Host),
		zap.Int("port", config.App.Port),
		zap.String("env", config.App.Env),
	)

	return app
}

// Run 启动应用程序
func (app *App) Run() error {
	// 可选启动调度器 (如果在配置中启用)
	if app.Config.Scheduler.Enabled {
		go func() {
			app.Logger.Info("Starting job registry...")
			if err := app.JobRegistry.Start(); err != nil {
				app.Logger.Error("Failed to start job registry", zap.Error(err))
			}
		}()
	}

	// 启动 HTTP 服务器
	go func() {
		app.Logger.Info("Starting HTTP server",
			zap.String("addr", app.Server.Addr),
		)
		if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.Logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.Logger.Info("Shutting down server...")

	// 停止调度器
	if app.JobRegistry != nil {
		if err := app.JobRegistry.Stop(); err != nil {
			app.Logger.Error("Failed to stop job registry", zap.Error(err))
		}
	}

	// 优雅关闭，等待现有连接完成
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Server.Shutdown(ctx); err != nil {
		app.Logger.Error("Server forced to shutdown", zap.Error(err))
		return err
	}

	app.Logger.Info("Server exited")
	return nil
}

// Stop 停止应用程序
func (app *App) Stop() error {
	// 停止调度器
	if app.JobRegistry != nil {
		if err := app.JobRegistry.Stop(); err != nil {
			app.Logger.Error("Failed to stop job registry", zap.Error(err))
		}
	}

	// 关闭数据库连接
	for name, db := range app.DataSources {
		if sqlDB, err := db.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				app.Logger.Error("Failed to close database connection",
					zap.String("name", name),
					zap.Error(err),
				)
			} else {
				app.Logger.Info("Database connection closed", zap.String("name", name))
			}
		}
	}

	// 关闭 Redis 连接
	if err := app.Redis.Close(); err != nil {
		app.Logger.Error("Failed to close Redis connection", zap.Error(err))
	} else {
		app.Logger.Info("Redis connection closed")
	}

	// 关闭 RabbitMQ 连接
	if err := app.RabbitMQ.Close(); err != nil {
		app.Logger.Error("Failed to close RabbitMQ connection", zap.Error(err))
	} else {
		app.Logger.Info("RabbitMQ connection closed")
	}

	// 同步日志
	app.Logger.Sync()

	return nil
}
