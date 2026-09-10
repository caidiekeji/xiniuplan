package api

import (
	"encoding/json"
	"fmt"

	"cloudpivot/internal/common"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	"cloudpivot/pkg/plugin"

	// 触发内置插件自动注册（init()）
	_ "cloudpivot/internal/plugin"

	"github.com/gin-gonic/gin"
)

// PluginHandler 插件管理 API Handler。
type PluginHandler struct {
	manager *plugin.PluginManager
}

// NewPluginHandler 创建插件 Handler。
func NewPluginHandler() *PluginHandler {
	return &PluginHandler{
		manager: plugin.DefaultManager(),
	}
}

// RegisterPluginRoutes 注册插件管理路由。
func RegisterPluginRoutes(r *gin.RouterGroup) {
	h := NewPluginHandler()

	r.GET("/plugins", middleware.RequirePermission(model.PermPluginView), h.ListPlugins)
	r.POST("/plugins/:name/enable", middleware.RequirePermission(model.PermPluginManage), h.EnablePlugin)
	r.POST("/plugins/:name/disable", middleware.RequirePermission(model.PermPluginManage), h.DisablePlugin)
	r.GET("/plugins/:name/config", middleware.RequirePermission(model.PermPluginView), h.GetPluginConfig)
	r.PUT("/plugins/:name/config", middleware.RequirePermission(model.PermPluginManage), h.UpdatePluginConfig)
	r.POST("/plugins/:name/invoke", middleware.RequirePermission(model.PermPluginManage), h.InvokePlugin)
}

// PluginItem 插件列表项（含运行状态）。
type PluginItem struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Icon        string `json:"icon"`
	Enabled     bool   `json:"enabled"`
	Running     bool   `json:"running"`
}

// ListPlugins GET /plugins — 已安装插件列表（含状态）。
func (h *PluginHandler) ListPlugins(c *gin.Context) {
	infos := h.manager.List()
	items := make([]PluginItem, 0, len(infos))
	for _, info := range infos {
		// 从数据库读取持久化的启用状态
		var cfg model.PluginConfig
		enabled := h.manager.IsEnabled(info.Name)
		if err := model.DB.Where("name = ?", info.Name).First(&cfg).Error; err == nil {
			enabled = cfg.Enabled
		}
		items = append(items, PluginItem{
			Name:        info.Name,
			Version:     info.Version,
			Description: info.Description,
			Author:      info.Author,
			Icon:        info.Icon,
			Enabled:     enabled,
			Running:     h.manager.IsRunning(info.Name),
		})
	}
	common.Success(c, gin.H{"items": items, "total": len(items)})
}

// EnablePlugin POST /plugins/:name/enable — 启用并启动插件。
func (h *PluginHandler) EnablePlugin(c *gin.Context) {
	name := c.Param("name")
	p := h.manager.Get(name)
	if p == nil {
		common.Fail(c, common.CodeNotFound, fmt.Sprintf("插件 %q 不存在", name))
		return
	}

	// 读取持久化配置
	var cfg model.PluginConfig
	configMap := make(map[string]interface{})
	if err := model.DB.Where("name = ?", name).First(&cfg).Error; err == nil {
		if cfg.Config != "" {
			_ = json.Unmarshal([]byte(cfg.Config), &configMap)
		}
	}

	// 启动插件
	if err := h.manager.StartPlugin(name, configMap); err != nil {
		common.Fail(c, common.CodeServerError, "启动插件失败: "+err.Error())
		return
	}

	// 持久化启用状态
	cfg.Name = name
	cfg.Enabled = true
	if cfg.Config == "" {
		cfg.Config = "{}"
	}
	if cfg.ID == 0 {
		model.DB.Create(&cfg)
	} else {
		model.DB.Save(&cfg)
	}

	common.Success(c, gin.H{"name": name, "enabled": true, "running": true})
}

