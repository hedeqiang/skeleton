//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/hedeqiang/skeleton/internal/app"

	"github.com/google/wire"
)

// InitializeApplication 初始化应用程序
// Wire 会自动生成这个函数的实现
func InitializeApplication() (*app.App, error) {
	wire.Build(AllSet)
	return &app.App{}, nil
}
