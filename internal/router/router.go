package router

import (
	v1 "github.com/hedeqiang/skeleton/internal/handler/v1"
	"github.com/hedeqiang/skeleton/internal/middleware"
	"github.com/hedeqiang/skeleton/internal/router/api"
	"github.com/hedeqiang/skeleton/internal/router/system"
	"github.com/hedeqiang/skeleton/pkg/i18n"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handlers 包含所有处理器的结构体
type Handlers struct {
	UserHandler      *v1.UserHandler
	HelloHandler     *v1.HelloHandler
	SchedulerHandler *v1.SchedulerHandler
}

// SetupRouter 设置路由
// 路由层只负责路由配置，不负责依赖创建
func SetupRouter(logger *zap.Logger, i18n *i18n.I18n, handlers *Handlers) *gin.Engine {
	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	// 注册中间件
	setupMiddleware(r, logger, i18n)

	// 注册系统路由（健康检查等）
	system.RegisterSystemRoutes(r, logger)

	// 注册 API 路由
	api.RegisterAPIRoutes(r, &api.Handlers{
		UserHandler:      handlers.UserHandler,
		HelloHandler:     handlers.HelloHandler,
		SchedulerHandler: handlers.SchedulerHandler,
	})

	return r
}

// setupMiddleware 设置中间件
func setupMiddleware(r *gin.Engine, logger *zap.Logger, i18n *i18n.I18n) {
	r.Use(middleware.RequestID())
	r.Use(middleware.NewLogger(logger))
	r.Use(middleware.NewRecovery(logger))
	r.Use(middleware.CORS())
	r.Use(middleware.NewI18n(i18n))
}
