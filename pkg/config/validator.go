package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ValidationError 配置验证错误
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("config validation error: field '%s' %s", e.Field, e.Message)
}

// Validator 配置验证器接口
type Validator interface {
	Validate() []ValidationError
}

// DefaultValidator 默认验证器
type DefaultValidator struct{}

func (v *DefaultValidator) Validate() []ValidationError {
	return nil
}

// ConfigValidation 配置验证工具
type ConfigValidation struct {
	errors []ValidationError
}

// NewConfigValidation 创建配置验证实例
func NewConfigValidation() *ConfigValidation {
	return &ConfigValidation{}
}

// Required 检查必填字段
func (cv *ConfigValidation) Required(field string, value interface{}) *ConfigValidation {
	if isZero(value) {
		cv.errors = append(cv.errors, ValidationError{
			Field:   field,
			Message: "is required",
		})
	}
	return cv
}

// MinLength 检查最小长度
func (cv *ConfigValidation) MinLength(field string, value string, min int) *ConfigValidation {
	if len(value) < min {
		cv.errors = append(cv.errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be at least %d characters", min),
		})
	}
	return cv
}

// MaxLength 检查最大长度
func (cv *ConfigValidation) MaxLength(field string, value string, max int) *ConfigValidation {
	if len(value) > max {
		cv.errors = append(cv.errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be no more than %d characters", max),
		})
	}
	return cv
}

// Min 检查最小值
func (cv *ConfigValidation) Min(field string, value int, min int) *ConfigValidation {
	if value < min {
		cv.errors = append(cv.errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be at least %d", min),
		})
	}
	return cv
}

// Max 检查最大值
func (cv *ConfigValidation) Max(field string, value int, max int) *ConfigValidation {
	if value > max {
		cv.errors = append(cv.errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be no more than %d", max),
		})
	}
	return cv
}

// OneOf 检查值是否在允许的范围内
func (cv *ConfigValidation) OneOf(field string, value interface{}, allowed []interface{}) *ConfigValidation {
	for _, allowedValue := range allowed {
		if reflect.DeepEqual(value, allowedValue) {
			return cv
		}
	}
	cv.errors = append(cv.errors, ValidationError{
		Field:   field,
		Message: fmt.Sprintf("must be one of %v", allowed),
	})
	return cv
}

// URL 检查URL格式
func (cv *ConfigValidation) URL(field string, value string) *ConfigValidation {
	if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
		cv.errors = append(cv.errors, ValidationError{
			Field:   field,
			Message: "must be a valid URL",
		})
	}
	return cv
}

// Email 检查邮箱格式
func (cv *ConfigValidation) Email(field string, value string) *ConfigValidation {
	if !strings.Contains(value, "@") || !strings.Contains(value, ".") {
		cv.errors = append(cv.errors, ValidationError{
			Field:   field,
			Message: "must be a valid email address",
		})
	}
	return cv
}

// Port 检查端口号
func (cv *ConfigValidation) Port(field string, value int) *ConfigValidation {
	if value < 1 || value > 65535 {
		cv.errors = append(cv.errors, ValidationError{
			Field:   field,
			Message: "must be a valid port number (1-65535)",
		})
	}
	return cv
}

// Duration 检查时间格式
func (cv *ConfigValidation) Duration(field string, value string) *ConfigValidation {
	_, err := time.ParseDuration(value)
	if err != nil {
		cv.errors = append(cv.errors, ValidationError{
			Field:   field,
			Message: "must be a valid duration (e.g., 1h, 30m, 10s)",
		})
	}
	return cv
}

// Validate 验证并返回错误
func (cv *ConfigValidation) Validate() []ValidationError {
	return cv.errors
}

// AllErrors 返回所有验证错误
func (cv *ConfigValidation) AllErrors() []ValidationError {
	return cv.errors
}

// isZero 检查值是否为零值
func isZero(value interface{}) bool {
	if value == nil {
		return true
	}
	
	switch v := value.(type) {
	case string:
		return v == ""
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(value).Int() == 0
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(value).Uint() == 0
	case float32, float64:
		return reflect.ValueOf(value).Float() == 0
	case bool:
		return !v
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		return reflect.ValueOf(value).IsZero()
	}
}

// GetEnv 获取环境变量，带默认值
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt 获取整数类型环境变量
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetEnvBool 获取布尔类型环境变量
func GetEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// GetEnvDuration 获取时间类型环境变量
func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if durationValue, err := time.ParseDuration(value); err == nil {
			return durationValue
		}
	}
	return defaultValue
}