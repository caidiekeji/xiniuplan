package database

import (
	"fmt"
	"strings"

	"cloudpivot/internal/dto"

	"gorm.io/gorm"
)

// 权限白名单
var allowedPrivileges = map[string]bool{
	"ALL":    true,
	"SELECT": true,
	"INSERT": true,
	"UPDATE": true,
	"DELETE": true,
}

// UserService 用户权限管理服务
type UserService struct {
	mysqlConnService
}

// NewUserService 创建用户权限管理服务
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		mysqlConnService: mysqlConnService{db: db},
	}
}

// ListUsers 列出 MySQL 用户
func (s *UserService) ListUsers(instanceID uint) ([]dto.UserInfo, error) {
	conn, _, err := s.connectInstance(instanceID, "mysql")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.Query("SELECT user, host FROM mysql.user ORDER BY user, host")
	if err != nil {
		return nil, fmt.Errorf("查询用户列表失败: %w", err)
	}
	defer rows.Close()

	users := make([]dto.UserInfo, 0)
	for rows.Next() {
		var u dto.UserInfo
		if err := rows.Scan(&u.User, &u.Host); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// CreateUser 创建 MySQL 用户（参数化）
func (s *UserService) CreateUser(instanceID uint, username, password, host string) error {
	if host == "" {
		host = "%"
	}

	conn, _, err := s.connectInstance(instanceID, "mysql")
	if err != nil {
		return err
	}
	defer conn.Close()

	// 使用参数化方式创建用户
	query := "CREATE USER ?@? IDENTIFIED BY ?"
	if _, err := conn.Exec(query, username, host, password); err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}
	return nil
}

// DeleteUser 删除 MySQL 用户
func (s *UserService) DeleteUser(instanceID uint, username, host string) error {
	if err := validateMySQLIdentifier(username); err != nil {
		return err
	}
	if host == "" {
		host = "%"
	}
	if err := validateMySQLHost(host); err != nil {
		return err
	}

	conn, _, err := s.connectInstance(instanceID, "mysql")
	if err != nil {
		return err
	}
	defer conn.Close()

	query := fmt.Sprintf("DROP USER '%s'@'%s'", username, host)
	if _, err := conn.Exec(query); err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}
	return nil
}

// GrantPrivileges 授权
func (s *UserService) GrantPrivileges(instanceID uint, username, host, dbName string, privileges []string) error {
	if err := validateMySQLIdentifier(username); err != nil {
		return err
	}
	if err := validateMySQLIdentifier(dbName); err != nil {
		return err
	}
	if host == "" {
		host = "%"
	}
	if err := validateMySQLHost(host); err != nil {
		return err
	}

	// 白名单校验
	privList := make([]string, 0, len(privileges))
	for _, p := range privileges {
		upper := strings.ToUpper(strings.TrimSpace(p))
		if !allowedPrivileges[upper] {
			return fmt.Errorf("不支持的权限: %s（仅支持 ALL/SELECT/INSERT/UPDATE/DELETE）", p)
		}
		privList = append(privList, upper)
	}

	conn, _, err := s.connectInstance(instanceID, "mysql")
	if err != nil {
		return err
	}
	defer conn.Close()

	privStr := strings.Join(privList, ", ")
	query := fmt.Sprintf("GRANT %s ON `%s`.* TO '%s'@'%s'", privStr, dbName, username, host)
	if _, err := conn.Exec(query); err != nil {
		return fmt.Errorf("授权失败: %w", err)
	}

	// 刷新权限
	if _, err := conn.Exec("FLUSH PRIVILEGES"); err != nil {
		return fmt.Errorf("刷新权限失败: %w", err)
	}
	return nil
}

// RevokePrivileges 撤销指定库的所有权限
func (s *UserService) RevokePrivileges(instanceID uint, username, host, dbName string) error {
	if err := validateMySQLIdentifier(username); err != nil {
		return err
	}
	if err := validateMySQLIdentifier(dbName); err != nil {
		return err
	}
	if host == "" {
		host = "%"
	}
	if err := validateMySQLHost(host); err != nil {
		return err
	}

	conn, _, err := s.connectInstance(instanceID, "mysql")
	if err != nil {
		return err
	}
	defer conn.Close()

	query := fmt.Sprintf("REVOKE ALL PRIVILEGES ON `%s`.* FROM '%s'@'%s'", dbName, username, host)
	if _, err := conn.Exec(query); err != nil {
		return fmt.Errorf("撤销权限失败: %w", err)
	}

	if _, err := conn.Exec("FLUSH PRIVILEGES"); err != nil {
		return fmt.Errorf("刷新权限失败: %w", err)
	}
	return nil
}
