package errors

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// 错误类型定义
type ErrorType string

const (
	ErrorTypeValidation   ErrorType = "validation"
	ErrorTypeNotFound     ErrorType = "not_found"
	ErrorTypeUnauthorized ErrorType = "unauthorized"
	ErrorTypeForbidden    ErrorType = "forbidden"
	ErrorTypeConflict     ErrorType = "conflict"
	ErrorTypeInternal     ErrorType = "internal"
	ErrorTypeDatabase     ErrorType = "database"
	ErrorTypeExternal     ErrorType = "external"
)

// AppError 应用错误结构
type AppError struct {
	Type      ErrorType              `json:"type"`
	Message   string                 `json:"message"`
	Code      int                    `json:"code"`
	Err       error                  `json:"-"`
	Details   string                 `json:"details,omitempty"`
	MessageID string                 `json:"-"` // i18n 消息键
	Data      map[string]interface{} `json:"-"` // 模板数据
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

// NewI18n 创建新的支持国际化的应用错误
func NewI18n(errorType ErrorType, messageID string, data map[string]interface{}) *AppError {
	code := getStatusCodeByType(errorType)
	return &AppError{
		Type:      errorType,
		Message:   messageID, // 作为默认消息
		Code:      code,
		MessageID: messageID,
		Data:      data,
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

// WrapI18n 包装现有错误，支持国际化
func WrapI18n(err error, errorType ErrorType, messageID string, data map[string]interface{}) *AppError {
	code := getStatusCodeByType(errorType)
	return &AppError{
		Type:      errorType,
		Message:   messageID,
		Code:      code,
		Err:       err,
		MessageID: messageID,
		Data:      data,
	}
}

// WithDetails 添加错误详情
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// LocalizedMessage 获取本地化错误消息
func (e *AppError) LocalizedMessage(ctx context.Context, i18n interface {
	T(context.Context, string, map[string]interface{}) string
}) string {
	if e.MessageID != "" && i18n != nil {
		return i18n.T(ctx, e.MessageID, e.Data)
	}
	return e.Message
}

// WithData 设置模板数据
func (e *AppError) WithData(data map[string]interface{}) *AppError {
	e.Data = data
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

// 预定义错误 (支持 i18n)
var (
	ErrUserNotFound    = NewI18n(ErrorTypeNotFound, "errors.user_not_found", nil)
	ErrUserExists      = NewI18n(ErrorTypeConflict, "errors.user_exists", nil)
	ErrInvalidPassword = NewI18n(ErrorTypeUnauthorized, "errors.invalid_password", nil)
	ErrAccountDisabled = NewI18n(ErrorTypeForbidden, "errors.account_disabled", nil)
	ErrInvalidToken    = NewI18n(ErrorTypeUnauthorized, "errors.invalid_token", nil)
	ErrTokenExpired    = NewI18n(ErrorTypeUnauthorized, "errors.token_expired", nil)
	ErrInvalidInput    = NewI18n(ErrorTypeValidation, "errors.invalid_input", nil)
	ErrDatabaseError   = NewI18n(ErrorTypeDatabase, "errors.database_error", nil)
	ErrExternalService = NewI18n(ErrorTypeExternal, "errors.external_service", nil)
	ErrInternalError   = NewI18n(ErrorTypeInternal, "errors.internal_error", nil)
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
