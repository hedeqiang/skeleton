package system

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RegisterSystemRoutes 注册系统路由
func RegisterSystemRoutes(router *gin.Engine, logger *zap.Logger) {
	// 健康检查路由
	RegisterHealthRoutes(router, logger)

	// 可以在这里添加其他系统路由
	// RegisterMetricsRoutes(router, logger)
	// RegisterDebugRoutes(router, logger)
}

// RegisterHealthRoutes 注册健康检查路由
func RegisterHealthRoutes(router *gin.Engine, logger *zap.Logger) {
	health := router.Group("/")
	{
		// 健康检查端点
		health.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "healthy",
				"service": "skeleton",
				"version": "1.0.0",
			})
		})

		// 就绪检查端点
		health.GET("/ready", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "ready",
			})
		})

		// 存活检查端点
		health.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})
	}

	logger.Info("Health check routes registered")
}
