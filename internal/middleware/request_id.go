package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDHeader is the default header name for request id.
const RequestIDHeader = "X-Request-ID"

// RequestID is a middleware that injects a request id into the context of each request.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从 header 中获取 request id
		requestID := c.Request.Header.Get(RequestIDHeader)

		// 如果 header 中没有，则生成一个新的
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 设置到 gin.Context 中，方便后续 handlers 使用
		c.Set("RequestID", requestID)

		// 设置到 response header 中，方便前端或调用方追踪
		c.Header(RequestIDHeader, requestID)

		// 继续处理请求
		c.Next()
	}
}
