package v1

import (
	"github.com/gin-gonic/gin"
	handlers "github.com/hedeqiang/skeleton/internal/handler/v1"
)

// RegisterUserRoutes 注册用户相关路由
func RegisterUserRoutes(group *gin.RouterGroup, userHandler *handlers.UserHandler) {
	users := group.Group("/users")
	{
		users.POST("", userHandler.CreateUser)       // 创建用户
		users.GET("/:id", userHandler.GetUser)       // 获取用户信息
		users.PUT("/:id", userHandler.UpdateUser)    // 更新用户信息
		users.DELETE("/:id", userHandler.DeleteUser) // 删除用户
		users.GET("", userHandler.ListUsers)         // 获取用户列表
	}
}

// RegisterAuthRoutes 注册认证相关路由
func RegisterAuthRoutes(group *gin.RouterGroup, userHandler *handlers.UserHandler) {
	auth := group.Group("/auth")
	{
		auth.POST("/login", userHandler.Login) // 用户登录

		// 未来可以添加其他认证相关路由
		// auth.POST("/register", userHandler.Register)     // 用户注册
		// auth.POST("/logout", userHandler.Logout)         // 用户登出
		// auth.POST("/refresh", userHandler.RefreshToken)  // 刷新令牌
		// auth.GET("/profile", userHandler.GetProfile)     // 获取用户档案
	}
}
