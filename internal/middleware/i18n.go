package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/hedeqiang/skeleton/pkg/i18n"
)

func NewI18n(i18nInstance *i18n.I18n) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Accept-Language header 获取语言偏好
		acceptLang := c.GetHeader("Accept-Language")
		lang := i18n.ParseAcceptLanguage(acceptLang)

		// 也可以从查询参数中获取语言设置 (可选)
		if queryLang := c.Query("lang"); queryLang != "" {
			if queryLang == "zh" || queryLang == "en" {
				lang = queryLang
			}
		}

		// 将语言信息设置到上下文中
		ctx := i18n.SetLanguageToContext(c.Request.Context(), lang)
		c.Request = c.Request.WithContext(ctx)

		// 设置响应头信息，让客户端知道当前使用的语言
		c.Header("Content-Language", lang)

		c.Next()
	}
}