// DisablePlugin POST /plugins/:name/disable — 停止并禁用插件。
func (h *PluginHandler) DisablePlugin(c *gin.Context) {
	name := c.Param("name")
	p := h.manager.Get(name)
	if p == nil {
		common.Fail(c, common.CodeNotFound, fmt.Sprintf("插件 %q 不存在", name))
		return
	}

	if err := h.manager.StopPlugin(name); err != nil {
		common.Fail(c, common.CodeServerError, "停止插件失败: "+err.Error())
		return
	}

	// 持久化禁用状态
	var cfg model.PluginConfig
	if err := model.DB.Where("name = ?", name).First(&cfg).Error; err == nil {
		cfg.Enabled = false
		model.DB.Save(&cfg)
	}

	common.Success(c, gin.H{"name": name, "enabled": false, "running": false})
}

// GetPluginConfig GET /plugins/:name/config — 获取插件配置。
func (h *PluginHandler) GetPluginConfig(c *gin.Context) {
	name := c.Param("name")
	p := h.manager.Get(name)
	if p == nil {
		common.Fail(c, common.CodeNotFound, fmt.Sprintf("插件 %q 不存在", name))
		return
	}

	var cfg model.PluginConfig
	configJSON := "{}"
	enabled := false
	if err := model.DB.Where("name = ?", name).First(&cfg).Error; err == nil {
		if cfg.Config != "" {
			configJSON = cfg.Config
		}
		enabled = cfg.Enabled
	}

	// 解析为 map 返回
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &configMap); err != nil {
		configMap = make(map[string]interface{})
	}

	common.Success(c, gin.H{
		"name":    name,
		"enabled": enabled,
		"config":  configMap,
		"raw":     configJSON,
	})
}

// UpdatePluginConfig PUT /plugins/:name/config — 更新插件配置。
func (h *PluginHandler) UpdatePluginConfig(c *gin.Context) {
	name := c.Param("name")
	p := h.manager.Get(name)
	if p == nil {
		common.Fail(c, common.CodeNotFound, fmt.Sprintf("插件 %q 不存在", name))
		return
	}

	var req struct {
		Config map[string]interface{} `json:"config"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "请求体格式错误: "+err.Error())
		return
	}

	configBytes, err := json.Marshal(req.Config)
	if err != nil {
		common.Fail(c, common.CodeParamError, "配置序列化失败: "+err.Error())
		return
	}

	var cfg model.PluginConfig
	if err := model.DB.Where("name = ?", name).First(&cfg).Error; err != nil {
		// 不存在则创建
		cfg = model.PluginConfig{
			Name:    name,
			Enabled: false,
			Config:  string(configBytes),
		}
		if err := model.DB.Create(&cfg).Error; err != nil {
			common.Fail(c, common.CodeServerError, "保存配置失败: "+err.Error())
			return
		}
	} else {
		cfg.Config = string(configBytes)
		if err := model.DB.Save(&cfg).Error; err != nil {
			common.Fail(c, common.CodeServerError, "更新配置失败: "+err.Error())
			return
		}
	}

	// 若插件正在运行，热更新配置（重新 Init）
	if h.manager.IsRunning(name) {
		if err := p.Init(req.Config); err != nil {
			common.Fail(c, common.CodeServerError, "热更新配置失败: "+err.Error())
			return
		}
	}

	common.Success(c, gin.H{"name": name, "config": req.Config})
}

// InvokePlugin POST /plugins/:name/invoke — 调用插件方法（预留扩展）。
func (h *PluginHandler) InvokePlugin(c *gin.Context) {
	name := c.Param("name")
	p := h.manager.Get(name)
	if p == nil {
		common.Fail(c, common.CodeNotFound, fmt.Sprintf("插件 %q 不存在", name))
		return
	}

	var req struct {
		Method string                 `json:"method"`
		Params map[string]interface{} `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "请求体格式错误: "+err.Error())
		return
	}

	// 预留：通过扩展点机制调用插件注册的方法
	extensions := h.manager.GetExtensions(plugin.ExtAPI)
	_ = extensions

	common.Success(c, gin.H{
		"name":   name,
		"method": req.Method,
		"result": "invoke endpoint reserved for future extension",
	})
}
