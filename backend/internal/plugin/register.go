// Package plugin 内置插件自动注册。
// 通过 init() 在包加载时自动将示例插件注册到全局插件管理器，
// 无需修改 main.go。
package plugin

import (
	"log/slog"

	"cloudpivot/pkg/plugin"
	"cloudpivot/plugins/notifier"
	"cloudpivot/plugins/systeminfo"
)

func init() {
	mgr := plugin.DefaultManager()

	// 注册系统信息插件
	if err := mgr.Register("systeminfo", systeminfo.New()); err != nil {
		slog.Error("[plugin:register] failed to register systeminfo", "error", err)
	}

	// 注册通知插件
	if err := mgr.Register("notifier", notifier.New()); err != nil {
		slog.Error("[plugin:register] failed to register notifier", "error", err)
	}

	slog.Info("[plugin:register] built-in plugins registered")
}
