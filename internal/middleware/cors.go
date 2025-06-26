package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS is a middleware that handles Cross-Origin Resource Sharing.
func CORS() gin.HandlerFunc {
	// 返回一个配置好的 cors 中间件
	return cors.New(cors.Config{
		// 允许的来源，"*" 表示允许所有来源。在生产环境中，应替换为您的前端域名。
		// e.g., []string{"http://www.example.com", "https://www.example.com"}
		AllowOrigins: []string{"*"},

		// 允许的 HTTP 方法
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},

		// 允许的请求头
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", RequestIDHeader},

		// 允许前端访问的响应头
		ExposeHeaders: []string{"Content-Length", RequestIDHeader},

		// 是否允许携带 cookie
		AllowCredentials: true,

		// 预检请求 (OPTIONS) 的缓存时间
		MaxAge: 12 * time.Hour,
	})
}
