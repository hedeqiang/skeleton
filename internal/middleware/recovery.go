package middleware

import (
	"github.com/hedeqiang/skeleton/pkg/response"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// NewRecovery 创建一个使用指定 logger 的恢复中间件
func NewRecovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取请求ID
				requestID, _ := c.Get("RequestID")

				// 检查连接是否断开
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				// 获取请求的 http dump
				httpRequest, _ := httputil.DumpRequest(c.Request, false)

				if brokenPipe {
					logger.Error("broken pipe",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.Any("request_id", requestID),
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				// 记录 panic 日志和堆栈跟踪
				logger.Error("recovery from panic",
					zap.Any("error", err),
					zap.String("request", string(httpRequest)),
					zap.String("stack", string(debug.Stack())),
					zap.Any("request_id", requestID),
				)

				// 返回统一的 JSON 错误响应
				response.FailWithCode(c, http.StatusInternalServerError, "Internal Server Error")
				c.Abort()
			}
		}()
		c.Next()
	}
}
