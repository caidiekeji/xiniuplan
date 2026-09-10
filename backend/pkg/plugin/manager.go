package plugin

import (
	"fmt"
	"log/slog"
	"sync"
)

// PluginManager 插件管理器，负责插件的注册、生命周期管理与扩展点调度。
// 全局单例通过 DefaultManager() 获取。
type PluginManager struct {
	mu         sync.RWMutex
	plugins    map[string]Plugin
	enabled    map[string]bool
	running    map[string]bool
	extensions map[ExtensionPoint][]interface{}
	pluginDir  string
}

var (
	defaultManager     *PluginManager
	defaultManagerOnce sync.Once
)

// DefaultManager 返回全局默认的插件管理器单例。
func DefaultManager() *PluginManager {
	defaultManagerOnce.Do(func() {
		defaultManager = &PluginManager{
			plugins:    make(map[string]Plugin),
			enabled:    make(map[string]bool),
			running:    make(map[string]bool),
			extensions: make(map[ExtensionPoint][]interface{}),
		}
	})
	return defaultManager
}

// NewManager 创建一个新的插件管理器实例（通常用于测试）。
func NewManager() *PluginManager {
	return &PluginManager{
		plugins:    make(map[string]Plugin),
		enabled:    make(map[string]bool),
		running:    make(map[string]bool),
		extensions: make(map[ExtensionPoint][]interface{}),
	}
}

// Register 注册一个插件实例。若同名插件已存在则返回错误。
func (m *PluginManager) Register(name string, p Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[name]; exists {
		return fmt.Errorf("plugin %q already registered", name)
	}
	m.plugins[name] = p
	// 默认禁用，由配置或 API 决定是否启用
	m.enabled[name] = false
	slog.Info("[plugin] registered", "name", name, "version", p.Info().Version)
	return nil
}

// LoadAll 扫描插件目录并加载插件。
// 当前实现优先使用内置注册表（通过 init() 注册的插件），
// 同时保留对 Go plugin 包动态加载 .so 的扩展能力。
func (m *PluginManager) LoadAll(pluginDir string) error {
	m.mu.Lock()
	m.pluginDir = pluginDir
	m.mu.Unlock()

	slog.Info("[plugin] loading plugins from registry", "dir", pluginDir, "count", len(m.plugins))

	// 内置注册表中的插件已通过 Register 加载，此处可扩展为
	// 扫描 pluginDir 下的 .so 文件并用 plugin.Open 动态加载。
	return nil
}

// Get 根据名称获取已注册的插件实例。
func (m *PluginManager) Get(name string) Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.plugins[name]
}

// List 返回所有已注册插件的元信息列表。
func (m *PluginManager) List() []PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]PluginInfo, 0, len(m.plugins))
	for _, p := range m.plugins {
		infos = append(infos, p.Info())
	}
	return infos
}

// IsEnabled 查询插件是否启用。
func (m *PluginManager) IsEnabled(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled[name]
}

// IsRunning 查询插件是否正在运行。
func (m *PluginManager) IsRunning(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running[name]
}

// SetEnabled 设置插件启用状态（不自动启动/停止，由调用方决定）。
func (m *PluginManager) SetEnabled(name string, enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.plugins[name]; exists {
		m.enabled[name] = enabled
	}
}

// StartPlugin 初始化并启动指定插件。
func (m *PluginManager) StartPlugin(name string, config map[string]interface{}) error {
	m.mu.Lock()
	p, exists := m.plugins[name]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("plugin %q not found", name)
	}
	if m.running[name] {
		m.mu.Unlock()
		return fmt.Errorf("plugin %q already running", name)
	}
	m.running[name] = true
	m.enabled[name] = true
	m.mu.Unlock()

	if config == nil {
		config = make(map[string]interface{})
	}

	if err := p.Init(config); err != nil {
		m.mu.Lock()
		m.running[name] = false
		m.mu.Unlock()
		return fmt.Errorf("plugin %q init failed: %w", name, err)
	}

	// Run 在独立 goroutine 中执行，避免阻塞调用方
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("[plugin] run panic", "name", name, "panic", r)
			}
			m.mu.Lock()
			m.running[name] = false
			m.mu.Unlock()
		}()
		if err := p.Run(); err != nil {
			slog.Error("[plugin] run error", "name", name, "error", err)
		}
	}()

	slog.Info("[plugin] started", "name", name)
	return nil
}

// StopPlugin 停止指定插件。
func (m *PluginManager) StopPlugin(name string) error {
	m.mu.Lock()
	p, exists := m.plugins[name]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("plugin %q not found", name)
	}
	if !m.running[name] {
		m.mu.Unlock()
		return nil
	}
	m.running[name] = false
	m.enabled[name] = false
	m.mu.Unlock()

	if err := p.Stop(); err != nil {
		return fmt.Errorf("plugin %q stop failed: %w", name, err)
	}
	slog.Info("[plugin] stopped", "name", name)
	return nil
}

// StartAll 启动所有已启用的插件。
func (m *PluginManager) StartAll() {
	m.mu.RLock()
	names := make([]string, 0, len(m.plugins))
	for name := range m.plugins {
		if m.enabled[name] && !m.running[name] {
			names = append(names, name)
		}
	}
	m.mu.RUnlock()

	for _, name := range names {
		if err := m.StartPlugin(name, nil); err != nil {
			slog.Error("[plugin] failed to start", "name", name, "error", err)
		}
	}
}

// StopAll 停止所有正在运行的插件。
func (m *PluginManager) StopAll() {
	m.mu.RLock()
	names := make([]string, 0, len(m.plugins))
	for name := range m.plugins {
		if m.running[name] {
			names = append(names, name)
		}
	}
	m.mu.RUnlock()

	for _, name := range names {
		if err := m.StopPlugin(name); err != nil {
			slog.Error("[plugin] failed to stop", "name", name, "error", err)
		}
	}
}

// RegisterExtension 向指定扩展点注册数据。
func (m *PluginManager) RegisterExtension(point ExtensionPoint, data interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.extensions[point] = append(m.extensions[point], data)
}

// GetExtensions 获取指定扩展点下注册的所有数据。
func (m *PluginManager) GetExtensions(point ExtensionPoint) []interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	items := make([]interface{}, len(m.extensions[point]))
	copy(items, m.extensions[point])
	return items
}

// PluginDir 返回插件目录路径。
func (m *PluginManager) PluginDir() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pluginDir
}
