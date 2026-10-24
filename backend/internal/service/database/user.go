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

// CreateUser 创建 MySQL 用户
//
// MySQL 的 CREATE USER 语法中 'user'@'host' 是标识符位置，JDBC 风格的 ? 占位符
// 在这里不合法（驱动会把 ? 原样发给服务端，实测报 Error 1064 语法错误，
// 导致「创建用户」按钮 100% 失败）。因此账号名走白名单校验 + 单引号引用，
// 密码走完整转义后内联（MySQL 不支持为 CREATE USER 做服务端预编译）。
func (s *UserService) CreateUser(instanceID uint, username, password, host string) error {
	if err := validateMySQLIdentifier(username); err != nil {
		return err
	}
	if host == "" {
		host = "%"
	}
	if err := validateMySQLHost(host); err != nil {
		return err
	}
	if password == "" {
		return NewInvalidRequestError("密码不能为空")
	}

	conn, _, err := s.connectInstance(instanceID, "mysql")
	if err != nil {
		return err
	}
	defer conn.Close()

	query := fmt.Sprintf("CREATE USER %s@%s IDENTIFIED BY %s",
		quoteString(conn, username), quoteString(conn, host), quoteString(conn, password))
	if _, err := conn.Exec(query); err != nil {
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

	query := fmt.Sprintf("DROP USER %s@%s", quoteString(conn, username), quoteString(conn, host))
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
			return NewInvalidRequestError("不支持的权限: %s（仅支持 ALL/SELECT/INSERT/UPDATE/DELETE）", p)
		}
		privList = append(privList, upper)
	}

	conn, _, err := s.connectInstance(instanceID, "mysql")
	if err != nil {
		return err
	}
	defer conn.Close()

	privStr := strings.Join(privList, ", ")
	// dbName 已过 validateMySQLIdentifier 白名单（仅字母数字下划线$），
	// 用户名与 host 同样过白名单；仍统一走引用函数，防止白名单将来放宽后出漏洞。
	query := fmt.Sprintf("GRANT %s ON %s.* TO %s@%s",
		privStr, quoteMySQLIdentifier(dbName),
		quoteString(conn, username), quoteString(conn, host))
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

	query := fmt.Sprintf("REVOKE ALL PRIVILEGES ON %s.* FROM %s@%s",
		quoteMySQLIdentifier(dbName), quoteString(conn, username), quoteString(conn, host))
	if _, err := conn.Exec(query); err != nil {
		return fmt.Errorf("撤销权限失败: %w", err)
	}

	if _, err := conn.Exec("FLUSH PRIVILEGES"); err != nil {
		return fmt.Errorf("刷新权限失败: %w", err)
	}
	return nil
}
