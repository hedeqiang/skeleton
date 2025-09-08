package i18n

import (
	"context"
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

const (
	ContextKeyLanguage = "language"
	DefaultLanguage    = "zh"
	FallbackLanguage   = "en"
)

type I18n struct {
	bundle      *i18n.Bundle
	localizers  map[string]*i18n.Localizer
	defaultLang string
}

type Config struct {
	DefaultLanguage string   `mapstructure:"default_language"`
	SupportLangs    []string `mapstructure:"support_languages"`
	MessagesPath    string   `mapstructure:"messages_path"`
}

func New(config Config) (*I18n, error) {
	if config.DefaultLanguage == "" {
		config.DefaultLanguage = DefaultLanguage
	}
	if config.MessagesPath == "" {
		config.MessagesPath = "./locales"
	}
	if len(config.SupportLangs) == 0 {
		config.SupportLangs = []string{"zh", "en"}
	}

	bundle := i18n.NewBundle(language.Make(config.DefaultLanguage))
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	bundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)

	localizers := make(map[string]*i18n.Localizer)

	for _, lang := range config.SupportLangs {
		messageFile := fmt.Sprintf("%s/%s.yaml", config.MessagesPath, lang)
		if _, err := bundle.LoadMessageFile(messageFile); err != nil {
			return nil, fmt.Errorf("failed to load message file for %s: %w", lang, err)
		}
		localizers[lang] = i18n.NewLocalizer(bundle, lang)
	}

	return &I18n{
		bundle:      bundle,
		localizers:  localizers,
		defaultLang: config.DefaultLanguage,
	}, nil
}

func (i *I18n) GetLocalizer(lang string) *i18n.Localizer {
	if localizer, exists := i.localizers[lang]; exists {
		return localizer
	}
	return i.localizers[i.defaultLang]
}

func (i *I18n) T(ctx context.Context, messageID string, templateData map[string]interface{}) string {
	lang := GetLanguageFromContext(ctx)
	if lang == "" {
		lang = i.defaultLang
	}

	localizer := i.GetLocalizer(lang)

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	})
	if err != nil {
		fallbackLocalizer := i.localizers[FallbackLanguage]
		if fallbackLocalizer != nil {
			if fallbackMsg, fallbackErr := fallbackLocalizer.Localize(&i18n.LocalizeConfig{
				MessageID:    messageID,
				TemplateData: templateData,
			}); fallbackErr == nil {
				return fallbackMsg
			}
		}
		return messageID
	}

	return msg
}

func (i *I18n) TWithLang(lang, messageID string, templateData map[string]interface{}) string {
	localizer := i.GetLocalizer(lang)

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	})
	if err != nil {
		fallbackLocalizer := i.localizers[FallbackLanguage]
		if fallbackLocalizer != nil {
			if fallbackMsg, fallbackErr := fallbackLocalizer.Localize(&i18n.LocalizeConfig{
				MessageID:    messageID,
				TemplateData: templateData,
			}); fallbackErr == nil {
				return fallbackMsg
			}
		}
		return messageID
	}

	return msg
}

func GetLanguageFromContext(ctx context.Context) string {
	if lang, ok := ctx.Value(ContextKeyLanguage).(string); ok {
		return lang
	}
	return ""
}

func SetLanguageToContext(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, ContextKeyLanguage, lang)
}

func ParseAcceptLanguage(acceptLang string) string {
	tags, _, err := language.ParseAcceptLanguage(acceptLang)
	if err != nil || len(tags) == 0 {
		return DefaultLanguage
	}

	for _, tag := range tags {
		lang := tag.String()
		if lang == "zh" || lang == "zh-CN" || lang == "zh-Hans" {
			return "zh"
		}
		if lang == "en" || lang == "en-US" {
			return "en"
		}
	}

	return DefaultLanguage
}
