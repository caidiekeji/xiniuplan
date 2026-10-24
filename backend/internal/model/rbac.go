package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// StringSlice 是 []string 的 JSON 持久化类型，用于 GORM 存储权限列表
type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	b, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("StringSlice.Value: %w", err)
	}
	return string(b), nil
}

func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("StringSlice.Scan: unsupported type %T", value)
	}
	if len(b) == 0 {
		*s = nil
		return nil
	}
	return json.Unmarshal(b, s)
}

// Role 角色模型
type Role struct {
	ID          uint        `gorm:"primaryKey" json:"id"`
	Name        string      `gorm:"size:64;not null" json:"name"`
	Code        string      `gorm:"size:32;uniqueIndex;not null" json:"code"`
	Description string      `gorm:"size:255" json:"description"`
	Permissions StringSlice `gorm:"type:text" json:"permissions"`
	IsSystem    bool        `gorm:"default:false" json:"is_system"`
	CreatedAt   time.Time   `json:"created_at"`
}

func (Role) TableName() string { return "roles" }

// 权限码常量 —— 格式 module:action
const (
	// 网站管理
	PermSiteView   = "site:view"
	PermSiteCreate = "site:create"
	PermSiteUpdate = "site:update"
	PermSiteDelete = "site:delete"

	// 数据库管理
	PermDatabaseView   = "database:view"
	PermDatabaseCreate = "database:create"
	PermDatabaseUpdate = "database:update"
	PermDatabaseDelete = "database:delete"

	// 文件管理
	PermFileView   = "file:view"
	PermFileCreate = "file:create"
	PermFileUpdate = "file:update"
	PermFileDelete = "file:delete"

	// 终端
	PermTerminalView   = "terminal:view"
	PermTerminalCreate = "terminal:create"

	// 容器管理
	PermContainerView   = "container:view"
	PermContainerCreate = "container:create"
	PermContainerUpdate = "container:update"
	PermContainerDelete = "container:delete"

	// 应用商店
	PermAppStoreView    = "appstore:view"
	PermAppStoreInstall = "appstore:install"

	// 防火墙
	PermFirewallView   = "firewall:view"
	PermFirewallUpdate = "firewall:update"

	// 备份
	PermBackupView   = "backup:view"
	PermBackupCreate = "backup:create"
	PermBackupDelete = "backup:delete"

	// 监控
	PermMonitorView   = "monitor:view"
	PermMonitorUpdate = "monitor:update"

	// 用户管理
	PermUserView   = "user:view"
	PermUserCreate = "user:create"
	PermUserUpdate = "user:update"
	PermUserDelete = "user:delete"

	// 审计日志
	PermAuditView = "audit:view"

	// 系统设置
	PermSettingView   = "setting:view"
	PermSettingUpdate = "setting:update"

	// 告警
	PermAlertView   = "alert:view"
	PermAlertCreate = "alert:create"
	PermAlertUpdate = "alert:update"
	PermAlertDelete = "alert:delete"

	// 计划任务
	PermCronView   = "cron:view"
	PermCronCreate = "cron:create"
	PermCronUpdate = "cron:update"
	PermCronDelete = "cron:delete"

	// 插件管理
	PermPluginView   = "plugin:view"
	PermPluginManage = "plugin:manage"

	// Redis 管理
	PermRedisView   = "redis:view"
	PermRedisManage = "redis:manage"

	// PostgreSQL 管理
	PermPostgresView   = "postgres:view"
	PermPostgresManage = "postgres:manage"

	// SSH 安全
	PermSSHView   = "ssh:view"
	PermSSHManage = "ssh:manage"
)

// AllPermissions 全部权限码列表
var AllPermissions = []string{
	PermSiteView, PermSiteCreate, PermSiteUpdate, PermSiteDelete,
	PermDatabaseView, PermDatabaseCreate, PermDatabaseUpdate, PermDatabaseDelete,
	PermFileView, PermFileCreate, PermFileUpdate, PermFileDelete,
	PermTerminalView, PermTerminalCreate,
	PermContainerView, PermContainerCreate, PermContainerUpdate, PermContainerDelete,
	PermAppStoreView, PermAppStoreInstall,
	PermFirewallView, PermFirewallUpdate,
	PermBackupView, PermBackupCreate, PermBackupDelete,
	PermMonitorView, PermMonitorUpdate,
	PermUserView, PermUserCreate, PermUserUpdate, PermUserDelete,
	PermAuditView,
	PermSettingView, PermSettingUpdate,
	PermAlertView, PermAlertCreate, PermAlertUpdate, PermAlertDelete,
	PermCronView, PermCronCreate, PermCronUpdate, PermCronDelete,
	PermPluginView, PermPluginManage,
	PermRedisView, PermRedisManage,
	PermPostgresView, PermPostgresManage,
	PermSSHView, PermSSHManage,
}

// PermissionGroup 权限树分组
type PermissionGroup struct {
	Module string           `json:"module"`
	Label  string           `json:"label"`
	Items  []PermissionItem `json:"items"`
}

// PermissionItem 权限树条目
type PermissionItem struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

