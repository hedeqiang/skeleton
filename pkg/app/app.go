package app

import (
	"context"
	"fmt"
	"time"

	"github.com/hedeqiang/skeleton/pkg/config"
	"github.com/hedeqiang/skeleton/pkg/errors"
	"go.uber.org/zap"
)

// App 应用程序接口
type App interface {
	Start() error
	Stop() error
	Run() error
	GetLogger() *zap.Logger
	GetConfig() config.Config
}

// BaseApp 基础应用程序结构
type BaseApp struct {
	name    string
	version string
	logger  *zap.Logger
	config  config.Config
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewBaseApp 创建基础应用程序
func NewBaseApp(name, version string, logger *zap.Logger, config config.Config) *BaseApp {
	ctx, cancel := context.WithCancel(context.Background())
	return &BaseApp{
		name:    name,
		version: version,
		logger:  logger,
		config:  config,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start 启动应用程序
func (app *BaseApp) Start() error {
	app.logger.Info("Starting application",
		zap.String("name", app.name),
		zap.String("version", app.version),
	)
	return nil
}

// Stop 停止应用程序
func (app *BaseApp) Stop() error {
	app.logger.Info("Stopping application",
		zap.String("name", app.name),
		zap.String("version", app.version),
	)
	app.cancel()
	return nil
}

// Run 运行应用程序
func (app *BaseApp) Run() error {
	if err := app.Start(); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}
	
	// 等待上下文取消
	<-app.ctx.Done()
	
	if err := app.Stop(); err != nil {
		return fmt.Errorf("failed to stop application: %w", err)
	}
	
	return nil
}

// GetLogger 获取日志器
func (app *BaseApp) GetLogger() *zap.Logger {
	return app.logger
}

// GetConfig 获取配置
func (app *BaseApp) GetConfig() config.Config {
	return app.config
}

// Context 获取上下文
func (app *BaseApp) Context() context.Context {
	return app.ctx
}

// GracefulShutdown 优雅关闭
func (app *BaseApp) GracefulShutdown(timeout time.Duration) error {
	app.logger.Info("Starting graceful shutdown")
	
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	done := make(chan struct{})
	go func() {
		app.Stop()
		close(done)
	}()
	
	select {
	case <-done:
		app.logger.Info("Graceful shutdown completed")
		return nil
	case <-shutdownCtx.Done():
		return errors.New(errors.ErrorTypeInternal, "graceful shutdown timeout")
	}
}

// ValidateConfig 验证配置
func (app *BaseApp) ValidateConfig(validator config.Validator) error {
	if validator == nil {
		return nil
	}
	
	if validationErrors := validator.Validate(); len(validationErrors) > 0 {
		return errors.New(errors.ErrorTypeValidation, "config validation failed: "+validationErrors[0].Error())
	}
	
	return nil
}

// Health 健康检查
func (app *BaseApp) Health() map[string]interface{} {
	return map[string]interface{}{
		"name":     app.name,
		"version":  app.version,
		"status":   "healthy",
		"uptime":   time.Now().Format(time.RFC3339),
	}
}