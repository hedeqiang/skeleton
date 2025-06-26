package api

import (
	"github.com/gin-gonic/gin"
	handlers "github.com/hedeqiang/skeleton/internal/handler/v1"
	v1 "github.com/hedeqiang/skeleton/internal/router/api/v1"
)

// Handlers 包含所有处理器的结构体
type Handlers struct {
	UserHandler      *handlers.UserHandler
	HelloHandler     *handlers.HelloHandler
	SchedulerHandler *handlers.SchedulerHandler
}

// RegisterAPIRoutes 注册 API 路由
func RegisterAPIRoutes(router *gin.Engine, handlers *Handlers) {
	api := router.Group("/api")
	{
		// 注册 v1 版本的 API
		v1.RegisterV1Routes(api, &v1.Handlers{
			UserHandler:      handlers.UserHandler,
			HelloHandler:     handlers.HelloHandler,
			SchedulerHandler: handlers.SchedulerHandler,
		})

		// 未来可以在这里添加其他版本的 API
		// v2.RegisterV2Routes(api, handlers)
	}
}
