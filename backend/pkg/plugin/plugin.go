// Package plugin 提供插件系统的核心接口与扩展点定义。
package plugin

// Plugin 所有插件必须实现的统一生命周期接口。
type Plugin interface {
	// Info 返回插件元信息（名称、版本、描述等）。
	Info() PluginInfo
	// Init 在插件注册后、Run 之前调用，用于读取配置、初始化资源。
	Init(config map[string]interface{}) error
	// Run 启动插件的核心逻辑（可阻塞，由管理器在 goroutine 中调用）。
	Run() error
	// Stop 停止插件并释放资源。
	Stop() error
}

// PluginInfo 插件元信息。
type PluginInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Icon        string `json:"icon"`
}

// ExtensionPoint 扩展点标识，插件可向指定扩展点注册数据，
// 宿主系统在对应位置读取并消费这些数据。
type ExtensionPoint string

const (
	// ExtMenu 菜单扩展点：插件可注册额外的菜单项。
	ExtMenu ExtensionPoint = "menu"
	// ExtAPI API 扩展点：插件可注册额外的 HTTP 路由或接口描述。
	ExtAPI ExtensionPoint = "api"
	// ExtDashboard 仪表盘扩展点：插件可注册额外的仪表盘卡片或指标。
	ExtDashboard ExtensionPoint = "dashboard"
	// ExtNotify 通知扩展点：插件可注册统一的消息发送通道。
	ExtNotify ExtensionPoint = "notify"
)
