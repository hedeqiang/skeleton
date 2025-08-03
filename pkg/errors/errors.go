package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// 错误类型定义
type ErrorType string

const (
	ErrorTypeValidation    ErrorType = "validation"
	ErrorTypeNotFound      ErrorType = "not_found"
	ErrorTypeUnauthorized  ErrorType = "unauthorized"
	ErrorTypeForbidden     ErrorType = "forbidden"
	ErrorTypeConflict      ErrorType = "conflict"
	ErrorTypeInternal      ErrorType = "internal"
	ErrorTypeDatabase      ErrorType = "database"
	ErrorTypeExternal      ErrorType = "external"
)

// AppError 应用错误结构
type AppError struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	Code    int       `json:"code"`
	Err     error    `json:"-"`
	Details string    `json:"details,omitempty"`
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Err.Error())
	}
	return e.Message
}

// Unwrap 解包内部错误
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is 比较错误类型
func (e *AppError) Is(target error) bool {
	if other, ok := target.(*AppError); ok {
		return e.Type == other.Type
	}
	return false
}

// StatusCode 获取HTTP状态码
func (e *AppError) StatusCode() int {
	return e.Code
}

// New 创建新的应用错误
func New(errorType ErrorType, message string) *AppError {
	code := getStatusCodeByType(errorType)
	return &AppError{
		Type:    errorType,
		Message: message,
		Code:    code,
	}
}

// Wrap 包装现有错误
func Wrap(err error, errorType ErrorType, message string) *AppError {
	code := getStatusCodeByType(errorType)
	return &AppError{
		Type:    errorType,
		Message: message,
		Code:    code,
		Err:     err,
	}
}

// WithDetails 添加错误详情
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// getStatusCodeByType 根据错误类型获取HTTP状态码
func getStatusCodeByType(errorType ErrorType) int {
	switch errorType {
	case ErrorTypeValidation:
		return http.StatusBadRequest
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeUnauthorized:
		return http.StatusUnauthorized
	case ErrorTypeForbidden:
		return http.StatusForbidden
	case ErrorTypeConflict:
		return http.StatusConflict
	case ErrorTypeInternal, ErrorTypeDatabase, ErrorTypeExternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// 预定义错误
var (
	ErrUserNotFound     = New(ErrorTypeNotFound, "用户不存在")
	ErrUserExists       = New(ErrorTypeConflict, "用户已存在")
	ErrInvalidPassword  = New(ErrorTypeUnauthorized, "密码错误")
	ErrAccountDisabled  = New(ErrorTypeForbidden, "账户已禁用")
	ErrInvalidToken     = New(ErrorTypeUnauthorized, "无效的令牌")
	ErrTokenExpired     = New(ErrorTypeUnauthorized, "令牌已过期")
	ErrInvalidInput     = New(ErrorTypeValidation, "输入参数无效")
	ErrDatabaseError    = New(ErrorTypeDatabase, "数据库错误")
	ErrExternalService  = New(ErrorTypeExternal, "外部服务错误")
	ErrInternalError    = New(ErrorTypeInternal, "内部服务器错误")
)

// 便利函数
func ValidationError(message string) *AppError {
	return New(ErrorTypeValidation, message)
}

func NotFoundError(message string) *AppError {
	return New(ErrorTypeNotFound, message)
}

func UnauthorizedError(message string) *AppError {
	return New(ErrorTypeUnauthorized, message)
}

func ForbiddenError(message string) *AppError {
	return New(ErrorTypeForbidden, message)
}

func ConflictError(message string) *AppError {
	return New(ErrorTypeConflict, message)
}

func InternalError(message string) *AppError {
	return New(ErrorTypeInternal, message)
}

// 检查错误类型
func IsNotFoundError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeNotFound
}

func IsConflictError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeConflict
}

func IsValidationError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeValidation
}

func IsUnauthorizedError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeUnauthorized
}

func IsForbiddenError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeForbidden
}

func IsDatabaseError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeDatabase
}

func IsExternalError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeExternal
}

func IsInternalError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeInternal
}

// GetHTTPStatus 获取错误对应的HTTP状态码
func GetHTTPStatus(errorType ErrorType) int {
	return getStatusCodeByType(errorType)
}