package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"cloudpivot/internal/model"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// ErrUnsupportedType 实例类型不支持当前操作。
//
// 这类错误是「用户对实例类型用错了功能入口」（例如对 Redis 实例调用
// MySQL 用户管理接口），属于客户端语义错误，不是服务端故障。
// 早期一律包装成 CodeServerError(500)，前端拿到 500 会显示成系统异常弹窗，
// 掩盖了真实原因；这里用哨兵错误让 API 层能翻译成 400 + 明确提示。
var ErrUnsupportedType = fmt.Errorf("实例类型不支持此操作")

// ErrInvalidRequest 请求参数不合法（非法标识符、越权权限、密码为空等）。
//
// 与 ErrUnsupportedType 同理：这是用户输入问题，应返回 400 而不是 500。
var ErrInvalidRequest = fmt.Errorf("请求参数不合法")

// NewUnsupportedTypeError 构造带具体实例类型的提示错误
func NewUnsupportedTypeError(instanceType, operation string) error {
	return fmt.Errorf("%w: 实例类型 %s，%s", ErrUnsupportedType, instanceType, operation)
}

// NewInvalidRequestError 构造参数校验失败错误
func NewInvalidRequestError(format string, args ...interface{}) error {
	return fmt.Errorf("%w: %s", ErrInvalidRequest, fmt.Sprintf(format, args...))
}

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
		return nil, NewUnsupportedTypeError(instance.Type, "该操作仅支持 MySQL 实例")
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

// noBackslashEscapes 检测服务器是否开启 NO_BACKSLASH_ESCAPES。
//
// 该模式下反斜杠不是转义符，字符串字面量只能用 ” 翻倍来转义单引号；
// 若按默认模式把 \ 转义成 \\，密码会被写成两个真实反斜杠导致无法登录。
// 检测失败时按默认模式处理（MySQL 8 默认不开该模式）。
func noBackslashEscapes(conn *sql.DB) bool {
	var mode string
	if err := conn.QueryRow("SELECT @@SESSION.sql_mode").Scan(&mode); err != nil {
		return false
	}
	return strings.Contains(strings.ToUpper(mode), "NO_BACKSLASH_ESCAPES")
}

// quoteString 按服务器 sql_mode 选择合适的字面量引用方式
func quoteString(conn *sql.DB, s string) string {
	if noBackslashEscapes(conn) {
		return quoteMySQLStringNoBackslash(s)
	}
	return quoteMySQLString(s)
}

// systemDatabases 需要过滤的系统数据库
var systemDatabases = map[string]bool{
	"information_schema": true,
	"mysql":              true,
	"performance_schema": true,
	"sys":                true,
}
