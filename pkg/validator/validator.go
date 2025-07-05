package validator

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

// CustomValidator 定义了我们的自定义校验器
type CustomValidator struct {
	Validate *validator.Validate
	Trans    ut.Translator
}

// Initialize a singleton instance of the validator and translator
var defaultValidator *CustomValidator

func init() {
	// 初始化翻译器
	en := en.New()
	zh := zh.New()
	uni := ut.New(en, zh)

	// 获取翻译实例 (默认使用中文)
	trans, _ := uni.GetTranslator("zh")

	// 初始化校验器
	validate := validator.New()

	// 注册一个函数，用于自定义校验
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// 注册翻译器
	_ = zh_translations.RegisterDefaultTranslations(validate, trans)
	_ = en_translations.RegisterDefaultTranslations(validate, trans)

	defaultValidator = &CustomValidator{
		Validate: validate,
		Trans:    trans,
	}

	// 将我们的自定义校验器注册为 Gin 的默认校验器
	binding.Validator = defaultValidator
}

// ValidateStruct 实现了 binding.StructValidator 接口
func (v *CustomValidator) ValidateStruct(obj interface{}) error {
	return v.Validate.Struct(obj)
}

// Engine 实现了 binding.StructValidator 接口
func (v *CustomValidator) Engine() interface{} {
	return v.Validate
}

// Translate 将校验错误翻译成更友好的格式
func Translate(err error) map[string]string {
	if defaultValidator == nil {
		return map[string]string{"error": "validator not initialized"}
	}
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		return map[string]string{"error": "not a validation error"}
	}
	return removeTopStruct(errs.Translate(defaultValidator.Trans))
}

// removeTopStruct 去除翻译后错误信息中的顶级结构体名称
func removeTopStruct(fields map[string]string) map[string]string {
	res := map[string]string{}
	for field, err := range fields {
		res[field[strings.Index(field, ".")+1:]] = err
	}
	return res
}
