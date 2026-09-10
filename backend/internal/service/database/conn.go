package database

import (
	"database/sql"
	"fmt"
	"time"

	"cloudpivot/internal/model"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// mysqlConnService 提供 MySQL 连接的基础服务
type mysqlConnService struct {
	db *gorm.DB
}

// getInstance 获取实例并校验类型
func (s *mysqlConnService) getInstance(instanceID uint) (*model.DatabaseInstance, error) {
	var instance model.DatabaseInstance
	if err := s.db.First(&instance, instanceID).Error; err != nil {
		return nil, fmt.Errorf("实例不存在: %d", instanceID)
	}
	if instance.Type != "mysql" {
		return nil, fmt.Errorf("实例类型 %s 不支持此操作", instance.Type)
	}
	return &instance, nil
}

// openMySQL 打开 MySQL 连接
func (s *mysqlConnService) openMySQL(instance *model.DatabaseInstance, dbName string) (*sql.DB, error) {
	username := instance.Username
	if username == "" {
		username = "root"
	}
	host := instance.Host
	if host == "" {
		host = "127.0.0.1"
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s",
		username, instance.RootPassword, host, instance.Port, dbName)
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开 MySQL 连接失败: %w", err)
	}
	conn.SetMaxOpenConns(5)
	conn.SetMaxIdleConns(2)
	conn.SetConnMaxLifetime(5 * time.Minute)
	return conn, nil
}

// connectInstance 连接到指定实例的指定数据库
func (s *mysqlConnService) connectInstance(instanceID uint, dbName string) (*sql.DB, *model.DatabaseInstance, error) {
	instance, err := s.getInstance(instanceID)
	if err != nil {
		return nil, nil, err
	}
	conn, err := s.openMySQL(instance, dbName)
	if err != nil {
		return nil, nil, err
	}
	return conn, instance, nil
}

// systemDatabases 需要过滤的系统数据库
var systemDatabases = map[string]bool{
	"information_schema": true,
	"mysql":              true,
	"performance_schema": true,
	"sys":                true,
}