// PermissionTree 返回所有权限码的树状结构
func PermissionTree() []PermissionGroup {
	return []PermissionGroup{
		{
			Module: "site", Label: "网站管理",
			Items: []PermissionItem{
				{PermSiteView, "查看网站"}, {PermSiteCreate, "创建网站"},
				{PermSiteUpdate, "编辑网站"}, {PermSiteDelete, "删除网站"},
			},
		},
		{
			Module: "database", Label: "数据库管理",
			Items: []PermissionItem{
				{PermDatabaseView, "查看数据库"}, {PermDatabaseCreate, "创建数据库"},
				{PermDatabaseUpdate, "编辑数据库"}, {PermDatabaseDelete, "删除数据库"},
			},
		},
		{
			Module: "file", Label: "文件管理",
			Items: []PermissionItem{
				{PermFileView, "查看文件"}, {PermFileCreate, "上传文件"},
				{PermFileUpdate, "编辑文件"}, {PermFileDelete, "删除文件"},
			},
		},
		{
			Module: "terminal", Label: "Web终端",
			Items: []PermissionItem{
				{PermTerminalView, "查看终端"}, {PermTerminalCreate, "打开终端"},
			},
		},
		{
			Module: "container", Label: "容器管理",
			Items: []PermissionItem{
				{PermContainerView, "查看容器"}, {PermContainerCreate, "创建容器"},
				{PermContainerUpdate, "编辑容器"}, {PermContainerDelete, "删除容器"},
			},
		},
		{
			Module: "appstore", Label: "应用商店",
			Items: []PermissionItem{
				{PermAppStoreView, "查看应用"}, {PermAppStoreInstall, "安装应用"},
			},
		},
		{
			Module: "firewall", Label: "防火墙",
			Items: []PermissionItem{
				{PermFirewallView, "查看规则"}, {PermFirewallUpdate, "修改规则"},
			},
		},
		{
			Module: "backup", Label: "备份恢复",
			Items: []PermissionItem{
				{PermBackupView, "查看备份"}, {PermBackupCreate, "创建备份"},
				{PermBackupDelete, "删除备份"},
			},
		},
		{
			Module: "monitor", Label: "监控",
			Items: []PermissionItem{{PermMonitorView, "查看监控"}, {PermMonitorUpdate, "进程管理"}},
		},
		{
			Module: "user", Label: "用户管理",
			Items: []PermissionItem{
				{PermUserView, "查看用户"}, {PermUserCreate, "创建用户"},
				{PermUserUpdate, "编辑用户"}, {PermUserDelete, "删除用户"},
			},
		},
		{
			Module: "audit", Label: "审计日志",
			Items: []PermissionItem{{PermAuditView, "查看审计"}},
		},
		{
			Module: "setting", Label: "系统设置",
			Items: []PermissionItem{
				{PermSettingView, "查看设置"}, {PermSettingUpdate, "修改设置"},
			},
		},
		{
			Module: "alert", Label: "告警管理",
			Items: []PermissionItem{
				{PermAlertView, "查看告警"}, {PermAlertCreate, "创建告警"},
				{PermAlertUpdate, "编辑告警"}, {PermAlertDelete, "删除告警"},
			},
		},
		{
			Module: "cron", Label: "计划任务",
			Items: []PermissionItem{
				{PermCronView, "查看任务"}, {PermCronCreate, "创建任务"},
				{PermCronUpdate, "编辑任务"}, {PermCronDelete, "删除任务"},
			},
		},
		{
			Module: "plugin", Label: "插件管理",
			Items: []PermissionItem{
				{PermPluginView, "查看插件"}, {PermPluginManage, "管理插件"},
			},
		},
		{
			Module: "redis", Label: "Redis 管理",
			Items: []PermissionItem{
				{PermRedisView, "查看 Redis"}, {PermRedisManage, "操作 Redis"},
			},
		},
		{
			Module: "postgres", Label: "PostgreSQL 管理",
			Items: []PermissionItem{
				{PermPostgresView, "查看 PostgreSQL"}, {PermPostgresManage, "操作 PostgreSQL"},
			},
		},
		{
			Module: "ssh", Label: "SSH 安全",
			Items: []PermissionItem{
				{PermSSHView, "查看 SSH"}, {PermSSHManage, "管理 SSH"},
			},
		},
	}
}

// adminPermissions 返回 admin 角色权限（除 user:delete / setting:update 外的全部）
func adminPermissions() StringSlice {
	perms := make([]string, 0, len(AllPermissions))
	for _, p := range AllPermissions {
		if p == PermUserDelete || p == PermSettingUpdate {
			continue
		}
		perms = append(perms, p)
	}
	return StringSlice(perms)
}

// userPermissions 返回普通用户权限（仅 view 类）
func userPermissions() StringSlice {
	viewPerms := []string{
		PermSiteView, PermDatabaseView, PermFileView, PermTerminalView,
		PermContainerView, PermAppStoreView, PermFirewallView, PermBackupView,
		PermMonitorView, PermUserView, PermAuditView, PermSettingView,
		PermAlertView, PermCronView, PermPluginView, PermRedisView,
		PermPostgresView, PermSSHView,
	}
	return StringSlice(viewPerms)
}

// BuiltInRoles 内置角色定义
var BuiltInRoles = []Role{
	{
		Name:        "超级管理员",
		Code:        "super_admin",
		Description: "拥有全部权限",
		Permissions: StringSlice{"*"},
		IsSystem:    true,
	},
	{
		Name:        "管理员",
		Code:        "admin",
		Description: "除用户删除和系统设置更新外的全部权限",
		Permissions: adminPermissions(),
		IsSystem:    true,
	},
	{
		Name:        "普通用户",
		Code:        "user",
		Description: "仅拥有查看类权限",
		Permissions: userPermissions(),
		IsSystem:    true,
	},
}
