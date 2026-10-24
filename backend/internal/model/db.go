package model

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init 初始化数据库连接并自动迁移
func Init(dbPath string) (*gorm.DB, error) {
	// 确保目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      logger.Warn,
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// 连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 自动迁移所有模型
	if err := db.AutoMigrate(
		&User{},
		&PasswordResetToken{},
		&Role{},
		&RecoveryCode{},
		&Site{},
		&DatabaseInstance{},
		&Backup{},
		&CronTask{},
		&CronLog{},
		&AuditLog{},
		&MonitorMetric{},
		&FirewallRule{},
		&Setting{},
		&InstalledApp{},
		&AppSource{},
		&AlertRule{},
		&AlertChannel{},
		&AlertHistory{},
		&DNSAccount{},
		&PluginConfig{},
		&RemoteStorage{},
		&APIKey{},
		&SiteTamperBaseline{},
		&RecycleItem{},
		&ProbeMonitor{},
		&ProbeResult{},
		&DaemonProcess{},
		&DeployApp{},
		&DeployLog{},
		&Node{},
	); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	// 初始化默认设置
	seedSettings(db)

	// 初始化内置角色
	seedBuiltInRoles(db)

	// 初始化内置官方应用源
	seedAppSources(db)

	DB = db
	return db, nil
}

// seedSettings 初始化默认设置项（不存在则插入）
func seedSettings(db *gorm.DB) {
	for _, s := range DefaultSettings {
		var count int64
		db.Model(&Setting{}).Where("`key` = ?", s.Key).Count(&count)
		if count == 0 {
			db.Create(&s)
		}
	}
}

// seedBuiltInRoles 初始化内置角色（不存在则插入，已存在则不覆盖）
func seedBuiltInRoles(db *gorm.DB) {
	for _, role := range BuiltInRoles {
		var count int64
		db.Model(&Role{}).Where("code = ?", role.Code).Count(&count)
		if count == 0 {
			db.Create(&role)
		}
	}
}

// seedAppSources 初始化内置官方应用源（不存在则创建）
func seedAppSources(db *gorm.DB) {
	var count int64
	db.Model(&AppSource{}).Where("type = ?", "official").Count(&count)
	if count == 0 {
		db.Create(&AppSource{
			Name:    "官方源",
			Type:    "official",
			Path:    "../store/apps",
			Enabled: true,
		})
	}
}
