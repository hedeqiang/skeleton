package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 是返回给客户端的标准 API 格式
type Response struct {
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id"`
}

// PageResponse 分页响应结构
type PageResponse struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

const (
	// SuccessCode 表示业务处理成功
	SuccessCode = 0
	// ErrorCode 表示业务处理失败
	ErrorCode = 1
)

// Result 是一个通用的辅助函数，用于构建和发送响应
func Result(code int, msg string, data interface{}, c *gin.Context) {
	requestID, _ := c.Get("RequestID")
	c.JSON(http.StatusOK, Response{
		Code:      code,
		Msg:       msg,
		Data:      data,
		RequestID: requestID.(string),
	})
}

// ResultWithStatus 是一个通用的辅助函数，用于构建和发送带有自定义HTTP状态码的响应
func ResultWithStatus(httpStatus, code int, msg string, data interface{}, c *gin.Context) {
	requestID, _ := c.Get("RequestID")
	c.JSON(httpStatus, Response{
		Code:      code,
		Msg:       msg,
		Data:      data,
		RequestID: requestID.(string),
	})
}

// Success 发送一个成功的响应
func Success(c *gin.Context, data interface{}) {
	Result(SuccessCode, "success", data, c)
}

// SuccessWithMsg 发送一个带有自定义消息的成功响应
func SuccessWithMsg(c *gin.Context, httpStatus int, msg string, data interface{}) {
	ResultWithStatus(httpStatus, SuccessCode, msg, data, c)
}

// Error 发送一个错误响应
func Error(c *gin.Context, httpStatus int, msg string) {
	ResultWithStatus(httpStatus, ErrorCode, msg, nil, c)
}

// Fail 发送一个失败的响应
func Fail(c *gin.Context, msg string) {
	Result(ErrorCode, msg, nil, c)
}

// FailWithCode 发送一个带有自定义错误码的失败响应 (预留)
func FailWithCode(c *gin.Context, code int, msg string) {
	Result(code, msg, nil, c)
}
