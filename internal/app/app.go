package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hedeqiang/skeleton/internal/router"
	"github.com/hedeqiang/skeleton/internal/scheduler"
	"github.com/hedeqiang/skeleton/pkg/i18n"
	"github.com/hedeqiang/skeleton/pkg/idgen"

	"github.com/hedeqiang/skeleton/internal/config"
	v1 "github.com/hedeqiang/skeleton/internal/handler/v1"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Application 接口定义了应用的核��方法
type Application interface {
	Run() error
	Stop(ctx context.Context) error
	Logger() *zap.Logger
}

// App 应用程序结构体，直接包含所有依赖
type App struct {
	// HTTP 服务
	Engine *gin.Engine
	Server *http.Server

	// 基础设施依赖
	logger      *zap.Logger
	Config      *config.Config
	DataSources map[string]*gorm.DB
	MainDB      *gorm.DB
	Redis       *redis.Client
	RabbitMQ    *amqp.Connection
	IDGenerator idgen.IDGenerator
	I18n        *i18n.I18n

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
	idGenerator idgen.IDGenerator,
	i18n *i18n.I18n,
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
	engine := router.SetupRouter(logger, i18n, handlers)
	logger.Info("Router initialized successfully")

	// 初始化 HTTP Server
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.App.Host, config.App.Port),
		Handler: engine,
	}

	app := &App{
		Engine:           engine,
		Server:           server,
		logger:           logger,
		Config:           config,
		DataSources:      dataSources,
		MainDB:           mainDB,
		Redis:            redis,
		RabbitMQ:         rabbitMQ,
		IDGenerator:      idGenerator,
		I18n:             i18n,
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

// Run 启动应用程序，此方法会阻塞直到服务器关闭
func (app *App) Run() error {
	// 可选启动调度器 (如果在配置中启用)
	if app.Config.Scheduler.Enabled {
		go func() {
			app.logger.Info("Starting job registry...")
			if err := app.JobRegistry.Start(); err != nil {
				app.logger.Error("Failed to start job registry", zap.Error(err))
			}
		}()
	}

	// 启动 HTTP 服务器
	app.logger.Info("Starting HTTP server",
		zap.String("addr", app.Server.Addr),
	)
	// ListenAndServe 是一个阻塞操作，只有在服务器关闭时才会返回
	if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Stop 优雅地停止应用程序
func (app *App) Stop(ctx context.Context) error {
	app.logger.Info("Shutting down server...")

	// 优雅关闭 HTTP 服务器
	if err := app.Server.Shutdown(ctx); err != nil {
		app.logger.Error("Server forced to shutdown", zap.Error(err))
	}

	// 停止调度器
	if app.Config.Scheduler.Enabled && app.JobRegistry != nil {
		if err := app.JobRegistry.Stop(); err != nil {
			app.logger.Error("Failed to stop job registry", zap.Error(err))
		} else {
			app.logger.Info("Job registry stopped")
		}
	}

	// 关闭数据库连接
	for name, db := range app.DataSources {
		if sqlDB, err := db.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				app.logger.Error("Failed to close database connection",
					zap.String("name", name),
					zap.Error(err),
				)
			} else {
				app.logger.Info("Database connection closed", zap.String("name", name))
			}
		}
	}

	// 关闭 Redis 连接
	if app.Redis != nil {
		if err := app.Redis.Close(); err != nil {
			app.logger.Error("Failed to close Redis connection", zap.Error(err))
		} else {
			app.logger.Info("Redis connection closed")
		}
	}

	// 关闭 RabbitMQ 连接
	if app.RabbitMQ != nil && !app.RabbitMQ.IsClosed() {
		if err := app.RabbitMQ.Close(); err != nil {
			app.logger.Error("Failed to close RabbitMQ connection", zap.Error(err))
		} else {
			app.logger.Info("RabbitMQ connection closed")
		}
	}

	// 同步日志
	app.logger.Sync()

	app.logger.Info("Server exited")
	return nil
}

// Logger 返回应用的 logger 实例
func (app *App) Logger() *zap.Logger {
	return app.logger
}
