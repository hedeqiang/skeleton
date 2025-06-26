package v1

import (
	"github.com/gin-gonic/gin"
	handlers "github.com/hedeqiang/skeleton/internal/handler/v1"
)

// Handlers 包含所有处理器的结构体
type Handlers struct {
	UserHandler      *handlers.UserHandler
	HelloHandler     *handlers.HelloHandler
	SchedulerHandler *handlers.SchedulerHandler
}

// RegisterV1Routes 注册 v1 版本的 API 路由
func RegisterV1Routes(apiGroup *gin.RouterGroup, handlers *Handlers) {
	v1Group := apiGroup.Group("/v1")
	{
		// 用户相关路由
		if handlers.UserHandler != nil {
			RegisterUserRoutes(v1Group, handlers.UserHandler)
			RegisterAuthRoutes(v1Group, handlers.UserHandler)
		}

		// 消息队列路由
		if handlers.HelloHandler != nil {
			RegisterMessageRoutes(v1Group, handlers.HelloHandler)
		}

		// 计划任务路由
		if handlers.SchedulerHandler != nil {
			RegisterSchedulerRoutes(v1Group, handlers.SchedulerHandler)
		}

		// 未来可以在这里添加其他业务模块路由
		// RegisterOrderRoutes(v1Group, handlers.OrderHandler)
		// RegisterPaymentRoutes(v1Group, handlers.PaymentHandler)
	}
}
