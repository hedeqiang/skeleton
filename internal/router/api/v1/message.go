package v1

import (
	"github.com/gin-gonic/gin"
	handlers "github.com/hedeqiang/skeleton/internal/handler/v1"
)

// RegisterMessageRoutes 注册消息队列相关路由
func RegisterMessageRoutes(group *gin.RouterGroup, helloHandler *handlers.HelloHandler) {
	message := group.Group("/messages")
	{
		// Hello 消息发布
		message.POST("/hello/publish", helloHandler.PublishHelloMessage)

		// 未来可以添加其他消息类型的路由
		// message.POST("/notification/publish", notificationHandler.PublishNotification)
		// message.POST("/email/publish", emailHandler.PublishEmail)
		// message.GET("/status/:id", messageHandler.GetMessageStatus)
	}

	// 保持旧的路由兼容性
	group.POST("/hello/publish", helloHandler.PublishHelloMessage)
}
